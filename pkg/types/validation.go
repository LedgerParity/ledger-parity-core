package types

import (
	"fmt"
	"github.com/LedgerParity/ledger-parity-core/pkg/utils"
	"regexp"
)

var assetCodePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateAsset(kind, code, issuer, contract string) error {
	if contract != "" {
		return fmt.Errorf("contract tokens are unsupported")
	}
	switch kind {
	case "native":
		if code == "XLM" && issuer == "" {
			return nil
		}
	case "credit_alphanum4", "credit_alphanum12":
		max, min := 4, 1
		if kind == "credit_alphanum12" {
			max, min = 12, 5
		}
		if len(code) >= min && len(code) <= max && assetCodePattern.MatchString(code) && issuer != "" {
			return nil
		}
	}
	return fmt.Errorf("incomplete or invalid classic asset identity")
}

func ValidateInternal(p InternalPayment) error {
	if p.OperationType != "payment" {
		return fmt.Errorf("only ordinary classic payment expectations are supported")
	}
	if p.ID == "" || p.Network == "" || p.Sender == "" || p.Recipient == "" || p.Timestamp.IsZero() || p.Status == "" {
		return fmt.Errorf("missing internal identity, direction, timestamp or status")
	}
	if _, err := utils.ParsePaymentAmount(p.Amount); err != nil {
		return err
	}
	return ValidateAsset(p.AssetType, p.Asset, p.AssetIssuer, p.AssetContract)
}

func ValidateOnChain(p OnChainPayment) error {
	if p.Network == "" || p.OperationID == "" || p.TransactionHash == "" || p.Account == "" || p.Destination == "" || p.Timestamp.IsZero() || p.OperationType != "payment" {
		return fmt.Errorf("incomplete identity or unsupported operation type")
	}
	if _, err := utils.ParsePaymentAmount(p.Amount); err != nil {
		return err
	}
	return ValidateAsset(p.AssetType, p.AssetCode, p.AssetIssuer, p.AssetContract)
}
