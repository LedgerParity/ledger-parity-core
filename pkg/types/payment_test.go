package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeRepresentationsJSON(t *testing.T) {
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range []InternalPayment{{Timestamp: now}, {SettlementStart: now, SettlementEnd: now.Add(time.Hour)}} {
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		json.Unmarshal(raw, &fields)
		_, timestamp := fields["timestamp"]
		_, interval := fields["settlement_start"]
		if timestamp == interval {
			t.Fatal("expected one representation", string(raw))
		}
		var back InternalPayment
		if err = json.Unmarshal(raw, &back); err != nil {
			t.Fatal(err)
		}
		if !back.Timestamp.Equal(p.Timestamp) || !back.SettlementStart.Equal(p.SettlementStart) || !back.SettlementEnd.Equal(p.SettlementEnd) {
			t.Fatal("time roundtrip")
		}
	}
}
