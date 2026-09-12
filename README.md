# ledger-parity-core

![LedgerParity](assets/lp-banner.png)

[![Go](https://img.shields.io/badge/Go-1.22.2%2B-3FE0C4?style=flat&logo=go&logoColor=white&labelColor=0B0E1E)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/LedgerParity/ledger-parity-core?style=flat&color=7A5CFF)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/LedgerParity/ledger-parity-core/ci.yml?branch=main&style=flat&label=CI&logo=github&labelColor=0B0E1E)](.github/workflows/ci.yml)

See the [LedgerParity documentation](https://ledgerparity.github.io/) for the full organization overview, concepts, quick start, and evidence discipline.

Read-only reconciliation of application payment expectations against Stellar **ordinary classic payment operations**. For backend engineers investigating missed settlement notifications, incorrect amounts or duplicated internal records. Developer preview; no demonstrated operator adoption or production-readiness claim.

Amounts stay exact to one stroop, including at the signed-int64 limit. Matches require the same network passphrase, sender, recipient, asset type/code/issuer and time window. An operation ID disambiguates multiple payments in one transaction. Unproven coverage, ambiguous candidates and malformed observations remain `UNKNOWN`.

Expectations can now use an explicit `settlement_start`/`settlement_end` interval instead of `timestamp`. Supply exactly one representation. Explicit intervals are closed and are not widened by time tolerance; they must fit the reconciliation window. `business_reference` is separate from the transaction-hash `reference_id`. This supports the release-scoped SDP export workflow in the CLI without inventing a settlement timestamp from CSV update time. A separate [RPC equivalence corpus](tools/rpc-corpus/README.md) checks captured transaction evidence; runtime ingestion remains Horizon-only.

From this repository alone, with Go 1.22.2 or newer:

```sh
go test ./...
go vet ./...
go build ./...
```

The offline operator demonstration and JSON/CSV input workflow live in [ledger-parity-cli](https://github.com/LedgerParity/ledger-parity-cli). Core has no external runtime dependencies. Cross-repository consumers must pin the updated core revision; old input records need migration.

Library flow: normalize and validate `types.InternalPayment`; set `HorizonIngestor.Network` to the exact expected passphrase; call `Fetch(ctx, accounts, start, end)`; pass its `Coverage` to `engine.ReconcileOptions`. Fetch a window expanded by `TimeframeToleranceSec` around the internal window. Set `Coverage.InternalComplete` only if the export covers every relevant ordinary payment for the monitored accounts. `Reconcile` accepts incomplete evidence and reports it separately; `Fetch` returns errors for unsuccessful scans. The legacy slice-only `FetchOnChainPayments` wrapper discards coverage and is unsuitable for definitive absence conclusions.

Input migration: require `network`, `operation_type: "payment"`, `asset_type`, sender, recipient, status, timestamp (or explicit settlement interval) and positive decimal-string amount. Native XLM uses `asset_type: "native"`, `asset: "XLM"`, no issuer. Credit assets require a case-sensitive code and issuer. `reference_id` means transaction hash; memo matching was removed. `operation_id` is recommended. Monetary tolerance and asset alias options were removed; matching is exact. `completed`, `success`, `settled` (case-insensitive) declare expected completed settlement. Other statuses with observed settlement produce a status discrepancy; without settlement they remain unresolved.

A complete scan means provider-declared retained history covers the requested window, pages were traversed, and the retention boundary stayed stable. It trusts Horizon and does not prove cryptographic or continuous history. Accounts are exact opaque identifiers; StrKey/checksum validation is not implemented. Muxed identities are preserved, with conservative base-account coverage limitations. Reports distinguish matches, discrepancies and unknowns; unknowns are not counted as financial discrepancies.

Not supported: Soroban/RPC events, contract tokens, path payments, account creation/merge settlement, memo enrichment, transaction signing or remediation. The old `pkg/store` API is not connected to ingestion or CLI: MemoryStore is process-local; SQLiteStore needs a caller-registered driver and stores aggregate reports only. There is no durable automatic checkpoint/resume capability. Scans restart safely instead of advancing an unverified checkpoint.

Read the [audit](docs/GAP_ASSESSMENT.md), [source review](docs/SOURCES.md), [roadmap](ROADMAP.md), [decisions](DECISIONS.md), [handoff](PROJECT_HANDOFF.md), [contributor tasks](docs/backlog.md), [contribution guide](CONTRIBUTING.md) and [security guidance](SECURITY.md). Existing [MIT license](LICENSE) is unchanged. Drips admission is not claimed.
