# Ordinary-payment RPC equivalence corpus

This developer tool compares public Horizon observations with a decoded RPC envelope, successful operation results and V4 operation-scoped classic asset events. It is **not a Go RPC ingestor, account indexer, or general Soroban implementation**. The Go engine retains no external runtime dependencies; Node is only needed for this separate corpus check.

Requires Node 22.12+ (CI uses 24). From this directory:

```sh
npm ci --ignore-scripts --no-fund
npm test
```

Installation downloads locked dependencies; the tests themselves are offline. SDK 17.0.1 is pinned in package.json/package-lock.json. The committed testnet.json contains actual public-provider responses with capture time, network and endpoint provenance. It contains public account/payment data and is not an independently obtained application export. Tests also synthesize a two-payment batch and mutate failures/identity/event availability; these mutations were never submitted to Stellar.

Optional fresh public-testnet reads:

```sh
npm run capture -- fresh-testnet.json
```

Capture creates a new file only and can fail if no supported recent transaction is available. It checks both providers' network identities. RPC POSTs call only getNetwork/getTransaction; Horizon requests are GETs. No signing, funding or submission occurs. Capture alone is not verification; the checked-in test suite currently reads testnet.json. Review and verify a new capture before replacing that fixture.

The verifier checks the network-dependent envelope hash, successful per-operation result, operation source/destination, exact stroops, full asset identity, asset-contract ID, event economics, close time and ledger/transaction/operation-derived Horizon ID. Mirrored RPC event fields are compared with metadata rather than counted twice. Duplicate Horizon observations deduplicate; conflicting duplicates fail. Missing classic events and NOT_FOUND remain UNKNOWN. Every result has coverage_complete=false because looking up a transaction cannot prove complete account history.

Limits: the captured positive example is native XLM; a credit-asset fixture (`testnet-credit-asset.json`, CETES alphanum12) is now included. Fee-bump, muxed identities, non-V4 metadata, mixed/path/contract operations and complete history ingestion are not supported by this experiment. Event memo/muxed metadata is not substituted for envelope account identity. Providers are trusted observations, not cryptographic ledger-inclusion proofs. Synthetic metadata mutations test comparison logic, not full ledger state transition validation.

Sources checked 2026-09-12: [RPC getTransaction](https://developers.stellar.org/docs/data/apis/rpc/api-reference/methods/getTransaction), [Horizon migration](https://developers.stellar.org/docs/data/apis/migrate-from-horizon-to-rpc), [SDK release](https://github.com/stellar/js-stellar-sdk/tree/v17.0.1). Before a Go RPC ingestor, add real credit-asset and fee-bump fixtures, a scope/retention design and restart/coverage tests. Do not mark the full migration gate complete from this corpus alone.
