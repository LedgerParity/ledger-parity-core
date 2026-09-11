package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// HorizonIngestor reads ordinary classic payment operations only.
// It does not ingest path payments, account creation/merge or Soroban events.
type HorizonIngestor struct {
	BaseURL    string
	Network    string // Expected exact network passphrase, required
	HTTPClient *http.Client
	MaxPages   int // Per account; hitting the bound returns an error, never a complete scan
	MaxRetries int
}

func NewHorizonIngestor(baseURL string) *HorizonIngestor {
	if baseURL == "" {
		baseURL = "https://horizon-testnet.stellar.org"
	}
	return &HorizonIngestor{BaseURL: strings.TrimRight(baseURL, "/"), HTTPClient: &http.Client{Timeout: 15 * time.Second}, MaxPages: 1000, MaxRetries: 2}
}

type horizonPaymentRecord struct {
	ID              string    `json:"id"`
	PagingToken     string    `json:"paging_token"`
	Successful      *bool     `json:"transaction_successful"`
	Type            string    `json:"type"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionHash string    `json:"transaction_hash"`
	From            string    `json:"from"`
	To              string    `json:"to"`
	FromMuxed       string    `json:"from_muxed"`
	ToMuxed         string    `json:"to_muxed"`
	Amount          string    `json:"amount"`
	AssetType       string    `json:"asset_type"`
	AssetCode       string    `json:"asset_code"`
	AssetIssuer     string    `json:"asset_issuer"`
}
type horizonPaymentsResponse struct {
	Embedded *struct {
		Records []horizonPaymentRecord `json:"records"`
	} `json:"_embedded"`
}
type FetchResult struct {
	Payments []types.OnChainPayment `json:"payments"`
	Coverage types.Coverage         `json:"coverage"`
}

// FetchOnChainPayments is the compatibility wrapper. Call Fetch for coverage evidence.
func (h *HorizonIngestor) FetchOnChainPayments(ctx context.Context, accounts []string, start, end time.Time) ([]types.OnChainPayment, error) {
	r, err := h.Fetch(ctx, accounts, start, end)
	if err != nil {
		return nil, err
	}
	return r.Payments, nil
}
func (h *HorizonIngestor) get(ctx context.Context, path string, dst any) error {
	if h.MaxRetries < 0 || h.MaxRetries > 5 {
		return fmt.Errorf("max retries must be 0..5")
	}
	client := http.Client{Timeout: 15 * time.Second}
	if h.HTTPClient != nil {
		client = *h.HTTPClient
	}
	// Provider-controlled redirects cannot escape the configured endpoint.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return fmt.Errorf("redirects are unsupported") }
	for attempt := 0; attempt <= h.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(h.BaseURL, "/")+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "LedgerParity/preview")
		resp, err := client.Do(req)
		delay := time.Duration(attempt+1) * 100 * time.Millisecond
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				data, readErr := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024+1))
				resp.Body.Close()
				if readErr != nil {
					return readErr
				}
				if len(data) > 8*1024*1024 {
					return fmt.Errorf("Horizon response exceeds 8 MiB")
				}
				return json.Unmarshal(data, dst)
			}
			status := resp.StatusCode
			if v := resp.Header.Get("Retry-After"); v != "" {
				if seconds, e := strconv.Atoi(v); e == nil {
					delay = time.Duration(seconds) * time.Second
				} else if when, e := http.ParseTime(v); e == nil {
					delay = time.Until(when)
				}
				if delay < 0 {
					delay = 0
				}
				if delay > 30*time.Second {
					resp.Body.Close()
					return fmt.Errorf("Horizon requested retry beyond 30 second budget")
				}
			}
			resp.Body.Close()
			err = fmt.Errorf("Horizon HTTP %d", status)
			if status != 429 && status < 500 {
				return err
			}
		}
		if attempt == h.MaxRetries {
			return err
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("retry exhausted")
}
func (h *HorizonIngestor) Fetch(ctx context.Context, accounts []string, start, end time.Time) (FetchResult, error) {
	fail := func(err error) (FetchResult, error) { return FetchResult{}, err }
	if h.Network == "" || len(accounts) == 0 || start.IsZero() || end.IsZero() || end.Before(start) || h.MaxPages <= 0 {
		return fail(fmt.Errorf("network, accounts, ordered closed window and positive max pages required"))
	}
	base, err := url.Parse(h.BaseURL)
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return fail(fmt.Errorf("invalid Horizon URL"))
	}
	var root struct {
		Network string `json:"network_passphrase"`
		Elder   int64  `json:"history_elder_ledger"`
		Latest  int64  `json:"history_latest_ledger"`
	}
	if err := h.get(ctx, "/", &root); err != nil {
		return fail(err)
	}
	if root.Network != h.Network {
		return fail(fmt.Errorf("Horizon network passphrase mismatch"))
	}
	result := FetchResult{Payments: []types.OnChainPayment{}, Coverage: types.Coverage{Network: h.Network, Accounts: append([]string(nil), accounts...), Start: start, End: end, Source: h.BaseURL, Reason: "Provider retention bounds unavailable"}}
	if root.Elder > 0 && root.Latest >= root.Elder {
		var elder, latest struct {
			ClosedAt time.Time `json:"closed_at"`
		}
		if err := h.get(ctx, fmt.Sprintf("/ledgers/%d", root.Elder), &elder); err != nil {
			return fail(err)
		}
		if err := h.get(ctx, fmt.Sprintf("/ledgers/%d", root.Latest), &latest); err != nil {
			return fail(err)
		}
		// Strictly older boundary avoids treating a partial first timestamp as complete.
		result.Coverage.Complete = !elder.ClosedAt.IsZero() && elder.ClosedAt.Before(start) && !latest.ClosedAt.Before(end)
		result.Coverage.Reason = "Provider retention/freshness does not enclose requested window"
		if result.Coverage.Complete {
			result.Coverage.Reason = "Ordinary payments scanned within provider-declared retained history; provider continuity trusted"
		}
	}
	seen := map[string]types.OnChainPayment{}
	for _, account := range accounts {
		if account == "" {
			return fail(fmt.Errorf("empty account"))
		}
		cursor := ""
		done := false
		previousTime := time.Time{}
		for page := 0; page < h.MaxPages; page++ {
			q := url.Values{"limit": {"200"}, "order": {"desc"}, "include_failed": {"false"}}
			if cursor != "" {
				q.Set("cursor", cursor)
			}
			var response horizonPaymentsResponse
			if err := h.get(ctx, "/accounts/"+url.PathEscape(account)+"/payments?"+q.Encode(), &response); err != nil {
				return fail(fmt.Errorf("account scan failed: %w", err))
			}
			result.Coverage.Pages++
			if response.Embedded == nil || response.Embedded.Records == nil {
				return fail(fmt.Errorf("missing payments records envelope"))
			}
			records := response.Embedded.Records
			if len(records) == 0 {
				done = true
				break
			}
			for _, rec := range records {
				token, ok := new(big.Int).SetString(rec.PagingToken, 10)
				if !ok || token.Sign() <= 0 {
					return fail(fmt.Errorf("invalid paging token"))
				}
				if cursor != "" {
					prev, _ := new(big.Int).SetString(cursor, 10)
					if token.Cmp(prev) >= 0 {
						return fail(fmt.Errorf("non-advancing descending cursor"))
					}
				}
				cursor = rec.PagingToken
				if rec.CreatedAt.IsZero() || (!previousTime.IsZero() && rec.CreatedAt.After(previousTime)) {
					return fail(fmt.Errorf("invalid or unordered payment timestamp"))
				}
				previousTime = rec.CreatedAt
				if rec.CreatedAt.Before(start) {
					done = true
					break
				}
				if rec.CreatedAt.After(end) {
					continue
				}
				if rec.Type == "" {
					return fail(fmt.Errorf("missing operation type"))
				}
				if rec.Type != "payment" {
					continue
				}
				if rec.Successful == nil {
					return fail(fmt.Errorf("missing transaction success state"))
				}
				if !*rec.Successful {
					continue
				}
				code := rec.AssetCode
				if rec.AssetType == "native" {
					code = "XLM"
				}
				from, to := rec.From, rec.To
				if rec.FromMuxed != "" {
					from = rec.FromMuxed
				}
				if rec.ToMuxed != "" {
					to = rec.ToMuxed
				}
				if rec.From != account && rec.To != account && from != account && to != account {
					return fail(fmt.Errorf("payment is unrelated to requested account"))
				}
				p := types.OnChainPayment{Network: h.Network, OperationType: rec.Type, OperationID: rec.ID, TransactionHash: rec.TransactionHash, Account: from, Destination: to, Amount: rec.Amount, AssetType: rec.AssetType, AssetCode: code, AssetIssuer: rec.AssetIssuer, Timestamp: rec.CreatedAt, Successful: true}
				if err := types.ValidateOnChain(p); err != nil {
					return fail(err)
				}
				if old, ok := seen[p.OperationID]; ok {
					if !reflect.DeepEqual(old, p) {
						return fail(fmt.Errorf("conflicting duplicate operation"))
					}
					continue
				}
				seen[p.OperationID] = p
				result.Payments = append(result.Payments, p)
			}
			if done {
				break
			}
		}
		if !done {
			return fail(fmt.Errorf("max pages reached; scan incomplete"))
		}
	}
	var after struct {
		Network string `json:"network_passphrase"`
		Elder   int64  `json:"history_elder_ledger"`
	}
	if err := h.get(ctx, "/", &after); err != nil {
		return fail(err)
	}
	if after.Network != root.Network {
		return fail(fmt.Errorf("network changed during scan"))
	}
	if after.Elder != root.Elder {
		result.Coverage.Complete = false
		result.Coverage.Reason = "Retention boundary changed during scan; rerun a bounded window"
	}
	return result, nil
}
