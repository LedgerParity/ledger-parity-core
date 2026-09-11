package utils

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// ParseScaledAmount parses a decimal string into a big.Int scaled by decimals (e.g. 7 for Stellar XLM stroops).
func ParseScaledAmount(amountStr string, decimals int) (*big.Int, error) {
	if decimals < 0 || decimals > 18 || !decimalPattern.MatchString(amountStr) {
		return nil, fmt.Errorf("invalid decimal amount or scale")
	}

	parts := strings.Split(amountStr, ".")
	wholePart := parts[0]
	fractionPart := ""
	if len(parts) > 1 {
		fractionPart = parts[1]
	}

	if len(fractionPart) > decimals {
		return nil, fmt.Errorf("amount exceeds %d decimal places", decimals)
	} else {
		fractionPart = fractionPart + strings.Repeat("0", decimals-len(fractionPart))
	}

	combined := wholePart + fractionPart
	combined = strings.TrimLeft(combined, "0")
	if combined == "" {
		return big.NewInt(0), nil
	}

	res, ok := new(big.Int).SetString(combined, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount numeric string: %s", amountStr)
	}

	return res, nil
}

// FormatScaledAmount converts a scaled big.Int back to a standard decimal string representation.
func FormatScaledAmount(scaled *big.Int, decimals int) string {
	if decimals == 0 && scaled != nil {
		return scaled.String()
	}
	if scaled == nil {
		return "0"
	}

	str := scaled.String()
	negative := false
	if strings.HasPrefix(str, "-") {
		negative = true
		str = str[1:]
	}

	if len(str) <= decimals {
		str = strings.Repeat("0", decimals-len(str)+1) + str
	}

	splitIdx := len(str) - decimals
	whole := str[:splitIdx]
	frac := str[splitIdx:]

	res := fmt.Sprintf("%s.%s", whole, frac)
	if negative {
		res = "-" + res
	}
	return res
}

var decimalPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)

// ParsePaymentAmount enforces Stellar classic's positive signed-int64 stroop range.
func ParsePaymentAmount(s string) (*big.Int, error) {
	n, err := ParseScaledAmount(s, 7)
	if err != nil {
		return nil, err
	}
	if n.Sign() <= 0 || !n.IsInt64() {
		return nil, fmt.Errorf("payment amount must be positive and fit int64 stroops")
	}
	return n, nil
}

// AmountDeltaScaled returns the difference (a - b) as a formatted decimal string.
func AmountDeltaScaled(amtA, amtB string, decimals int) (string, error) {
	scaledA, err := ParseScaledAmount(amtA, decimals)
	if err != nil {
		return "", err
	}
	scaledB, err := ParseScaledAmount(amtB, decimals)
	if err != nil {
		return "", err
	}

	delta := new(big.Int).Sub(scaledA, scaledB)
	return FormatScaledAmount(delta, decimals), nil
}
