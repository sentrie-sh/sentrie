# local-tooling

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-13 16:16 - Makefile targets, GOWORK and the sandbox template
- Follow-up from `agentrules-convention` (#181): the Makefile had no `test`, `lint` or `help` target and didn't set `GOWORK=off`, `make build` stamped the wall-clock date and always `treeState=dirty`, and the `sandbox.go` template failed `go vet`.
- Branched off `agentrules-convention`, because this updates rule files #181 adds. Worktree: `worktrees/local-tooling`. Repo tooling, so no issue and no title prefix (`pr-conventions`).
- Makefile:
  - `export GOWORK := off`, so every target ignores the ambient `~/go.work` that resolves worktree packages to the primary checkout.
  - New `test` (`go test ./...`), `lint` (`go vet ./...` then `$(GOLANGCI_LINT) run --timeout=5m`) and `help` (greps `## ` target comments) targets, plus `.PHONY`.
  - `make build`: `date` is the commit date in UTC (`TZ=UTC git log -1 --format=%cd --date=format-local:...`), matching GoReleaser's `CommitDate`. `treeState` comes from `git status --porcelain`.
  - `build` is still the default target, and the output directory is unchanged.
- Verified: `make help` lists all targets. `make build OUTPUT_DIR=<scratch>` produced a binary whose `--version` shows the commit date, `dirty` (the Makefile was modified at the time) and `make`. `make test` passed, 19 packages ok. `make docs-graph` passed.
- Not verified: `make lint` against golangci-lint.
  - The locally installed v2.11.4 was built with Go 1.26 and panics on the local Go 1.27.1 standard library ("file requires newer Go version go1.27").
  - CI's v2.4.0 runs with Go 1.25 and is unaffected.
  - Noted in `build-and-test`, with the `GOLANGCI_LINT` override.
- Corrected a false claim from #181: `build-and-test` said `gofmt -l .` must print nothing, but 9 existing files are not gofmt-clean, and CI doesn't check formatting. The rule now says to gofmt the files you touch and never reformat others.
- `sandbox.go`:
  - Template gets `_ = ctx // remove once your test code uses ctx`, so it compiles and passes `go vet -tags sandbox sandbox.go`.
  - The run comment gains `GOWORK=off`.
  - Legacy Apache header rewritten to the two-line SPDX form for 2026 (`license-headers`, rewrite on touch).
  - `sandbox.md` shows the same template, and drops the "known gotcha".
- Rules: `build-and-test` (targets, stamps, the GOWORK export, golangci-lint caveat, formatting), `stacked-prs` and `token-efficiency` (now use `make test` and `make lint`), and the `AGENTS.md` entry.
