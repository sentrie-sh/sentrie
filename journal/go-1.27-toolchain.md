# go-1.27-toolchain

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-09 19:05 - Pick up the Go 1.27 toolchain bump
- Asked to move Sentrie to Go 1.27. Local Go is 1.27.1; `go.mod` and CI pin 1.25, and the local golangci-lint (built with go1.26) panics on Go 1.27 sources, so `make lint` only ran with `GOTOOLCHAIN=go1.25.0`.
- Scope: `go.mod`, CI and knowledge-graph workflows, the golangci-lint pin (support for go1.27 landed in v2.13.0; latest is v2.15.0), and the rules and graph node that name the Go version.
- No issue: a toolchain and CI change is repo tooling, per the PR conventions.
- _Model: claude-opus-5-5, tokens: ~4k_
