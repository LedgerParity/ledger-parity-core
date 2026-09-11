package types

import (
	"fmt"
	"time"
)

// DiscrepancyType identifies specific payment mismatch classifications.
type DiscrepancyType string

const (
	DiscrepancyNone              DiscrepancyType = "NONE"
	DiscrepancyMissingOnChain    DiscrepancyType = "MISSING_ON_CHAIN"
	DiscrepancyAmountMismatch    DiscrepancyType = "AMOUNT_MISMATCH"
	DiscrepancyDuplicateInternal DiscrepancyType = "DUPLICATE_INTERNAL"
	DiscrepancyOrphanedOnChain   DiscrepancyType = "ORPHANED_ON_CHAIN"
	DiscrepancyStatusMismatch    DiscrepancyType = "STATUS_MISMATCH"
	DiscrepancyUnresolved        DiscrepancyType = "UNRESOLVED"
	DiscrepancyInvalid           DiscrepancyType = "INVALID_DATA"
)

// InternalPayment represents a payment record stored in a target application's database or API.
type InternalPayment struct {
	SettlementStart   time.Time         `json:"settlement_start,omitempty"`
	SettlementEnd     time.Time         `json:"settlement_end,omitempty"`
	BusinessReference string            `json:"business_reference,omitempty"`
	OperationType     string            `json:"operation_type"` // Must be payment
	Network           string            `json:"network"`        // Exact Stellar network passphrase
	OperationID       string            `json:"operation_id,omitempty"`
	AssetType         string            `json:"asset_type"`
	AssetIssuer       string            `json:"asset_issuer,omitempty"`
	AssetContract     string            `json:"asset_contract,omitempty"` // Unsupported, never silently aliased
	ID                string            `json:"id"`
	SourceApp         string            `json:"source_app"`
	ReferenceID       string            `json:"reference_id,omitempty"`
	Sender            string            `json:"sender"`
	Recipient         string            `json:"recipient"`
	Amount            string            `json:"amount"`
	Asset             string            `json:"asset"` // e.g. "XLM", "USDC", "native"
	Timestamp         time.Time         `json:"timestamp"`
	Status            string            `json:"status"` // e.g. "completed", "pending", "success"
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// MatchingWindow uses an explicit asserted interval without widening it. Legacy
// timestamp expectations retain their configured symmetric time tolerance.
func (p InternalPayment) MatchingWindow(tolerance time.Duration) (time.Time, time.Time) {
	if !p.SettlementStart.IsZero() {
		return p.SettlementStart, p.SettlementEnd
	}
	return p.Timestamp.Add(-tolerance), p.Timestamp.Add(tolerance)
}

func (p InternalPayment) InWindow(start, end time.Time) bool {
	a, b := p.Timestamp, p.Timestamp
	if !p.SettlementStart.IsZero() {
		a, b = p.SettlementStart, p.SettlementEnd
	}
	return (start.IsZero() || !a.Before(start)) && (end.IsZero() || !b.After(end))
}

// OnChainPayment represents a provider observation (or an explicitly offline fixture).
type OnChainPayment struct {
	Network         string    `json:"network"`
	OperationType   string    `json:"operation_type"` // Only payment supported
	AssetType       string    `json:"asset_type"`
	AssetContract   string    `json:"asset_contract,omitempty"`
	TransactionHash string    `json:"transaction_hash"`
	OperationID     string    `json:"operation_id"`
	Account         string    `json:"account"`     // Source account
	Destination     string    `json:"destination"` // Recipient account
	Amount          string    `json:"amount"`
	AssetCode       string    `json:"asset_code"`   // "XLM" or asset code
	AssetIssuer     string    `json:"asset_issuer"` // Empty for native XLM
	Timestamp       time.Time `json:"timestamp"`
	Memo            string    `json:"memo,omitempty"`
	LedgerSequence  int64     `json:"ledger_sequence"`
	Successful      bool      `json:"successful"`
}

// MatchStatus represents the outcome of cross-referencing an internal record against on-chain data.
type MatchStatus string

const (
	MatchExact       MatchStatus = "EXACT"
	MatchTolerant    MatchStatus = "TOLERANT_MATCH"
	MatchDiscrepancy MatchStatus = "DISCREPANCY"
	MatchUnknown     MatchStatus = "UNKNOWN"
)

// MatchResult represents the detailed reconciliation verdict for a pair or orphan.
type MatchResult struct {
	CandidateOperationIDs []string         `json:"candidate_operation_ids,omitempty"`
	InternalPayment       *InternalPayment `json:"internal_payment,omitempty"`
	OnChainPayment        *OnChainPayment  `json:"on_chain_payment,omitempty"`
	Status                MatchStatus      `json:"status"`
	Discrepancy           DiscrepancyType  `json:"discrepancy"`
	Notes                 string           `json:"notes,omitempty"`
	TimeDeltaSec          int64            `json:"time_delta_sec,omitempty"`
	AmountDelta           string           `json:"amount_delta,omitempty"`
}

// DiscrepancyReport aggregates all match results and high-level metrics for a reconciliation run.
type DiscrepancyReport struct {
	Coverage           Coverage                `json:"coverage"`
	TotalUnknown       int                     `json:"total_unknown"`
	GeneratedAt        time.Time               `json:"generated_at"`
	TargetApp          string                  `json:"target_app"`
	TimeWindowStart    time.Time               `json:"time_window_start"`
	TimeWindowEnd      time.Time               `json:"time_window_end"`
	TotalInternal      int                     `json:"total_internal"`
	TotalOnChain       int                     `json:"total_on_chain"`
	TotalMatched       int                     `json:"total_matched"`
	TotalDiscrepancies int                     `json:"total_discrepancies"`
	DiscrepancyCounts  map[DiscrepancyType]int `json:"discrepancy_counts"`
	Results            []MatchResult           `json:"results"`
}

// Coverage describes a closed time window for ordinary classic payments.
// InternalComplete is an operator assertion about the application export.
type Coverage struct {
	Network          string    `json:"network"`
	Accounts         []string  `json:"accounts"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	Complete         bool      `json:"complete"`
	InternalComplete bool      `json:"internal_complete"`
	Reason           string    `json:"reason"`
	Source           string    `json:"source"`
	Pages            int       `json:"pages"`
}

// Summary returns a concise human-readable summary string of the report.
func (r *DiscrepancyReport) Summary() string {
	return fmt.Sprintf(
		"Reconciliation Summary [%s]: Internal: %d | On-Chain: %d | Matched: %d | Discrepancies: %d | Unknown: %d",
		r.TargetApp, r.TotalInternal, r.TotalOnChain, r.TotalMatched, r.TotalDiscrepancies, r.TotalUnknown,
	)
}
