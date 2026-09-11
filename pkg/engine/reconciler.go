package engine

import (
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
	"github.com/LedgerParity/ledger-parity-core/pkg/utils"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Amounts are always exact. Time tolerance applies to candidates, not money.
type ReconcileOptions struct {
	TimeframeToleranceSec int64          `json:"timeframe_tolerance_sec"`
	Coverage              types.Coverage `json:"coverage"`
}

func DefaultOptions() ReconcileOptions { return ReconcileOptions{TimeframeToleranceSec: 600} }

type Reconciler struct{ Opts ReconcileOptions }

func NewReconciler(opts ...ReconcileOptions) *Reconciler {
	o := DefaultOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	return &Reconciler{Opts: o}
}
func settled(status string) bool {
	switch strings.ToLower(status) {
	case "completed", "success", "settled":
		return true
	}
	return false
}
func observedAccount(c types.Coverage, a, b string) bool {
	for _, v := range c.Accounts {
		if v == a || v == b {
			return true
		}
	}
	return false
}
func unresolved(r *types.MatchResult, reason string) {
	r.Status = types.MatchUnknown
	r.Discrepancy = types.DiscrepancyUnresolved
	r.Notes = reason
}

// Reconcile never treats an unproven scan/export as evidence of absence.
func (r *Reconciler) Reconcile(app string, start, end time.Time, ips []types.InternalPayment, ops []types.OnChainPayment) *types.DiscrepancyReport {
	c := r.Opts.Coverage
	report := &types.DiscrepancyReport{GeneratedAt: time.Now().UTC(), TargetApp: app, TimeWindowStart: start, TimeWindowEnd: end, TotalInternal: len(ips), TotalOnChain: len(ops), Coverage: c, DiscrepancyCounts: map[types.DiscrepancyType]int{}, Results: []types.MatchResult{}}
	validWindow := !start.IsZero() && !end.Before(start) && r.Opts.TimeframeToleranceSec >= 0 && r.Opts.TimeframeToleranceSec <= 86400
	tolerance := time.Duration(r.Opts.TimeframeToleranceSec) * time.Second
	seen := map[string]int{}
	unique := []types.OnChainPayment{}
	badChain := false
	for _, p := range ops {
		if types.ValidateOnChain(p) != nil {
			badChain = true
			continue
		}
		key := p.Network + "\x00" + p.OperationID
		if idx, ok := seen[key]; ok {
			if !reflect.DeepEqual(unique[idx], p) {
				badChain = true
			}
			continue
		}
		seen[key] = len(unique)
		unique = append(unique, p)
	}
	ids := map[string]int{}
	for _, p := range ips {
		ids[p.SourceApp+"\x00"+p.ID]++
	}
	claims := make([][]int, len(unique))
	candidates := make([][]int, len(ips))
	for i := range ips {
		p := &ips[i]
		result := types.MatchResult{InternalPayment: p}
		unresolved(&result, "No unique settlement established")
		if err := types.ValidateInternal(*p); err != nil || !validWindow || !p.InWindow(start, end) {
			result.Discrepancy = types.DiscrepancyInvalid
			result.Notes = "Invalid record or reconciliation window"
		} else if badChain {
			unresolved(&result, "Malformed or conflicting on-chain observations")
		} else {
			if ids[p.SourceApp+"\x00"+p.ID] > 1 {
				result.Status = types.MatchDiscrepancy
				result.Discrepancy = types.DiscrepancyDuplicateInternal
				result.Notes = "Repeated internal ID within source"
			}
			for j, o := range unique {
				if o.Network != p.Network || o.Account != p.Sender || o.Destination != p.Recipient || o.AssetType != p.AssetType || o.AssetCode != p.Asset || o.AssetIssuer != p.AssetIssuer {
					continue
				}
				a, b := p.MatchingWindow(tolerance)
				if o.Timestamp.Before(a) || o.Timestamp.After(b) {
					continue
				}
				if p.OperationID != "" && p.OperationID != o.OperationID {
					continue
				}
				// ReferenceID means transaction hash only. Memos are not unique identifiers.
				if p.ReferenceID != "" && p.ReferenceID != o.TransactionHash {
					continue
				}
				candidates[i] = append(candidates[i], j)
				claims[j] = append(claims[j], i)
			}
		}
		for _, j := range candidates[i] {
			result.CandidateOperationIDs = append(result.CandidateOperationIDs, unique[j].OperationID)
		}
		sort.Strings(result.CandidateOperationIDs)
		report.Results = append(report.Results, result)
	}
	completeFor := func(p types.InternalPayment) bool {
		a, b := p.MatchingWindow(tolerance)
		return c.Complete && c.Network == p.Network && observedAccount(c, p.Sender, p.Recipient) && !c.Start.IsZero() && !c.End.IsZero() && !c.Start.After(a) && !c.End.Before(b)
	}
	for i, p := range ips {
		result := &report.Results[i]
		if result.Discrepancy == types.DiscrepancyInvalid || result.Discrepancy == types.DiscrepancyDuplicateInternal || badChain {
			continue
		}
		cs := candidates[i]
		switch {
		case len(cs) > 1:
			unresolved(result, "Multiple candidate operations; supply an operation ID")
		case len(cs) == 1:
			j := cs[0]
			o := &unique[j]
			if len(claims[j]) > 1 {
				unresolved(result, "Multiple internal records claim this operation")
				continue
			}
			if p.OperationID == "" && !completeFor(p) {
				unresolved(result, "Candidate observed but full matching window is unproven; supply operation ID")
				continue
			}
			result.OnChainPayment = o
			if !o.Successful {
				unresolved(result, "Failed transaction is not settlement")
				continue
			}
			delta, _ := utils.AmountDeltaScaled(p.Amount, o.Amount, 7)
			if delta != "0.0000000" && p.OperationID == "" && p.ReferenceID == "" {
				unresolved(result, "Nearby payment has a different amount; no explicit identity proves its relationship")
				continue
			}
			if !settled(p.Status) {
				result.Status = types.MatchDiscrepancy
				result.Discrepancy = types.DiscrepancyStatusMismatch
				result.Notes = "Settlement observed but internal status is not completed/success/settled"
				continue
			}
			result.AmountDelta = delta
			if delta != "0.0000000" {
				result.Status = types.MatchDiscrepancy
				result.Discrepancy = types.DiscrepancyAmountMismatch
				result.Notes = "Explicit identity matched; delivered amount differs"
				continue
			}
			result.Status = types.MatchExact
			result.Discrepancy = types.DiscrepancyNone
			result.Notes = "Unique ordinary payment; exact amount, asset and direction"
			if !p.Timestamp.IsZero() && !p.Timestamp.Equal(o.Timestamp) {
				result.Status = types.MatchTolerant
			}
			d := time.Duration(0)
			if !p.Timestamp.IsZero() {
				d = o.Timestamp.Sub(p.Timestamp)
			}
			if d < 0 {
				d = -d
			}
			result.TimeDeltaSec = int64(d / time.Second)
		default:
			if settled(p.Status) && completeFor(p) {
				result.Status = types.MatchDiscrepancy
				result.Discrepancy = types.DiscrepancyMissingOnChain
				result.Notes = "No matching ordinary payment in the covered candidate window"
			} else {
				unresolved(result, "No settlement found; status or coverage does not justify absence")
			}
		}
	}
	for j := range unique {
		o := &unique[j]
		if len(claims[j]) > 0 || !o.Successful || !validWindow || o.Timestamp.Before(start) || o.Timestamp.After(end) {
			continue
		}
		v := types.MatchResult{OnChainPayment: o}
		unresolved(&v, "Unclaimed observation; export completeness/scope unproven")
		// Invalid internal rows can hide claims, so suppress definite orphan conclusions.
		validExport := true
		for _, ip := range ips {
			if types.ValidateInternal(ip) != nil || !ip.InWindow(start, end) || ids[ip.SourceApp+"\x00"+ip.ID] > 1 {
				validExport = false
			}
		}
		if validExport && !badChain && c.InternalComplete && c.Network == o.Network && observedAccount(c, o.Account, o.Destination) && !c.Start.IsZero() && !c.Start.After(start) && !c.End.Before(end) {
			v.Status = types.MatchDiscrepancy
			v.Discrepancy = types.DiscrepancyOrphanedOnChain
			v.Notes = "No claim in operator-declared complete export"
		}
		report.Results = append(report.Results, v)
	}
	if badChain {
		report.Coverage.Complete = false
		report.Coverage.Reason = "Malformed or conflicting on-chain observations"
		if len(ips) == 0 {
			v := types.MatchResult{}
			unresolved(&v, report.Coverage.Reason)
			report.Results = append(report.Results, v)
		}
	}
	if !validWindow && len(report.Results) == 0 {
		report.Results = append(report.Results, types.MatchResult{Status: types.MatchUnknown, Discrepancy: types.DiscrepancyInvalid, Notes: "Invalid reconciliation window"})
	}
	if len(report.Results) == 0 && (!c.Complete || !c.InternalComplete) {
		v := types.MatchResult{}
		unresolved(&v, "Empty inputs with unproven scan/export completeness")
		report.Results = append(report.Results, v)
	}
	for _, v := range report.Results {
		switch v.Status {
		case types.MatchExact, types.MatchTolerant:
			report.TotalMatched++
		case types.MatchUnknown:
			report.TotalUnknown++
		default:
			report.TotalDiscrepancies++
			report.DiscrepancyCounts[v.Discrepancy]++
		}
	}
	return report
}
