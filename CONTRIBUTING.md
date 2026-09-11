# Contributing

Use Go 1.22.2+ and run `go test ./...`, `go vet ./...`, `go build ./...`; CI also runs `go test -race ./...` with a C compiler. A sibling checkout is not required. Format changes with gofmt.

Choose a bounded task in docs/backlog.md. Open a PR explaining the concrete input/failure, behavior change, tests and limitations. For upstream-dependent work, cite the official schema/API and revision and use synthetic fixtures. Include negative cases, preserve exact amounts and UNKNOWN outcomes, and never hide a failed check. Keep PRs small enough for the maintainer to review code and semantics. Confirm maintainer availability before accepting time-sensitive assignments; no review-time guarantee is currently established.

No signing keys, customer exports, invented adoption or application compatibility. Named adapters require provenance before compatibility claims. Follow SECURITY.md for sensitive findings. Existing MIT licensing applies; preserve upstream notices.

Drips participation/approval is not established. Issues should arise from actual engineering needs; do not add Wave labels or points merely to create activity.
