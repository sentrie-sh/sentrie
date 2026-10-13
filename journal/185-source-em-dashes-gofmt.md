# 185-source-em-dashes-gofmt

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-13 16:20 - Em dashes and gofmt drift in source files
- Follow-up from #181's list, filed as #185 because it changes source files. Branched off `main` (`11070ea`).
- The work was done in a worktree named `source-hygiene`, then renamed once the issue existed, before any push.
- Em dashes: 15 Go files and both grammar files. Each was replaced with " - ", with no trailing space where the dash ended a comment line (that tripped `gofmt` on `index/policy_stmt.go` until fixed).
  - Two are in error messages in `runtime/eval_call.go`, after "is not permitted inside a derive".
  - Tests assert only that prefix (`index/derive_test.go`, `index/derive_coverage_test.go`, `runtime/eval_call_test.go`), and no graph node quotes the full text.
- gofmt: `gofmt -w` on the 9 files `gofmt -l .` listed. `gofmt -l .` is now empty.
- Headers on every touched file (`license-headers`, rewrite on touch):
  - full legacy Apache blocks and the short `SPDX-License-Identifier` plus `Copyright` form both became the two-line SPDX header for 2026;
  - existing SPDX headers were already 2026;
  - the grammar files had no header, and now carry one in `/* */` comments. No code reads the grammar files.
- Verified, with `GOWORK=off`: `go build ./...`, `go vet ./...`, `go test ./... -count=1` (19 packages ok) and `make docs-graph` all pass.
