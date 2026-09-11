# Bounded contributor tasks

Each task needs a maintainer to confirm availability before assignment. No points, activity or response-time promises are invented. Submit a small PR with reproduction, test evidence and remaining limits; maintainer reviews behavior and source semantics before merge.

1. **Crash-safe ingestion checkpoint design and prototype.** Current scans restart from a closed window; legacy store has no atomic observation/checkpoint transaction. Deliver an interface scoped by network/account/provider and a fault-injected in-memory reference test. Acceptance: failures before/after durable report commit never advance past lost observations; replay is idempotent. Persistent driver choice and production migration are a separate PR.
2. **Path-payment contract.** Current ingestion skips path payments. Deliver a proposal and offline Horizon fixtures distinguishing source asset/amount from delivered asset/amount for both path-payment types. Acceptance: wrong source amount cannot satisfy delivered amount; issuer and direction are explicit. No RPC/CLI expansion in this task.
3. **Muxed-account scope verification.** Current equality preserves M addresses; coverage from base G accounts can remain conservatively unknown for all-muxed records. Verify official address/account semantics, propose base-account-plus-muxed-ID representation using maintained Stellar tooling, test distinct muxed IDs and base-account scans. Never collapse recipient identities.
4. **Independent provider review.** Audit retention/freshness assumptions against a pinned Horizon implementation; add fixtures for changing history bounds and inconsistent providers. Acceptance: conflicting or unproven evidence cannot create definitive absence. Do not add a second service dependency without a separate proposal.
5. **Scale measurement.** Matching currently examines each internal/operation pair. Deliver reproducible synthetic benchmarks at 1k/10k records and candidate-index design; preserve order-independent ambiguity and exact identity tests. Optimize only with benchmark evidence.


## Implementation update 2026-09-12

Next bounded tasks: add a real credit-asset RPC/Horizon pair and a fee-bump pair with negative issuer/hash cases; design retained-ledger coverage and crash/restart semantics before RPC ingestion. Coordinate interval support with consumers. Real operator validation and durable checkpoints remain open.
