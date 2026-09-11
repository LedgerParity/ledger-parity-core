# Project handoff

2026-09-11. Current milestone: exact, coverage-aware ordinary Stellar payment reconciliation. Baseline tests passed in all three repositories, but baseline Horizon/file/CLI packages lacked tests. Read docs/GAP_ASSESSMENT.md, DECISIONS.md and ROADMAP.md for evidence and scope.

Implemented: strict decimal/stroop validation; required network/operation/asset identity; conservative ambiguous/duplicate handling; UNKNOWN separately counted; paginated Horizon reads with bounded retries, validation, retention/freshness evidence and failure propagation. Added engine and ingestion regressions including >200 payments. No fund movement, new repositories or license changes.

Local Go is C:/Go/bin/go.exe (1.24.4); it was not on PATH. Initial root-level go test failed because the workspace root is not a module; run inside each repository. Core local test/vet checks passed during development. Final build/race and remote CI results will be recorded in the completion update. No inherited ProofGrid tests or permissions are treated as LedgerParity evidence.

Maintainer recollection: rejection concerned Stellar relevance/impact. Original feedback/application unavailable. No exact rejection explanation, operator adoption or admission is claimed. Continue downstream canonical connectors and CLI demo, pin released source revisions, run isolated builds/tests, then record push/CI evidence.

Core verification before first implementation commit: `go test ./...`, `go vet ./...`, and `go build ./...` passed on Go 1.24.4. `go test -race ./...` initially could not start because CGO is disabled; compiler availability is being checked. New remote CI is not yet verified. Deterministic fixtures are not live settlement evidence.

## Completion update (supersedes in-progress notes above)

Core runtime revisions 48e0036, bbd6faf and 89aa1ec were committed and pushed normally. The final review preserved candidate operation IDs, made competing duplicate claims block arbitrary matching, and classified unreferenced nearby amount differences as UNKNOWN. Tests, vet and build passed locally on Go 1.24.4. Isolated archive verification uses GOWORK=off and a fresh module cache; its result is recorded below when complete.

The CLI now has an offline end-to-end demo and a successful read-only public-testnet check (CLI docs/LIVE_TESTNET_RESULT.json). That check derives a synthetic expectation from the observed payment and is not real application adoption. Automatic durable checkpointing, path payments, Soroban, verified product adapters, provider-independent history and production operation remain out of scope.

CI now covers minimum Go 1.22.2 and the supported stable alias, with manual dispatch, builds and race tests. Updated action runtimes follow official actions/setup-go usage (https://github.com/actions/setup-go, reviewed 2026-09-11). Remote runs remain unverified: the initial public API returned zero runs; later requests hit rate limits. No green remote CI claim. Local -race could not start: CGO disabled/no C compiler; a portable compiler download was cancelled after slow progress, its partial archive removed, and no compiler installed.

The bounded implementation milestone is complete subject to the separately stated verification limits. Next: obtain a consenting operator's sanitized export/known discrepancy and validate practical utility; inspect signed-in Actions and Drips appeal state. The original rejection/application are absent; the user's recollection concerns Stellar relevance/impact. No submission, promise of approval, fund movement, new remote repositories or license change.

Isolated verification passed from an exported commit archive with GOWORK=off and a fresh module cache: 37 passing test cases including subtests, go vet and go build. This tests core runtime 89aa1ec independently of sibling working trees. Later changes in this completion commit affect documentation/CI only.
