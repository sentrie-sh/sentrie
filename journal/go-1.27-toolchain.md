# go-1.27-toolchain

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-09 19:05 - Pick up the Go 1.27 toolchain bump
- Asked to move Sentrie to Go 1.27. Local Go is 1.27.1; `go.mod` and CI pin 1.25, and the local golangci-lint (built with go1.26) panics on Go 1.27 sources, so `make lint` only ran with `GOTOOLCHAIN=go1.25.0`.
- Scope: `go.mod`, CI and knowledge-graph workflows, the golangci-lint pin (support for go1.27 landed in v2.13.0; latest is v2.15.0), and the rules and graph node that name the Go version.
- No issue: a toolchain and CI change is repo tooling, per the PR conventions.
- _Model: claude-opus-5-5, tokens: ~4k_

### 2026-11-09 19:20 - Bump to Go 1.27 and golangci-lint v2.15.0
- `go.mod` is `go 1.27.0`; `go mod tidy` changed nothing else. `ci.yml` and `docs-graph.yml` set up Go 1.27 (the knowledge-graph cache key follows). `deps-policy-check.yml` and `release.yml` already read `go.mod` through `go-version-file`, so they follow without an edit.
- golangci-lint goes from v2.4.0 to v2.15.0 (go1.27 support landed in v2.13.0). The Makefile help text, `build-and-test.md` and `github-actions.md` name the new pin.
- The newer govet `inline` analyzer flagged `reflect.Ptr` in `trinary/tristate.go`; it is `reflect.Pointer` now (same constant). The file's legacy header was rewritten to the two-line SPDX form, per the header rule.
- Gate on go1.27.1: `make build`, `make docs-graph`, `make test`, and `make lint` with a golangci-lint v2.15.0 built with go1.27.1 (0 issues).
- _Model: claude-opus-5-5, tokens: ~12k_
