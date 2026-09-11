# Security and trust

LedgerParity reads exports and provider data; it never needs secret keys or signs transactions. Treat reports as sensitive: they contain account identifiers, amounts and application references. Keep customer exports and reports out of public commits. JSON files use private creation permissions where the OS supports them; Windows ACLs and existing files need operator management.

Horizon is a trusted data provider, not a locally verified ledger proof. Configure an endpoint you trust. Do not expose the CLI as a service accepting arbitrary URLs or files. Offline fixtures are asserted data, not on-chain verification. Exact account equality does not validate StrKey checksums. Unsupported payment types require separate mapping and verification.

For a sensitive bug use the repository's GitHub Security > Report a vulnerability option if available. If private reporting is unavailable, open a minimal issue asking for a private contact without disclosing secrets, customer data or exploit details. No dedicated security mailbox or response-time promise has been established.
