package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

// ReconcileOptions sets parameters for the matching engine.
type ReconcileOptions struct {
	TimeframeToleranceSec int64             `json:"timeframe_tolerance_sec"` // Default 600 seconds (10 mins)
	AmountTolerance       float64           `json:"amount_tolerance"`       // Default 0.0001 (floating point epsilon)
	AssetAliases          map[string]string `json:"asset_aliases"`          // e.g. {"native": "XLM", "XLM": "XLM"}
	IgnoreFailedOnChain   bool              `json:"ignore_failed_on_chain"` // Exclude failed Stellar txs
}

// DefaultOptions provides sensible defaults for Stellar reconciliation.
func DefaultOptions() ReconcileOptions {
	return ReconcileOptions{
		TimeframeToleranceSec: 600,
		AmountTolerance:       0.00001,
		AssetAliases: map[string]string{
			"native": "XLM",
			"xlm":    "XLM",
			"usdc":   "USDC",
		},
		IgnoreFailedOnChain: true,
	}
}

// Reconciler executes reconciliation algorithms between internal records and on-chain payments.
type Reconciler struct {
	Opts ReconcileOptions
}

// NewReconciler creates a Reconciler with custom or default options.
func NewReconciler(opts ...ReconcileOptions) *Reconciler {
	o := DefaultOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	return &Reconciler{Opts: o}
}

// Reconcile processes internal payments against on-chain payments and returns a full DiscrepancyReport.
func (r *Reconciler) Reconcile(
	targetApp string,
	windowStart, windowEnd time.Time,
	internals []types.InternalPayment,
	onChains []types.OnChainPayment,
) *types.DiscrepancyReport {

	report := &types.DiscrepancyReport{
		GeneratedAt:       time.Now(),
		TargetApp:         targetApp,
		TimeWindowStart:   windowStart,
		TimeWindowEnd:     windowEnd,
		TotalInternal:     len(internals),
		TotalOnChain:      len(onChains),
		DiscrepancyCounts: make(map[types.DiscrepancyType]int),
	}

	// Filter on-chain payments if configured
	var activeOnChains []types.OnChainPayment
	for _, oc := range onChains {
		if r.Opts.IgnoreFailedOnChain && !oc.Successful {
			continue
		}
		activeOnChains = append(activeOnChains, oc)
	}

	matchedOnChainIndices := make(map[int]int) // onChainIdx -> count of internal records matching it
	onChainToInternalMap := make(map[int][]int)

	results := make([]types.MatchResult, 0, len(internals)+len(activeOnChains))

	// Track matches for each internal payment
	for _, ip := range internals {
		bestIdx, status, discType, notes, deltaSec, amountDelta := r.findBestOnChainMatch(ip, activeOnChains)

		res := types.MatchResult{
			InternalPayment: &ip,
			Status:          status,
			Discrepancy:     discType,
			Notes:           notes,
			TimeDeltaSec:    deltaSec,
			AmountDelta:     amountDelta,
		}

		if bestIdx >= 0 {
			res.OnChainPayment = &activeOnChains[bestIdx]
			matchedOnChainIndices[bestIdx]++
			onChainToInternalMap[bestIdx] = append(onChainToInternalMap[bestIdx], len(results))
		}

		results = append(results, res)
	}

	// Post-processing Pass 1: Identify Duplicate Internal Records
	for ocIdx, internalIndices := range onChainToInternalMap {
		if len(internalIndices) > 1 {
			for _, resIdx := range internalIndices {
				results[resIdx].Status = types.MatchDiscrepancy
				results[resIdx].Discrepancy = types.DiscrepancyDuplicateInternal
				results[resIdx].Notes = fmt.Sprintf(
					"Duplicate internal records (%d) matched to single on-chain tx %s",
					len(internalIndices), activeOnChains[ocIdx].TransactionHash,
				)
			}
		}
	}

	// Post-processing Pass 2: Identify Orphaned On-Chain Payments
	for i, oc := range activeOnChains {
		if matchedOnChainIndices[i] == 0 {
			ocCopy := oc
			results = append(results, types.MatchResult{
				OnChainPayment: &ocCopy,
				Status:         types.MatchDiscrepancy,
				Discrepancy:    types.DiscrepancyOrphanedOnChain,
				Notes:          fmt.Sprintf("Orphaned on-chain payment %s for recipient %s with no internal record", oc.TransactionHash, oc.Destination),
			})
		}
	}

	// Calculate aggregates
	matchedCount := 0
	discrepancyCount := 0

	for _, res := range results {
		if res.Status == types.MatchExact || res.Status == types.MatchTolerant {
			matchedCount++
		} else {
			discrepancyCount++
			report.DiscrepancyCounts[res.Discrepancy]++
		}
	}

	report.TotalMatched = matchedCount
	report.TotalDiscrepancies = discrepancyCount
	report.Results = results

	return report
}

func (r *Reconciler) findBestOnChainMatch(
	ip types.InternalPayment,
	onChains []types.OnChainPayment,
) (bestIdx int, status types.MatchStatus, discType types.DiscrepancyType, notes string, deltaSec int64, amountDelta string) {

	bestIdx = -1
	minTimeDelta := int64(math.MaxInt64)
	ipAsset := r.normalizeAsset(ip.Asset)
	ipAmt, _ := strconv.ParseFloat(ip.Amount, 64)

	// Step 1: Look for explicit ReferenceID / Hash match
	if ip.ReferenceID != "" {
		for i, oc := range onChains {
			if oc.TransactionHash == ip.ReferenceID || oc.Memo == ip.ReferenceID {
				timeDelta := int64(math.Abs(float64(oc.Timestamp.Sub(ip.Timestamp) / time.Second)))
				if r.amountsEqual(ip.Amount, oc.Amount) {
					return i, types.MatchExact, types.DiscrepancyNone, "Matched by explicit reference ID and amount", timeDelta, "0"
				}
				return i, types.MatchDiscrepancy, types.DiscrepancyAmountMismatch,
					fmt.Sprintf("Reference ID matched tx %s but amount internal (%s) != on-chain (%s)", oc.TransactionHash, ip.Amount, oc.Amount),
					timeDelta, r.calcAmountDelta(ip.Amount, oc.Amount)
			}
		}
	}

	// Step 2: Search by Recipient + Asset + Time Window
	for i, oc := range onChains {
		ocAsset := r.normalizeAsset(oc.AssetCode)
		if !strings.EqualFold(ipAsset, ocAsset) {
			continue
		}

		// Account/Recipient matching (case-insensitive)
		if !strings.EqualFold(ip.Recipient, oc.Destination) && !strings.EqualFold(ip.Recipient, oc.Account) {
			continue
		}

		timeDelta := int64(math.Abs(float64(oc.Timestamp.Sub(ip.Timestamp) / time.Second)))
		if timeDelta > r.Opts.TimeframeToleranceSec {
			continue
		}

		ocAmt, _ := strconv.ParseFloat(oc.Amount, 64)
		amtDiff := math.Abs(ipAmt - ocAmt)

		if amtDiff <= r.Opts.AmountTolerance {
			if timeDelta < minTimeDelta {
				minTimeDelta = timeDelta
				bestIdx = i
				if timeDelta == 0 {
					status = types.MatchExact
					notes = "Exact recipient, amount, asset, and timestamp match"
				} else {
					status = types.MatchTolerant
					notes = fmt.Sprintf("Matched within %ds tolerance window", timeDelta)
				}
				discType = types.DiscrepancyNone
				deltaSec = timeDelta
				amountDelta = "0"
			}
		} else if bestIdx == -1 {
			// Candidate found with recipient and time, but amount differs
			bestIdx = i
			status = types.MatchDiscrepancy
			discType = types.DiscrepancyAmountMismatch
			notes = fmt.Sprintf("On-chain record found for recipient %s but amount internal (%s) != on-chain (%s)", oc.Destination, ip.Amount, oc.Amount)
			deltaSec = timeDelta
			amountDelta = r.calcAmountDelta(ip.Amount, oc.Amount)
		}
	}

	if bestIdx != -1 {
		return bestIdx, status, discType, notes, deltaSec, amountDelta
	}

	// Step 3: No match found on-chain
	return -1, types.MatchDiscrepancy, types.DiscrepancyMissingOnChain,
		fmt.Sprintf("No on-chain transaction found for internal payment ID %s (Recipient: %s, Amount: %s %s)", ip.ID, ip.Recipient, ip.Amount, ip.Asset),
		0, ip.Amount
}

func (r *Reconciler) normalizeAsset(asset string) string {
	upper := strings.ToUpper(strings.TrimSpace(asset))
	if aliased, ok := r.Opts.AssetAliases[strings.ToLower(upper)]; ok {
		return aliased
	}
	return upper
}

func (r *Reconciler) amountsEqual(amt1, amt2 string) bool {
	f1, err1 := strconv.ParseFloat(amt1, 64)
	f2, err2 := strconv.ParseFloat(amt2, 64)
	if err1 != nil || err2 != nil {
		return strings.TrimSpace(amt1) == strings.TrimSpace(amt2)
	}
	return math.Abs(f1-f2) <= r.Opts.AmountTolerance
}

func (r *Reconciler) calcAmountDelta(amtInternal, amtOnChain string) string {
	f1, _ := strconv.ParseFloat(amtInternal, 64)
	f2, _ := strconv.ParseFloat(amtOnChain, 64)
	return fmt.Sprintf("%.7f", f1-f2)
}
