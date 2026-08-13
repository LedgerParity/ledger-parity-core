package engine_test

import (
	"testing"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/engine"
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

func TestReconciler_ExactAndTolerantMatch(t *testing.T) {
	rec := engine.NewReconciler()
	now := time.Now()

	internals := []types.InternalPayment{
		{
			ID:        "pay_1",
			Recipient: "GAXYZ1234567890ACCOUNT",
			Amount:    "100.0000000",
			Asset:     "XLM",
			Timestamp: now,
		},
		{
			ID:        "pay_2",
			Recipient: "GABC9876543210ACCOUNT",
			Amount:    "50.5000000",
			Asset:     "USDC",
			Timestamp: now.Add(-2 * time.Minute),
		},
	}

	onChains := []types.OnChainPayment{
		{
			TransactionHash: "hash_1",
			OperationID:     "op_1",
			Destination:     "GAXYZ1234567890ACCOUNT",
			Amount:          "100.0000000",
			AssetCode:       "XLM",
			Timestamp:       now,
			Successful:      true,
		},
		{
			TransactionHash: "hash_2",
			OperationID:     "op_2",
			Destination:     "GABC9876543210ACCOUNT",
			Amount:          "50.5000000",
			AssetCode:       "USDC",
			Timestamp:       now,
			Successful:      true,
		},
	}

	report := rec.Reconcile("test_app", now.Add(-1*time.Hour), now.Add(1*time.Hour), internals, onChains)

	if report.TotalInternal != 2 {
		t.Fatalf("expected 2 total internal, got %d", report.TotalInternal)
	}
	if report.TotalMatched != 2 {
		t.Fatalf("expected 2 matched, got %d", report.TotalMatched)
	}
	if report.TotalDiscrepancies != 0 {
		t.Fatalf("expected 0 discrepancies, got %d", report.TotalDiscrepancies)
	}
}

func TestReconciler_MissingOnChain(t *testing.T) {
	rec := engine.NewReconciler()
	now := time.Now()

	internals := []types.InternalPayment{
		{
			ID:        "pay_missing",
			Recipient: "GNOTFOUNDACCOUNT",
			Amount:    "250.0000000",
			Asset:     "XLM",
			Timestamp: now,
		},
	}

	onChains := []types.OnChainPayment{}

	report := rec.Reconcile("test_app", now.Add(-1*time.Hour), now.Add(1*time.Hour), internals, onChains)

	if report.TotalDiscrepancies != 1 {
		t.Fatalf("expected 1 discrepancy, got %d", report.TotalDiscrepancies)
	}
	if report.Results[0].Discrepancy != types.DiscrepancyMissingOnChain {
		t.Fatalf("expected MISSING_ON_CHAIN, got %s", report.Results[0].Discrepancy)
	}
}

func TestReconciler_AmountMismatch(t *testing.T) {
	rec := engine.NewReconciler()
	now := time.Now()

	internals := []types.InternalPayment{
		{
			ID:        "pay_amt",
			Recipient: "GACC123",
			Amount:    "100.0000000",
			Asset:     "XLM",
			Timestamp: now,
		},
	}

	onChains := []types.OnChainPayment{
		{
			TransactionHash: "hash_amt",
			OperationID:     "op_amt",
			Destination:     "GACC123",
			Amount:          "80.0000000", // Mismatched amount
			AssetCode:       "XLM",
			Timestamp:       now,
			Successful:      true,
		},
	}

	report := rec.Reconcile("test_app", now.Add(-1*time.Hour), now.Add(1*time.Hour), internals, onChains)

	if report.TotalDiscrepancies != 1 {
		t.Fatalf("expected 1 discrepancy, got %d", report.TotalDiscrepancies)
	}
	if report.Results[0].Discrepancy != types.DiscrepancyAmountMismatch {
		t.Fatalf("expected AMOUNT_MISMATCH, got %s", report.Results[0].Discrepancy)
	}
}

func TestReconciler_DuplicateInternal(t *testing.T) {
	rec := engine.NewReconciler()
	now := time.Now()

	internals := []types.InternalPayment{
		{
			ID:        "pay_dup1",
			Recipient: "GACC123",
			Amount:    "100.0000000",
			Asset:     "XLM",
			Timestamp: now,
		},
		{
			ID:        "pay_dup2",
			Recipient: "GACC123",
			Amount:    "100.0000000",
			Asset:     "XLM",
			Timestamp: now,
		},
	}

	onChains := []types.OnChainPayment{
		{
			TransactionHash: "hash_single",
			OperationID:     "op_single",
			Destination:     "GACC123",
			Amount:          "100.0000000",
			AssetCode:       "XLM",
			Timestamp:       now,
			Successful:      true,
		},
	}

	report := rec.Reconcile("test_app", now.Add(-1*time.Hour), now.Add(1*time.Hour), internals, onChains)

	if report.DiscrepancyCounts[types.DiscrepancyDuplicateInternal] != 2 {
		t.Fatalf("expected 2 DUPLICATE_INTERNAL discrepancies, got %d", report.DiscrepancyCounts[types.DiscrepancyDuplicateInternal])
	}
}

func TestReconciler_OrphanedOnChain(t *testing.T) {
	rec := engine.NewReconciler()
	now := time.Now()

	internals := []types.InternalPayment{}

	onChains := []types.OnChainPayment{
		{
			TransactionHash: "hash_orphan",
			OperationID:     "op_orphan",
			Destination:     "GORPHANACCOUNT",
			Amount:          "500.0000000",
			AssetCode:       "XLM",
			Timestamp:       now,
			Successful:      true,
		},
	}

	report := rec.Reconcile("test_app", now.Add(-1*time.Hour), now.Add(1*time.Hour), internals, onChains)

	if report.TotalDiscrepancies != 1 {
		t.Fatalf("expected 1 discrepancy, got %d", report.TotalDiscrepancies)
	}
	if report.Results[0].Discrepancy != types.DiscrepancyOrphanedOnChain {
		t.Fatalf("expected ORPHANED_ON_CHAIN, got %s", report.Results[0].Discrepancy)
	}
}
