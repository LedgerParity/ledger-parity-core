# Roadmap

This is a bounded engineering roadmap, not an eligibility or impact claim.

1. **Correct classic-payment comparison (current milestone).** Exact stroops; full asset/network/direction identity; operation-aware matching; duplicate/ambiguous data stays reviewable. Acceptance: one-stroop-at-int64-limit, issuer, direction, duplicate and failed-transaction regressions pass.
2. **Coverage and reproducible operator workflow (current milestone).** Paginated read-only Horizon scan; retention/freshness bounds; errors never become absence; strict file input and CLI; deterministic offline end-to-end demo. Acceptance: >200 operations, later-page failure, rate limit, incomplete window, config/output failures and isolated module builds pass. Commit and push all three owning repositories.
3. **Contributor/reviewer readiness (current milestone).** Accurate READMEs, evidence audit, contributor tasks, reviewer walkthrough and feedback-aware appeal draft. Acceptance: new reviewer can run the documented demo without sibling checkouts, accounts or secrets; limitations visible in report/docs.
4. **Validate practical impact (next).** Find an operator willing to provide an anonymized classic-payment export specification and a known discrepancy. Acceptance: operator reproduces a useful finding, documents integration effort and identifies false positives; publish only with consent. No user/adoption claim before evidence.
5. **Durable ingestion and broader semantics (deferred).** Atomic observation/report/checkpoint persistence with network/account/provider scope; crash/restart/idempotency tests. Path payments need explicit source-versus-delivered amounts; Soroban needs contract/event identity, decimals and retention semantics. Acceptance: separate design and failure fixtures before advertising support. No signing or fund movement.

See [bounded tasks](docs/backlog.md), [audit](docs/GAP_ASSESSMENT.md), and [decisions](DECISIONS.md). Current check results belong in PROJECT_HANDOFF.md, not future milestone promises.

## Implemented milestone status

- [x] Exact amounts and explicit network/operation/asset/direction contract.
- [x] Conservative ambiguity, duplicate and incomplete-evidence results; candidate IDs retained.
- [x] Paginated Horizon scan and negative ingestion tests; errors never become empty successful scans.
- [x] Canonical connectors and reproducible CLI workflow, including mock Horizon success/failure and optional read-only testnet evidence.
- [x] Contributor/reviewer documentation and truthful appeal draft in CLI.
- [ ] Real operator validation, durable checkpointing and broader payment/event support.
- [ ] Remote CI confirmation and local race verification; current environment/evidence limits are in PROJECT_HANDOFF.md.
