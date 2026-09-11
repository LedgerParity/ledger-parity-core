package engine_test

import (
	"github.com/LedgerParity/ledger-parity-core/pkg/engine"
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

func pair() (types.InternalPayment, types.OnChainPayment) {
	p := types.InternalPayment{OperationType: "payment", ID: "1", SourceApp: "test", Network: "test-network", Sender: "sender", Recipient: "recipient", Asset: "XLM", AssetType: "native", Amount: "922337203685.4775807", Timestamp: now, Status: "completed", ReferenceID: "tx", OperationID: "1"}
	o := types.OnChainPayment{Network: p.Network, OperationType: "payment", OperationID: "1", TransactionHash: "tx", Account: p.Sender, Destination: p.Recipient, AssetCode: p.Asset, AssetType: p.AssetType, Amount: p.Amount, Timestamp: now, Successful: true}
	return p, o
}
func options() engine.ReconcileOptions {
	return engine.ReconcileOptions{TimeframeToleranceSec: 600, Coverage: types.Coverage{Network: "test-network", Accounts: []string{"sender"}, Start: now.Add(-2 * time.Hour), End: now.Add(2 * time.Hour), Complete: true, InternalComplete: true}}
}
func run(ips []types.InternalPayment, ops []types.OnChainPayment, opt engine.ReconcileOptions) *types.DiscrepancyReport {
	return engine.NewReconciler(opt).Reconcile("test", now.Add(-time.Hour), now.Add(time.Hour), ips, ops)
}
func TestExactAndFailureCases(t *testing.T) {
	for _, tc := range []struct {
		name   string
		edit   func(*types.InternalPayment, *types.OnChainPayment, *engine.ReconcileOptions)
		status types.MatchStatus
		disc   types.DiscrepancyType
	}{
		{"exact", func(*types.InternalPayment, *types.OnChainPayment, *engine.ReconcileOptions) {}, types.MatchExact, types.DiscrepancyNone},
		{"one stroop at maximum", func(p *types.InternalPayment, o *types.OnChainPayment, _ *engine.ReconcileOptions) {
			o.Amount = "922337203685.4775806"
		}, types.MatchDiscrepancy, types.DiscrepancyAmountMismatch},
		{"time tolerant", func(_ *types.InternalPayment, o *types.OnChainPayment, _ *engine.ReconcileOptions) {
			o.Timestamp = now.Add(time.Second)
		}, types.MatchTolerant, types.DiscrepancyNone},
		{"failed", func(_ *types.InternalPayment, o *types.OnChainPayment, _ *engine.ReconcileOptions) {
			o.Successful = false
		}, types.MatchUnknown, types.DiscrepancyUnresolved},
		{"pending with settlement", func(p *types.InternalPayment, _ *types.OnChainPayment, _ *engine.ReconcileOptions) {
			p.Status = "pending"
		}, types.MatchDiscrepancy, types.DiscrepancyStatusMismatch},
		{"invalid amount", func(p *types.InternalPayment, _ *types.OnChainPayment, _ *engine.ReconcileOptions) { p.Amount = "NaN" }, types.MatchUnknown, types.DiscrepancyInvalid},
		{"excess precision", func(p *types.InternalPayment, _ *types.OnChainPayment, _ *engine.ReconcileOptions) {
			p.Amount = "1.00000001"
		}, types.MatchUnknown, types.DiscrepancyInvalid},
		{"missing network", func(p *types.InternalPayment, _ *types.OnChainPayment, _ *engine.ReconcileOptions) { p.Network = "" }, types.MatchUnknown, types.DiscrepancyInvalid},
		{"contract asset", func(p *types.InternalPayment, _ *types.OnChainPayment, _ *engine.ReconcileOptions) {
			p.AssetContract = "Ctoken"
		}, types.MatchUnknown, types.DiscrepancyInvalid},
		{"bad chain amount", func(_ *types.InternalPayment, o *types.OnChainPayment, _ *engine.ReconcileOptions) { o.Amount = "" }, types.MatchUnknown, types.DiscrepancyUnresolved},
		{"incomplete fuzzy", func(p *types.InternalPayment, _ *types.OnChainPayment, opt *engine.ReconcileOptions) {
			p.OperationID = ""
			p.ReferenceID = ""
			opt.Coverage.Complete = false
		}, types.MatchUnknown, types.DiscrepancyUnresolved},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, o := pair()
			opt := options()
			tc.edit(&p, &o, &opt)
			r := run([]types.InternalPayment{p}, []types.OnChainPayment{o}, opt)
			v := r.Results[0]
			if v.Status != tc.status || v.Discrepancy != tc.disc {
				t.Fatalf("got %+v", v)
			}
		})
	}
}
func TestReferenceNeverBypassesIdentity(t *testing.T) {
	for _, edit := range []func(*types.OnChainPayment){func(o *types.OnChainPayment) { o.Network = "other" }, func(o *types.OnChainPayment) { o.Account = "other" }, func(o *types.OnChainPayment) { o.Destination = "Recipient" }, func(o *types.OnChainPayment) { o.Account, o.Destination = o.Destination, o.Account }, func(o *types.OnChainPayment) {
		o.AssetType = "credit_alphanum4"
		o.AssetCode = "XLM"
		o.AssetIssuer = "issuer"
	}, func(o *types.OnChainPayment) { o.TransactionHash = "other" }, func(o *types.OnChainPayment) { o.OperationID = "2" }, func(o *types.OnChainPayment) { o.Timestamp = now.Add(time.Hour) }} {
		p, o := pair()
		edit(&o)
		r := run([]types.InternalPayment{p}, []types.OnChainPayment{o}, options())
		if r.TotalMatched != 0 {
			t.Fatal("identity bypass")
		}
	}
	p, o := pair()
	p.Asset = "USDC"
	p.AssetType = "credit_alphanum4"
	p.AssetIssuer = "issuer1"
	o.AssetCode = p.Asset
	o.AssetType = p.AssetType
	o.AssetIssuer = "issuer2"
	if run([]types.InternalPayment{p}, []types.OnChainPayment{o}, options()).TotalMatched != 0 {
		t.Fatal("issuer ignored")
	}
}
func TestCoverageAbsenceAndPending(t *testing.T) {
	p, _ := pair()
	opt := options()
	if r := run([]types.InternalPayment{p}, nil, opt); r.Results[0].Discrepancy != types.DiscrepancyMissingOnChain {
		t.Fatal(r)
	}
	for _, edit := range []func(*engine.ReconcileOptions){func(o *engine.ReconcileOptions) { o.Coverage.Complete = false }, func(o *engine.ReconcileOptions) { o.Coverage.Start = now }, func(o *engine.ReconcileOptions) { o.Coverage.End = now }, func(o *engine.ReconcileOptions) { o.Coverage.Network = "other" }, func(o *engine.ReconcileOptions) { o.Coverage.Accounts = nil }} {
		o := options()
		edit(&o)
		if r := run([]types.InternalPayment{p}, nil, o); r.TotalUnknown != 1 {
			t.Fatal(r)
		}
	}
	p.Status = "pending"
	if r := run([]types.InternalPayment{p}, nil, opt); r.TotalDiscrepancies != 0 || r.TotalUnknown != 1 {
		t.Fatal(r)
	}
}
func TestAmbiguityAndDuplicateObservations(t *testing.T) {
	p, o := pair()
	o2 := o
	o2.OperationID = "2"
	p.OperationID = ""
	for _, ops := range [][]types.OnChainPayment{{o, o2}, {o2, o}} {
		r := run([]types.InternalPayment{p}, ops, options())
		if r.TotalMatched != 0 || len(r.Results) != 1 || r.TotalUnknown != 1 {
			t.Fatal(r)
		}
	}
	p.OperationID = "1"
	if r := run([]types.InternalPayment{p}, []types.OnChainPayment{o, o}, options()); r.TotalMatched != 1 || len(r.Results) != 1 {
		t.Fatal(r)
	}
	conflict := o
	conflict.Amount = "1"
	if r := run([]types.InternalPayment{p}, []types.OnChainPayment{o, conflict}, options()); r.TotalMatched != 0 {
		t.Fatal(r)
	}
	p2 := p
	p2.ID = "2"
	r := run([]types.InternalPayment{p, p2}, []types.OnChainPayment{o}, options())
	if r.TotalUnknown != 2 || r.TotalMatched != 0 {
		t.Fatal(r)
	}
	r = run([]types.InternalPayment{p, p}, []types.OnChainPayment{o}, options())
	if r.DiscrepancyCounts[types.DiscrepancyDuplicateInternal] != 2 || r.TotalMatched != 0 {
		t.Fatal(r)
	}
}
func TestOrphanRequiresCompleteExport(t *testing.T) {
	_, o := pair()
	opt := options()
	if r := run(nil, []types.OnChainPayment{o}, opt); r.DiscrepancyCounts[types.DiscrepancyOrphanedOnChain] != 1 {
		t.Fatal(r)
	}
	opt.Coverage.InternalComplete = false
	if r := run(nil, []types.OnChainPayment{o}, opt); r.TotalUnknown != 1 || r.TotalDiscrepancies != 0 {
		t.Fatal(r)
	}
}
