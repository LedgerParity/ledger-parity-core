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
	DiscrepancyAmountMismatch   DiscrepancyType = "AMOUNT_MISMATCH"
	DiscrepancyDuplicateInternal DiscrepancyType = "DUPLICATE_INTERNAL"
	DiscrepancyOrphanedOnChain   DiscrepancyType = "ORPHANED_ON_CHAIN"
	DiscrepancyStatusMismatch    DiscrepancyType = "STATUS_MISMATCH"
)

// InternalPayment represents a payment record stored in a target application's database or API.
type InternalPayment struct {
	ID          string            `json:"id"`
	SourceApp   string            `json:"source_app"`
	ReferenceID string            `json:"reference_id,omitempty"`
	Sender      string            `json:"sender"`
	Recipient   string            `json:"recipient"`
	Amount      string            `json:"amount"`
	Asset       string            `json:"asset"` // e.g. "XLM", "USDC", "native"
	Timestamp   time.Time         `json:"timestamp"`
	Status      string            `json:"status"` // e.g. "completed", "pending", "success"
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// OnChainPayment represents a payment operation or transaction verified on the Stellar ledger.
type OnChainPayment struct {
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
)

// MatchResult represents the detailed reconciliation verdict for a pair or orphan.
type MatchResult struct {
	InternalPayment *InternalPayment `json:"internal_payment,omitempty"`
	OnChainPayment  *OnChainPayment  `json:"on_chain_payment,omitempty"`
	Status          MatchStatus      `json:"status"`
	Discrepancy     DiscrepancyType  `json:"discrepancy"`
	Notes           string           `json:"notes,omitempty"`
	TimeDeltaSec    int64            `json:"time_delta_sec,omitempty"`
	AmountDelta     string           `json:"amount_delta,omitempty"`
}

// DiscrepancyReport aggregates all match results and high-level metrics for a reconciliation run.
type DiscrepancyReport struct {
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

// Summary returns a concise human-readable summary string of the report.
func (r *DiscrepancyReport) Summary() string {
	return fmt.Sprintf(
		"Reconciliation Summary [%s]: Internal: %d | On-Chain: %d | Matched: %d | Discrepancies: %d",
		r.TargetApp, r.TotalInternal, r.TotalOnChain, r.TotalMatched, r.TotalDiscrepancies,
	)
}
