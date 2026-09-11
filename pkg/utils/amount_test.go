package utils

import (
	"math/big"
	"testing"
)

func TestParseScaledAmount(t *testing.T) {
	tests := []struct {
		input    string
		decimals int
		expected *big.Int
	}{
		{"100.5", 7, big.NewInt(1005000000)},
		{"0.0000001", 7, big.NewInt(1)},
		{"10", 7, big.NewInt(100000000)},
		{"0", 7, big.NewInt(0)},
	}

	for _, tt := range tests {
		got, err := ParseScaledAmount(tt.input, tt.decimals)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tt.input, err)
		}
		if got.Cmp(tt.expected) != 0 {
			t.Errorf("input %s: got %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFormatScaledAmount(t *testing.T) {
	val := big.NewInt(1005000000)
	formatted := FormatScaledAmount(val, 7)
	if formatted != "100.5000000" {
		t.Errorf("expected 100.5000000, got %s", formatted)
	}

	valSmall := big.NewInt(1)
	formattedSmall := FormatScaledAmount(valSmall, 7)
	if formattedSmall != "0.0000001" {
		t.Errorf("expected 0.0000001, got %s", formattedSmall)
	}
}

func TestAmountDeltaScaled(t *testing.T) {
	delta, err := AmountDeltaScaled("100.50", "100.25", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delta != "0.2500000" {
		t.Errorf("expected 0.2500000, got %s", delta)
	}
}

func TestRejectLossyAndMalformedAmounts(t *testing.T) {
	for _, s := range []string{"", " ", "1.2.3", "1e2", "NaN", "+1", "1.", ".1", "50.123456789", "-1", "0", "922337203685.4775808"} {
		if _, err := ParsePaymentAmount(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
	for _, s := range []string{"0.0000001", "922337203685.4775807"} {
		if _, err := ParsePaymentAmount(s); err != nil {
			t.Error(err)
		}
	}
	if got, err := AmountDeltaScaled("922337203685.4775807", "922337203685.4775806", 7); err != nil || got != "0.0000001" {
		t.Fatal(got, err)
	}
	if got, err := AmountDeltaScaled("1", "2", 7); err != nil || got != "-1.0000000" {
		t.Fatal(got, err)
	}
}
