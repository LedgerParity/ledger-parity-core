# Focused engineering backlog

Six issue drafts reviewed against current source and existing open issues for the
October 9 submission. These are engineering tasks, not approval or earned points.
Proposed complexity is subject to maintainer review and Drips app configuration.

## 1. Define muxed-account identity fixtures for reconciliation

## Description & Context

Current account equality preserves M-address identity but base-account coverage needs a pinned semantics contract.

## Proposed Complexity

High (200 points proposed); planning label `complexity: high`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Cite official muxed-account semantics and version.
- [ ] Distinguish the base G account and muxed identifier.
- [ ] Add negative cases for different muxed IDs.
- [ ] Never collapse identities to create false matches.
- [ ] Scope implementation separately from ingestion expansion.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

pkg/types/; pkg/engine/; pkg/ingest/

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## 2. Specify delivered-asset matching for Stellar path payments

## Description & Context

Only ordinary classic payments are supported; path payments need explicit delivered-versus-source semantics.

## Proposed Complexity

High (200 points proposed); planning label `complexity: high`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Capture authoritative fixtures for strict-send and strict-receive.
- [ ] Specify delivered/source asset and amount fields.
- [ ] Reject wrong issuer/direction and wrong source-amount matches.
- [ ] Preserve existing ordinary-payment behavior.
- [ ] Propose the mapping before runtime expansion.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

pkg/types/; pkg/ingest/; tests/fixtures/

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## 3. Benchmark candidate indexing without changing ambiguity

## Description & Context

Pairwise matching needs scale evidence before optimization.

## Proposed Complexity

Medium (150 points proposed); planning label `complexity: medium`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Benchmark deterministic 1k/10k datasets.
- [ ] Test duplicate IDs, ambiguous candidates and shuffled input.
- [ ] Propose identity/time indexes with equivalent outcomes.
- [ ] Require benchmark evidence for any optimization.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

pkg/engine/; benchmarks/

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## 4. Extend provider freshness and retention-boundary fixtures

## Description & Context

Definitive absence depends on bounded and trustworthy provider coverage.

## Proposed Complexity

Medium (150 points proposed); planning label `complexity: medium`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Pin the relevant Horizon/RPC response contract.
- [ ] Test stale responses, truncated history and inconsistent retention bounds.
- [ ] Return incomplete coverage when freshness cannot be proven.
- [ ] Preserve UNKNOWN rather than infer absence.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

pkg/ingest/; tools/rpc-corpus/

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## 5. Prototype an atomic observation/checkpoint store contract

## Description & Context

Existing ingestion can restart a window and lacks an atomic durable checkpoint contract.

## Proposed Complexity

High (200 points proposed); planning label `complexity: high`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Link the broader checkpoint work in issue #3.
- [ ] Specify network/account/provider checkpoint identity.
- [ ] Add an in-memory fault-injected reference with replay idempotency.
- [ ] Prove failures before/after commit cannot skip observations.
- [ ] Defer the SQL driver/migration to a separately scoped change.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

pkg/store/; pkg/ingest/; docs/

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## 6. Add fee-bump RPC/Horizon equivalence fixtures

## Description & Context

Captured equivalence evidence needs fee-bump-specific negative identity cases.

## Proposed Complexity

Medium (150 points proposed); planning label `complexity: medium`.
Actual enrollment and points must be set in the Drips app after approval.

## Requirements & Acceptance Criteria

- [ ] Capture a permitted fee-bump transaction and corresponding operations with provenance.
- [ ] Distinguish inner and outer transaction identity.
- [ ] Test mismatched hash, account and asset issuer.
- [ ] Keep runtime ingestion Horizon-only unless separately designed.
- [ ] Include relevant positive/negative regression evidence; required CI passes.

## Relevant Files & Architecture

tools/rpc-corpus/; docs/VERIFICATION.md

## Verification

`go test ./...; go vet ./...; go build ./...`. Network checks remain opt-in.

## Contribution Guidelines

Agree bounded scope with xteesamz or EthTobi through GitHub. Use a focused PR
with `Closes #<issue_id>`, actual check results, and remaining limitations.

## Published issue links

- [Define muxed-account identity fixtures for reconciliation](https://github.com/LedgerParity/ledger-parity-core/issues/4)
- [Specify delivered-asset matching for Stellar path payments](https://github.com/LedgerParity/ledger-parity-core/issues/5)
- [Benchmark candidate indexing without changing ambiguity](https://github.com/LedgerParity/ledger-parity-core/issues/6)
- [Extend provider freshness and retention-boundary fixtures](https://github.com/LedgerParity/ledger-parity-core/issues/7)
- [Prototype an atomic observation/checkpoint store contract](https://github.com/LedgerParity/ledger-parity-core/issues/8)
- [Add fee-bump RPC/Horizon equivalence fixtures](https://github.com/LedgerParity/ledger-parity-core/issues/9)

These issues are published but have not been enrolled into Drips Wave or assigned.
