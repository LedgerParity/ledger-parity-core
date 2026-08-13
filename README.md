# ledger-parity-core

[![Go Reference](https://pkg.go.dev/badge/github.com/LedgerParity/ledger-parity-core.svg)](https://pkg.go.dev/github.com/LedgerParity/ledger-parity-core)
[![CI](https://github.com/LedgerParity/ledger-parity-core/actions/workflows/ci.yml/badge.svg)](https://github.com/LedgerParity/ledger-parity-core/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`ledger-parity-core` is the core Go library behind **LedgerParity** — a read-only reconciliation service that automatically cross-checks a Stellar application's internal payment records against actual on-chain ledger settlement data.

---

## 🌟 Overview

Applications built on Stellar (payroll systems, escrow platforms, rental deposit services) keep their own internal record of what they believe happened. The Stellar ledger keeps a separate, independent record of what actually settled on-chain. These two records can drift apart: backend bugs, missed webhook notifications, duplicate payment submissions, or unconfirmed transactions.

`ledger-parity-core` provides the core matching engine, Stellar Horizon & Soroban RPC ingestion logic, and embedded state persistence to detect and categorize discrepancies before they become user-facing incidents.

---

## 🚀 Key Features

- **Multi-Pass Reconciliation Engine:** Matches records by explicit Reference IDs or fuzzy recipient, asset, amount, and timeframe tolerance windows (`±10 mins` default).
- **Categorized Discrepancies:**
  - `MISSING_ON_CHAIN`: Internal record exists, but no corresponding on-chain settlement found.
  - `AMOUNT_MISMATCH`: On-chain transaction exists for recipient/time, but amount differs.
  - `DUPLICATE_INTERNAL`: Multiple internal records map to a single on-chain transaction.
  - `ORPHANED_ON_CHAIN`: On-chain payment detected for target account with no internal record.
- **On-Chain Ingestion:** Direct Stellar Horizon REST API client & Soroban RPC event ingestion.
- **Embedded Persistence:** Embedded SQLite store (`pkg/store`) for caching match states and checkpointing processed ledgers across runs.
- **Zero Remediation / Safe Read-Only:** Operates in strict read-only mode — no write keys required, no fund movement, no automated transaction submission.

---

## 💡 Differentiation

- **vs. `TrapTrace/soroban-error-index`:** That project is a static knowledge base of Soroban error strings and fixes. `LedgerParity` performs real-time operational reconciliation between internal database records and on-chain settlements.
- **vs. `sorolens`:** `sorolens` provides contract-execution observability (storage TTL, event monitoring, invocation health). `LedgerParity` is specifically focused on financial record matching across two systems.
- **vs. `sorokeep`:** `sorokeep` is an operations layer for deployed contracts. `LedgerParity` is a standalone cross-check tool between off-chain application DBs and on-chain Stellar payments.

---

## 📦 Installation & Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/engine"
	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

func main() {
	rec := engine.NewReconciler()
	
	// Sample internal payments and on-chain payments
	internals := []types.InternalPayment{ /* ... */ }
	onChains := []types.OnChainPayment{ /* ... */ }

	now := time.Now()
	report := rec.Reconcile("my_app", now.Add(-24*time.Hour), now, internals, onChains)

	fmt.Println(report.Summary())
}
```

---

## 🧪 Testing

Run the full unit test suite:

```bash
go test -v ./...
```

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines, testing policies, and Drips Wave issue tagging standards.

---

## 📄 License

[MIT License](LICENSE) © LedgerParity Maintainers.