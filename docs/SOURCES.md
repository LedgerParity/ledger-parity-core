# Official-source review

Reviewed 2026-09-11. Links are source entry points; no copied third-party code or product compatibility guarantee.

- [Stellar assets](https://developers.stellar.org/docs/learn/fundamentals/stellar-data-structures/assets): classic asset identity and signed int64 amounts scaled by 10^7 underpin exact validation. Contract-token support is separate.
- [Payment object](https://developers.stellar.org/docs/data/apis/horizon/api-reference/resources/payments/object), [payment collection](https://developers.stellar.org/docs/data/apis/horizon/api-reference/list-all-payments), [pagination arguments](https://developers.stellar.org/docs/data/apis/horizon/api-reference/structure/pagination/page-arguments): operation-level records and cursor traversal. The collection can include multiple payment-related types; this preview selects ordinary payment only.
- [Horizon testnet root](https://horizon-testnet.stellar.org/): direct read during inspection returned network_passphrase, history_elder_ledger and history_latest_ledger. Those provider-declared bounds are checked against ledger close times. This is not proof of uninterrupted history or production trustworthiness.
- [RPC getEvents](https://developers.stellar.org/docs/data/apis/rpc/api-reference/methods/getEvents): a distinct event API; no corresponding LedgerParity RPC/event adapter exists. No Soroban support advertised.
- [Drips maintainer onboarding and appeals](https://docs.drips.network/wave/maintainers/participating-in-a-wave/): public repositories apply to relevant programs; organizers approve. Rejected repositories use the dashboard appeal, with substantive changes, first-appeal two-week wait, later one-month cooldown and at most three appeals. Recheck dashboard eligibility when acting.
- [Stellar Wave page](https://www.drips.network/wave/stellar): no active/upcoming Wave displayed at review time. This does not establish whether applications/appeals are unavailable.
- [Application limits](https://docs.drips.network/wave/maintainers/repo-application-limits/) and [terms](https://docs.drips.network/wave/terms-and-rules/): limits are program-dependent; do not manipulate points or submit low-quality untested work. Actual account allowances and rejection date remain unavailable.

Exact monetary comparisons, negative tests, standalone builds and bounded tasks are our engineering acceptance criteria, not invented Drips admission rules. No current eligibility determination or submission was made.
