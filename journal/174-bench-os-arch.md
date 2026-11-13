# 174-bench-os-arch

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-13 15:55 - Pick up the bench os/arch rename
- Review of the v0.0.5 website release PR (sentrie-sh/website#73, thread on `cli-reference/bench.md` line 144) asked for `os` and `arch` in place of `goos` and `goarch` in the bench output. The user approved fixing it in the engine first, so the docs can follow.
- Scope: `sentrie bench` prints `os: <os>` and `arch: <arch>` configuration lines instead of `goos:` / `goarch:`. Result lines and their `BenchmarkExec/...` names are unchanged.
- Supersedes the `174-1-bench` decision to print `goos:` and `goarch:` lines. `benchstat` treats any `key: value` line before the results as configuration, so comparison still works.
- Branch `174-bench-os-arch` off `origin/main` at 74ead0f. Updates #174 (the issue stays open for its later slices).

### 16:10 - Rename the configuration lines
- `cmd/bench.go`: the line before the results loop prints `os: %s\narch: %s\n` (still `runtime.GOOS` / `runtime.GOARCH`, so the values are unchanged).
- `cmd/bench_test.go`: `TestBenchCmdFixedIterationsAndLineFormat` now checks that the output starts with exactly the `os:` and `arch:` lines and contains no `goos:` or `goarch:`. A plain `Contains("os: ...")` would also match `goos: ...`, hence the prefix check plus the two `NotContains`.
- Proved the test fails with the old format restored (prefix and both `NotContains` fail), then passes with the new one.
- `docs/codebase-graph/cmd.md`: the bench Behavior bullet names the new lines.
- `GOWORK=off go test -race ./cmd/ ./internal/docsgraph/`, `go vet ./cmd/`, `gofmt -l cmd` and `golangci-lint run ./cmd/...` are clean. No other file in the repo mentions the old lines except `.goreleaser.yml`, where `goos`/`goarch` are GoReleaser build keys, not bench output.
