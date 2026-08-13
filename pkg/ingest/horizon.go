package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

// HorizonIngestor queries Horizon API server for payment operations.
type HorizonIngestor struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewHorizonIngestor creates a new Horizon API client.
func NewHorizonIngestor(baseURL string) *HorizonIngestor {
	if baseURL == "" {
		baseURL = "https://horizon-testnet.stellar.org"
	}
	return &HorizonIngestor{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type horizonPaymentRecord struct {
	ID                 string `json:"id"`
	PagingToken        string `json:"paging_token"`
	Successful         bool   `json:"transaction_successful"`
	Type               string `json:"type"`
	TypeI              int    `json:"type_i"`
	CreatedAt          string `json:"created_at"`
	TransactionHash    string `json:"transaction_hash"`
	SourceAccount      string `json:"source_account"`
	From               string `json:"from"`
	To                 string `json:"to"`
	Amount             string `json:"amount"`
	AssetType          string `json:"asset_type"`
	AssetCode          string `json:"asset_code"`
	AssetIssuer        string `json:"asset_issuer"`
	Into               string `json:"into"`
	StartingBalance    string `json:"starting_balance"`
	Funder             string `json:"funder"`
	Account            string `json:"account"`
}

type horizonPaymentsResponse struct {
	Embedded struct {
		Records []horizonPaymentRecord `json:"records"`
	} `json:"_embedded"`
	Links struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
	} `json:"_links"`
}

func (h *HorizonIngestor) FetchOnChainPayments(ctx context.Context, accounts []string, start, end time.Time) ([]types.OnChainPayment, error) {
	var allPayments []types.OnChainPayment
	seenOps := make(map[string]bool)

	for _, account := range accounts {
		payments, err := h.fetchAccountPayments(ctx, account, start, end)
		if err != nil {
			return nil, fmt.Errorf("failed fetching payments for account %s: %w", account, err)
		}
		for _, p := range payments {
			if !seenOps[p.OperationID] {
				seenOps[p.OperationID] = true
				allPayments = append(allPayments, p)
			}
		}
	}
	return allPayments, nil
}

func (h *HorizonIngestor) fetchAccountPayments(ctx context.Context, account string, start, end time.Time) ([]types.OnChainPayment, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/payments", h.BaseURL, url.PathEscape(account))
	reqURL := fmt.Sprintf("%s?limit=200&order=desc", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "LedgerParity-Recon/1.0")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("horizon HTTP error status: %d", resp.StatusCode)
	}

	var horizonResp horizonPaymentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&horizonResp); err != nil {
		return nil, fmt.Errorf("failed to decode horizon json: %w", err)
	}

	var payments []types.OnChainPayment
	for _, rec := range horizonResp.Embedded.Records {
		createdAt, err := time.Parse(time.RFC3339, rec.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}

		if !start.IsZero() && createdAt.Before(start) {
			continue
		}
		if !end.IsZero() && createdAt.After(end) {
			continue
		}

		assetCode := rec.AssetCode
		if rec.AssetType == "native" || assetCode == "" {
			assetCode = "XLM"
		}

		destination := rec.To
		if destination == "" {
			destination = rec.Into
		}
		if destination == "" {
			destination = rec.Account
		}

		sourceAcc := rec.From
		if sourceAcc == "" {
			sourceAcc = rec.SourceAccount
		}

		amount := rec.Amount
		if amount == "" {
			amount = rec.StartingBalance
		}

		payments = append(payments, types.OnChainPayment{
			TransactionHash: rec.TransactionHash,
			OperationID:     rec.ID,
			Account:         sourceAcc,
			Destination:     destination,
			Amount:          amount,
			AssetCode:       assetCode,
			AssetIssuer:     rec.AssetIssuer,
			Timestamp:       createdAt,
			Successful:      rec.Successful,
		})
	}

	return payments, nil
}
