# Build and Test

Everything builds from the repository root with no per-package tooling knowledge: a clean checkout must build and test with root commands alone. `make` is the entry point for anything more than a plain `go` command.

## Prerequisites

- **Go:** the version in `go.mod` (`go 1.25.0`). With `GOTOOLCHAIN=auto`, the default, an older local Go fetches it automatically.
- **GNU Make.**
- **golangci-lint** for linting. CI pins `v2.4.0` in `ci.yml`, with no repo config file, so the default linters apply.
- **`GOWORK=off`** on every `go` invocation when a `go.work` file sits above the checkout (for example in your home directory). An ambient workspace that lists the primary checkout joins the module graph, and inside `worktrees/<branch>` it makes `go test ./...` fail with "main module does not contain package", or worse, build the primary checkout's packages instead of the worktree's. The Makefile does not set it yet, so pass it through: `GOWORK=off make docs-graph`.
- Nothing else. The JavaScript and TypeScript support (`runtime/js`) runs on Go libraries (`goja`, `esbuild`), so no Node toolchain is needed.

## Commands

Run these from the root.

- `make build` - build the `sentrie` binary into `$(OUTPUT_DIR)` (default `~/.local/bin`), with `version` from the latest git tag (or `0.0.1-dev`), `commit`, `date` and `builtBy=make` injected through `-ldflags`. For a binary that stays out of your `PATH`, use `make build OUTPUT_DIR=./dist` (`dist/` is ignored) or `go build ./...`.
- `make docs-graph` - validate the codebase knowledge graph (`go test -count=1 ./internal/docsgraph/`).
- `make gen` - run `go generate ./...`. Today that regenerates `ast/typeref_constraint_args_gen.go` from the constraint checkers in `constraints/`. The generator writes the license header from `git config user.name` and `user.email`, so it needs a git identity.
- `make clean-git` - on `main` only: `git fetch -p`, then delete local branches whose upstream is gone. It is an origin touch ([Batch origin access](origin-access.md)).
- **Tests:** `go test ./...` runs everything, including `lang_test` (every `.sentrie` fixture must parse, index and validate) and the graph check. Prefer the narrowest scope that proves the change ([Token efficiency](token-efficiency.md)). CI runs non-`runtime` packages with `-race` and `./runtime/...` without it, because the race detector runs out of memory on the 7 GiB runners. Locally, `go test -race` on the package you changed is enough.
- **Lint and format:** `go vet ./...`, `gofmt -l .` (must print nothing) and `golangci-lint run`.

## Release builds

Releases are built by GoReleaser (`.goreleaser.yml`) from a pushed `v*.*.*` tag, via `release.yml`. The release build sets `CGO_ENABLED=0` and `-trimpath`, stamps `main.date` with the **commit** date and `main.treeState` from git, and signs artifacts with cosign. `make build` differs: it stamps the wall-clock build date and always marks the tree `dirty`. Nothing checks that either build is reproducible.

**Why:** A root-only build means a new contributor or agent needs no tribal knowledge, and CI and local runs use the same commands.

**How to apply:** Add any new build, generate, or check step as a Makefile target, not as a script someone has to know about. Before pushing, run `go build ./...`, `go test ./...` and `golangci-lint run` on the change.

Related: [Go project layout](go-project-layout.md), [Token efficiency](token-efficiency.md), [GitHub Actions](github-actions.md)
