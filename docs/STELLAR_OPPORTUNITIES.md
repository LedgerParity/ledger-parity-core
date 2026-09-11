# Stellar opportunity assessment

Researched 2026-09-11. Recommendations are hypotheses, not established demand, compatibility, adoption or Drips eligibility.

Implementation update 2026-09-12: explicit expectation intervals, SDP 7.0.0 CSV mapping and CLI evidence replay are implemented in their owning repositories. The [RPC corpus](../tools/rpc-corpus/README.md) has a captured native-payment case and synthetic negative/batch tests; the full migration gate remains open. The research below preserves the original rationale and planned acceptance checks. No operator adoption was established.

## Recommended direction

Build independent, reproducible settlement verification for Stellar Disbursement Platform (SDP) operators. The intended user is an operations engineer or finance reviewer checking exported payment obligations against observed Stellar payments. Preserve business references, exact assets, candidate operations and evidence coverage so another reviewer can reproduce each conclusion.

SDP already submits and tracks payments. LedgerParity should complement that with an independent check and portable evidence, rather than another status dashboard. No operator has validated this hypothesis, and this research does not establish that the offering is unique.

## Evidence and alternatives

| Opportunity | Primary evidence | Assessment |
| --- | --- | --- |
| SDP export reconciliation | [Payments UI](https://developers.stellar.org/docs/platforms/stellar-disbursement-platform/admin-guide/user-interface/payments) supports CSV export with account/search/filter scope. [Architecture](https://developers.stellar.org/docs/platforms/stellar-disbursement-platform/admin-guide/design-and-architecture) already provides submission and tracking. | Strong fit with existing ordinary-payment engine. Independent evidence is the proposed additional value; operator demand remains unverified. |
| Horizon/RPC migration verification | [API overview](https://developers.stellar.org/docs/data/apis) says Horizon is nearing end-of-life. [Migration guide](https://developers.stellar.org/docs/data/apis/migrate-from-horizon-to-rpc) explains payment reconstruction using events and transaction metadata. | Prioritize a paired ordinary-payment corpus before an RPC ingestor. Account payment history is not a simple endpoint substitution; event availability and operation mapping matter. |
| Anchor lifecycle reconciliation | [Anchor Platform API](https://developers.stellar.org/docs/platforms/anchor-platform/api-reference/platform/rpc/methods/get_transactions) includes lifecycle states, external references and multiple Stellar transactions. | Potentially useful, but refunds/multiple legs require a new business model beyond current one-record matching. Defer. |
| General token/accounting platform | Current code deliberately supports ordinary classic payments only. | No investigated evidence justifies broadening into contracts, dashboards or fund movement. |

This compares adjacent official platforms, not every competitor. Search results cannot prove an equivalent tool does not exist.

## Source-level SDP contract findings

Read upstream backend revision `bc26006c7d8429939c671fd879754ae7e8c863cc`, the default-branch HEAD returned during research, not a promised supported release:

- [Export handler](https://github.com/stellar/stellar-disbursement-platform-backend/blob/bc26006c7d8429939c671fd879754ae7e8c863cc/internal/serve/httphandler/export_handler.go)
- [Export tests](https://github.com/stellar/stellar-disbursement-platform-backend/blob/bc26006c7d8429939c671fd879754ae7e8c863cc/internal/serve/httphandler/export_handler_test.go)

Source/tests were inspected over HTTPS; upstream tests were not run. Tests assert exported payment identity, string amount, transaction reference, status/type, asset code/issuer, receiver address, creation/update times and external payment reference. The handler also exports contact information and Circle identifiers. It does not export network, sending account, operation ID or a dedicated settlement timestamp. It applies caller wallet visibility as well as request filters. Successful download therefore does not establish account-wide completeness.

Proposed mapping, not implemented compatibility:

| Upstream information | Required treatment |
| --- | --- |
| Payment ID | Source record ID scoped to deployment/export origin |
| Amount, asset code and issuer | Exact decimals and explicit native/credit identity; reject contradictory fields |
| Receiver address | Recipient; missing address cannot become a settled record |
| Stellar transaction reference | Chain reference; multiple candidate operations remain ambiguous |
| External payment reference | Separate business metadata, never a chain hash |
| Network and sender | Explicit independently supplied scope; do not infer sender from the chain payment being checked |
| Creation/update times | Preserve original semantics; obtain a justified settlement interval rather than relabel either as settlement time |
| Circle route | Unsupported in first adapter; fail explicitly |

Exclude receiver phone/email from normalized records/reports. Pin a supported release and verify its export before claiming compatibility. The current canonical schema may need settlement-interval and business-metadata extensions first. Do not force a mapping that manufactures missing facts.

## Phased implementation and acceptance checks

1. **Contract and scope, connectors.** Select a release, record its schema provenance and generate synthetic contract fixtures. Cover exact amounts, issuer changes, missing recipient, duplicate IDs, multi-operation hashes, filtered exports and unsupported routes. Acceptance: independently supplied sender/network, justified time semantics, completeness false by default, and compatibility limited to the tested release. If CSV cannot support these requirements, investigate a documented API export instead of weakening matching.
2. **Replayable evidence, CLI/core.** Version the report envelope; hash source bytes and normalized observations, preserve tool revisions/configuration/coverage, and support offline re-evaluation. Acceptance: changed inputs change hashes; identical evidence reproduces findings; incomplete evidence stays UNKNOWN; secrets/contact data are excluded. Hashes establish byte identity, not source truth or cryptographic history proof.
3. **RPC equivalence corpus, core.** Pair Horizon observations with RPC transaction metadata/events for the same ordinary payments. Test batches, failed transactions, duplicated observations, missing classic events and retention gaps. Acceptance: explicit operation mapping, no event/operation double counting, identical supported settlement results and visible unsupported cases. Implement a read-only RPC ingestor only after this gate; do not advertise general Soroban support.
4. **Operator validation.** With consent, compare an independently produced sanitized export containing a known issue and a clean control. Record setup time, reviewed findings, false positives and unresolved cases. Acceptance: another reviewer reproduces a useful conclusion without editing Go code. A chain-derived synthetic expected record does not meet this gate.

These are bounded contributor tasks. Each needs source evidence, a reproduction, negative cases and scope limits. No outreach was sent and no operator endorsement obtained. Stop or change direction if the export cannot support reliable comparison or an operator cannot identify additional value over existing SDP tools.

## Drips relevance and evidence limits

The [official maintainer rules](https://docs.drips.network/wave/maintainers/participating-in-a-wave/) require organizer approval and substantive changes for an appeal. They do not guarantee acceptance for a particular feature. The [Stellar Wave page](https://www.drips.network/wave/stellar) showed no active/upcoming Waves when checked. This roadmap is engineering judgment, not an official eligibility checklist.

Original feedback remains unavailable; the maintainer recalls concerns about Stellar relevance and impact. An eventual appeal should cite shipped functionality and independently validated usefulness. SDP/RPC work above is planned, not delivered. This research adds no live integration or production evidence; prior test and CI results retain their recorded revision boundaries.
