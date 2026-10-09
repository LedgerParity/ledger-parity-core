# Exact Stellar reconciliation for backend developers

Prepared for the October 9, 2026 Stellar Wave submission.

## Purpose and implemented utility

The Go library matches internal payment expectations against Horizon observations using exact amounts, network and asset identity, settlement windows, and conservative coverage/ambiguity handling. The released CLI is an existing consumer.

## Reproduce the implementation

Go 1.22.2+; run `go test ./...`, `go vet ./...`, `go build ./...`. Tests and the CLI consumer demonstrate supported behavior without a sibling checkout.

## Evidence and supported scope

Local tests, vet, and build passed October 7. Runtime ingestion covers ordinary classic payments. Path payments, Soroban events and durable checkpoints remain separately scoped work; uncertain coverage stays UNKNOWN.

Evidence reference: [https://github.com/LedgerParity/ledger-parity-cli](https://github.com/LedgerParity/ledger-parity-cli).
Baseline source revision: `f9f665a550988b6b59b7ed05c56eb5cd2dc35716`. Final reviewed preparation revision and CI
results belong in [VERIFICATION_OCT09.md](VERIFICATION_OCT09.md).

## Maintainers and contributor work

Maintainers: xteesamz and EthTobi; contact via GitHub, available anytime.
See [MAINTAINERS.md](../MAINTAINERS.md), [CONTRIBUTING.md](../CONTRIBUTING.md),
[SECURITY.md](../SECURITY.md), and [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md).
The [focused engineering backlog](WAVE_BACKLOG.md) describes real work, relevant
files, tests, and acceptance criteria. Draft complexity values require maintainer
review and app enrollment; they do not establish approval or earned points.

## Before applying

- Confirm the Drips Wave App covers this repository and check application slots.
- Publish the reviewed backlog issues and preserve links to their acceptance checks.
- Publish these preparation changes through a reviewed PR with passing CI.
- Apply under the implemented scope above; no production/adoption claims are implied.
