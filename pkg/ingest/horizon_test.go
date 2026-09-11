package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var stamp = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

func record(id string) map[string]any {
	return map[string]any{"id": id, "paging_token": id, "transaction_successful": true, "type": "payment", "created_at": stamp.Format(time.RFC3339), "transaction_hash": "tx" + id, "from": "A", "to": "B", "amount": "10.0000001", "asset_type": "native"}
}
func page(w http.ResponseWriter, rows ...map[string]any) {
	if rows == nil {
		rows = []map[string]any{}
	}
	json.NewEncoder(w).Encode(map[string]any{"_embedded": map[string]any{"records": rows}})
}
func server(t *testing.T, fn http.HandlerFunc) (*HorizonIngestor, func()) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `{"network_passphrase":"test","history_elder_ledger":1,"history_latest_ledger":2}`)
		case "/ledgers/1":
			fmt.Fprint(w, `{"closed_at":"2026-08-31T00:00:00Z"}`)
		case "/ledgers/2":
			fmt.Fprint(w, `{"closed_at":"2026-09-02T00:00:00Z"}`)
		default:
			fn(w, r)
		}
	}))
	h := NewHorizonIngestor(s.URL)
	h.Network = "test"
	h.MaxRetries = 0
	return h, s.Close
}
func TestPaginationAndCrossAccountDedup(t *testing.T) {
	calls := 0
	h, close := server(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("order") != "desc" || r.URL.Query().Get("include_failed") != "false" {
			t.Error("query")
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			page(w, record("3"), record("2"))
		case "2":
			page(w, record("1"))
		default:
			page(w)
		}
	})
	defer close()
	res, err := h.Fetch(context.Background(), []string{"A", "B"}, stamp.Add(-time.Hour), stamp.Add(time.Hour))
	if err != nil || len(res.Payments) != 3 || !res.Coverage.Complete || calls != 6 {
		t.Fatalf("%+v %v calls %d", res, err, calls)
	}
}
func TestIngestionFailsClosed(t *testing.T) {
	for _, name := range []string{"later HTTP error", "malformed time", "missing success", "missing records", "cursor loop", "max pages", "wrong account", "bad amount", "conflicting duplicate"} {
		t.Run(name, func(t *testing.T) {
			h, close := server(t, func(w http.ResponseWriter, r *http.Request) {
				row := record("2")
				cursor := r.URL.Query().Get("cursor")
				switch name {
				case "later HTTP error":
					if cursor != "" {
						w.WriteHeader(503)
						return
					}
				case "malformed time":
					row["created_at"] = "garbage"
				case "missing success":
					delete(row, "transaction_successful")
				case "missing records":
					fmt.Fprint(w, `{}`)
					return
				case "wrong account":
					row["from"] = "Z"
					row["to"] = "Y"
				case "bad amount":
					row["amount"] = "NaN"
				case "conflicting duplicate":
					if strings.Contains(r.URL.Path, "/B/") {
						row["amount"] = "9"
					}
					if cursor != "" {
						page(w)
						return
					}
				}
				page(w, row)
			})
			defer close()
			if name == "max pages" {
				h.MaxPages = 1
			}
			res, err := h.Fetch(context.Background(), []string{"A", "B"}, stamp.Add(-time.Hour), stamp.Add(time.Hour))
			if err == nil || len(res.Payments) != 0 || res.Coverage.Complete {
				t.Fatalf("expected failure: %+v %v", res, err)
			}
		})
	}
}
func TestRetryNetworkAndCoverage(t *testing.T) {
	calls := 0
	h, close := server(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		page(w)
	})
	defer close()
	h.MaxRetries = 1
	res, err := h.Fetch(context.Background(), []string{"A"}, stamp.Add(-time.Hour), stamp.Add(time.Hour))
	if err != nil || !res.Coverage.Complete || calls != 2 {
		t.Fatal(res, err, calls)
	}
	res, err = h.Fetch(context.Background(), []string{"A"}, stamp.Add(-48*time.Hour), stamp.Add(time.Hour))
	if err != nil || res.Coverage.Complete {
		t.Fatal(res, err)
	}
	h.Network = "other"
	if _, err = h.Fetch(context.Background(), []string{"A"}, stamp, stamp); err == nil {
		t.Fatal("network mismatch accepted")
	}
	h.Network = "test"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = h.Fetch(ctx, []string{"A"}, stamp, stamp); err == nil {
		t.Fatal("cancel ignored")
	}
}
func TestUnsupportedAndFailedExcluded(t *testing.T) {
	h, close := server(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") != "" {
			page(w)
			return
		}
		a, b := record("2"), record("1")
		a["type"] = "path_payment_strict_receive"
		b["transaction_successful"] = false
		page(w, a, b)
	})
	defer close()
	res, err := h.Fetch(context.Background(), []string{"A"}, stamp, stamp)
	if err != nil || len(res.Payments) != 0 {
		t.Fatal(res, err)
	}
}

func TestMoreThan200Payments(t *testing.T) {
	h, close := server(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("cursor") {
		case "":
			rows := []map[string]any{}
			for i := 201; i > 1; i-- {
				rows = append(rows, record(fmt.Sprint(i)))
			}
			page(w, rows...)
		case "2":
			page(w, record("1"))
		default:
			page(w)
		}
	})
	defer close()
	res, err := h.Fetch(context.Background(), []string{"A"}, stamp, stamp)
	if err != nil || len(res.Payments) != 201 || res.Coverage.Pages != 3 {
		t.Fatal(len(res.Payments), res.Coverage, err)
	}
}
