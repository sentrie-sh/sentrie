# 140-perch-entry-count-caches

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 01:16 - Investigate #140 and measure the baseline
- Cause confirmed as described in #140:
  - `NewExecutor` built `perch.New[*ModuleBinding](100 << 20)` and `perch.New[any](10 << 20)`, then called `Reserve()` on both.
  - perch v0.0.3 divided the byte budget by `sizeof(T)`, so these became about 13.1M and 655k slots, each allocated eagerly as a full `entry[T]`.
- Baseline on `main` with `sentrie validate --pack-location ./example_pack`: about 4.8 GB peak RSS and about 0.65 s per warm run.
- Considered replacing perch with ristretto. Rejected:
  - Module bindings own `puddle` VM pools. Ristretto's admission policy and asynchronous `Set` can drop or delay a binding, and evicting one would need its pool closed.
  - Memoization would need `singleflight` on top to keep perch's behaviour, plus a `Close()` for ristretto's goroutines. Its cost model gains nothing when every entry costs 1.
  - perch is maintained alongside Sentrie, so fixing its sizing upstream is the smaller change.
- Found that the module-binding cache never caches:
  - `getModuleBinding` passes `ttl = -1`, which perch treats as "do not cache".
  - Every `ExecRule` rebuilds a VM pool per `use` and never closes the old one.
  - The `runtime.executor` graph node claimed the opposite.
- Kept out of this branch: caching bindings for real would turn #102 (export restriction leak) and #105 (VM state leak) into cross-evaluation bugs.

### 2026-10-21 01:18 - File #191 for the module-binding cache
- Filed sentrie-sh/sentrie#191: module bindings are never cached. It covers the impact, the interaction with #102 and #105, and a suggested fix: a lifetime map keyed by path plus exports, singleflight on first load, and an executor `Close()`.
- #140 deliberately leaves this behaviour unchanged.

### 2026-10-21 14:12 - Adopt perch v0.1.0 and size the caches by entry count
- perch v0.1.0 (binaek/perch#6):
  - `New[T](capacity)` now takes an entry count, and entries are allocated lazily.
  - `Reserve()` is now an optional warm-up.
- `runtime/executor.go`:
  - Module-binding cache: `moduleBindingCacheEntries = 1024`.
  - Memoization cache: `defaultCallMemoizeCacheEntries = 10_000`.
  - Dropped both `Reserve()` calls.
  - `WithCallMemoizeCacheSize` now takes an entry count, not megabytes. Its only callers are tests.
- Tests:
  - Hand-built executors in `executor_test.go`, `eval_call_test.go` and `eval_chain_gap_wave2_test.go` used `1 << 20`, which now means a million entries, and `testExecutorForModuleBinding` reserved them eagerly. Switched them to 16 entries and dropped the now-optional `Reserve()` calls.
  - Added `TestNewExecutorDoesNotPreallocateCaches`, which bounds `NewExecutor` at 1 MB of `TotalAlloc`. It measured about 6 KB, compared with gigabytes before.
  - `TestWithCallMemoizeCacheSizeAndToTrinaryHelpers` now asserts `Cap() == 2`.
- After: about 18.6 MB peak RSS and about 0.01 s per warm run for the same `validate` command.
- Graph:
  - `runtime`, `runtime.executor`, `runtime.eval_call` and `ext.dependencies` now describe the entry-count bounds and perch v0.1.0.
  - `runtime.executor` now documents that bindings are never retained (#191) instead of claiming a cache hit after first use.

### 2026-10-21 14:15 - Move the work into a worktree and verify
- The work started in the primary checkout on the merged `182-fix-execpolicy-panic-tests` branch, before `main` adopted `.agentrules/` (#181).
  - Moved it onto `140-perch-entry-count-caches`, off `main` at `190360b`, in `worktrees/140-perch-entry-count-caches`.
  - Carried the diff over as a patch.
  - Resolved a `.gitignore` conflict in favour of `main`, which already ignores `worktrees/` and `PR_DESCRIPTION_*.md`.
- Rewrote the legacy Apache headers of `runtime/executor.go`, `runtime/executor_test.go` and `runtime/eval_chain_gap_wave2_test.go` to the two-line SPDX form.
- The first `go get` ran inside the ambient `~/go.work`. Under `GOWORK=off`, `go mod tidy` also dropped the stale `perch v0.0.3` lines from `go.sum`.
- Verified with `GOWORK=off`:
  - `go build ./...`, `go vet` on `runtime`, `cmd` and `api`.
  - `go test -race` on `runtime`, `api` and `cmd`.
  - `make docs-graph`.
  - All pass.
