# LedgerParity readiness audit

Inspected 2026-09-11. Baselines: core d4fe255, connectors fe10911, CLI 7ad01c9. All three trees were clean; no applicable AGENTS.md found. Rejection feedback and original application requested but not supplied. These findings are code observations, **not rejection explanations**.

Intended users are backend engineers operating Stellar classic-payment applications who need to compare an export of expected settlements with account payment history after missed notifications or backend errors. Utility is a repeatable read-only comparison with evidence and uncertainty; adoption and demand remain unvalidated. Horizon supplies settlement observations, not application intent. Existing explorers and Horizon expose history; application-specific reconciliation still requires a mapping and completeness policy. No ecosystem-wide uniqueness claim is established.

| Observed gap | Evidence at baseline | Impact | Fix / acceptance check | Priority | Uncertainty |
| --- | --- | --- | --- | --- | --- |
| Float arithmetic and silent truncation | engine/reconciler.go ParseFloat; utils/amount.go slices fractions | Incorrect matches, lost stroops | Exact seven-place arithmetic; reject malformed/excess precision; large-value regression | P0 | Contract-token decimals out of scope |
| Reference bypasses asset/direction | findBestOnChainMatch reference branch | Wrong recipient/issuer can pass | Network, sender, destination, type, issuer, tx/op identity checked even with references | P0 | Upstream intent must be mapped correctly |
| Single-page Horizon query | fetchAccountPayments ignores next/cursor | Missing older payments | Bounded cursor traversal, failures return errors, retention/window coverage explicit | P0 | Provider honesty/history continuity assumed |
| Endpoint failure becomes absence | CLI warning then empty onChains | False missing settlements | Error exit and no successful report on ingestion failure | P0 | Availability varies |
| Greedy matching and duplicate observations | first reference match, no op dedup | Arbitrary settlement attribution | Reject conflicting observations; multiple candidates/shared candidates uncertain | P0 | Batch allocation deferred |
| Implicit incomplete-data conclusions | no coverage model; pending status ignored | False missing/orphan reports | Separate UNKNOWN and discrepancy; absence requires declared coverage | P0 | Application export completeness is operator assertion |
| Synthetic named adapters advertised as compatible | local structs, GitHub search links; no versioned contracts | Misleading integrations | Label experimental; canonical file route default; reject lossy values/timestamps | P1 | Upstream contracts need dedicated verification |
| CLI setup/config gaps | no CLI tests; config output ignored; implicit demo; fixed 24h window | Reviewer cannot trust automation | Explicit demo, strict JSON config, actual flags/output, reproducible offline fixture and failure tests | P1 | Live testnet evidence separate |
| Unsupported persistence/Soroban claims | no RPC code; SQLite requires external driver; no CLI store usage | Overstates maturity | Correct capability docs; defer durable ingestion until atomic checkpoint design tested | P1 | No production readiness |
| Thin contributor readiness | core contribution guide claims Wave participation | Misleading program status | Bounded tasks, contributor/security guidance, reviewer and appeal draft | P1 | Maintainer capacity unknown |

Repository boundaries remain core (types/matching/Horizon), connectors (application input), CLI (operator workflow). They are existing boundaries, not a funding strategy. Independently buildable dependency pins are required; no new repositories or license changes.

Reusable reference lessons: inspect before implementing; check official contracts and dates; implement the smallest complete workflow; test missing/conflicting evidence; distinguish fixture/local/live/remote evidence; maintain decisions and acceptance gates. ProofGrid product, architecture, license, permissions and test results do not transfer.

## Applied fixes and remaining evidence

The P0/P1 implementation slice is now in code: engine/types/utils implement exact identity and conservative results; ingest has cursor traversal and coverage; connectors validate canonical input and preserve exact amounts; CLI propagates errors, honors config/flags and includes an offline demo. Tests exercise these behaviors and their failure cases. SQLite/checkpoint and Soroban claims were removed, not falsely implemented. Named adapter compatibility remains unverified and labeled experimental. See CLI docs/LIVE_TESTNET_RESULT.json for a narrowly labeled read-only API check, and PROJECT_HANDOFF.md for executed versus unverified checks.

Feedback update: the maintainer recalls a concern about making LedgerParity clearly Stellar-based with practical ecosystem impact. Written feedback and original application remain unavailable. This supports prioritizing the payment-operator workflow; it does not prove every listed gap caused rejection. No user adoption, measured impact or organizer acceptance is established. Validation with an actual operator is the next product gate.
