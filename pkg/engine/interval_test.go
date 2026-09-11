package engine_test

import (
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
	"testing"
	"time"
)

func TestSettlementInterval(t *testing.T) {
	p, o := pair()
	p.Timestamp = time.Time{}
	p.OperationID = ""
	p.SettlementStart = now.Add(-time.Minute)
	p.SettlementEnd = now.Add(time.Minute)
	opt := options()
	opt.Coverage.Start = p.SettlementStart
	opt.Coverage.End = p.SettlementEnd
	r := run([]types.InternalPayment{p}, []types.OnChainPayment{o}, opt)
	if r.TotalMatched != 1 || r.Results[0].TimeDeltaSec != 0 {
		t.Fatalf("interval: %+v", r)
	}
	opt.Coverage.Start = now
	if r = run([]types.InternalPayment{p}, []types.OnChainPayment{o}, opt); r.TotalUnknown == 0 {
		t.Fatal("partial interval accepted")
	}
	opt = options()
	o.Timestamp = p.SettlementEnd.Add(time.Second)
	if r = run([]types.InternalPayment{p}, []types.OnChainPayment{o}, opt); r.TotalMatched != 0 {
		t.Fatal("tolerance widened explicit interval")
	}
	for _, edit := range []func(*types.InternalPayment){
		func(p *types.InternalPayment) { p.Timestamp = now },
		func(p *types.InternalPayment) { p.SettlementEnd = time.Time{} },
		func(p *types.InternalPayment) { p.SettlementEnd = p.SettlementStart.Add(-time.Second) },
	} {
		q := p
		edit(&q)
		if types.ValidateInternal(q) == nil {
			t.Fatal("invalid interval accepted")
		}
	}
	p.SettlementStart = now.Add(-2 * time.Hour)
	if r = run([]types.InternalPayment{p}, nil, opt); r.Results[0].Discrepancy != types.DiscrepancyInvalid {
		t.Fatal("interval outside report accepted")
	}
}
