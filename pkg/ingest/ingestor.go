package ingest

import (
	"context"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

// Ingestor defines the interface for pulling on-chain payment operations from Stellar networks.
type Ingestor interface {
	FetchOnChainPayments(ctx context.Context, accounts []string, start, end time.Time) ([]types.OnChainPayment, error)
}

// MockIngestor provides static or programmatic payments for testing and offline scenarios.
type MockIngestor struct {
	Payments []types.OnChainPayment
}

func NewMockIngestor(payments []types.OnChainPayment) *MockIngestor {
	return &MockIngestor{Payments: payments}
}

func (m *MockIngestor) FetchOnChainPayments(ctx context.Context, accounts []string, start, end time.Time) ([]types.OnChainPayment, error) {
	accountMap := make(map[string]bool)
	for _, acc := range accounts {
		accountMap[acc] = true
	}

	var filtered []types.OnChainPayment
	for _, p := range m.Payments {
		// Filter by account if accounts list is non-empty
		if len(accountMap) > 0 && !accountMap[p.Account] && !accountMap[p.Destination] {
			continue
		}
		// Filter by time window if set
		if !start.IsZero() && p.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && p.Timestamp.After(end) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, nil
}
