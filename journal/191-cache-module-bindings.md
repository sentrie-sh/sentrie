# 191-cache-module-bindings

## Binaek Sarkar (binaek89@gmail.com)

### 14:52 - Start #191 on top of 105-isolate-pooled-vms
- Stacked on `105-isolate-pooled-vms`: #191 says caching bindings for real must wait for #102 and #105, because it would turn both into cross-evaluation bugs. With both below this branch, caching is safe.
- Plan: hold loaded modules for the executor's lifetime in a plain map with singleflight on first load (no evicting cache), add `Executor.Close()` called from `exec`, `validate` and `serve`, test that the loader runs once across repeated `ExecRule` calls, and correct the `runtime.executor` node.
- Read `journal/140-perch-entry-count-caches.md`: #140 sized the perch caches by entry count and deliberately left `ttl = -1` alone.
- _Model: claude-opus-5-5_

### 14:58 - Lifetime module map, singleflight, Executor.Close
- `runtime/executor.go`: replaced the `moduleBindingPerch` perch cache (`ttl = -1`, so it never cached) with a plain `map[string]*moduleLoad` under a mutex, held for the executor's lifetime. Dropped `moduleBindingCacheEntries`; perch now backs only call memoization.
  - Singleflight without `golang.org/x/sync` (dropped in #105's branch): the first caller inserts a `moduleLoad` with a `done` channel and loads; others wait on `done` or their own `ctx`.
  - Failed loads are removed from the map, so a later evaluation retries (perch also never cached errors).
  - A waiter whose own context is live retries when the load failed with a context error, so in `serve` one request's deadline cannot fail a concurrent request that joined its load.
  - The map is initialised lazily so hand-built test executors keep working.
  - `moduleLoads` (atomic) counts loads for tests and diagnostics.
- `Executor.Close()`: marks the executor closed, closes every loaded `jsModule` (drops the probe, later `acquire` fails with `errModuleClosed`), and clears the map. Later loads fail with `ErrExecutorClosed`. A load that finishes after Close closes its own module.
  - It does not wait for in-flight calls: since #105 each call holds its own VM and nothing it uses is freed by Close, so waiting would only let one hung call block process exit.
- `cmd/exec.go` and `cmd/serve.go` defer `Close`; `cmd/validate.go` returns it. Rewrote the legacy headers of those three files.
- Data flow check for races: `moduleLoad.module`/`err` are written under `modulesMu` before `close(done)`, and read by waiters only after `<-done` (channel close is the happens-before edge) or by `Close` under `modulesMu` after `done`. `jsModule.closed` and `probe` are only touched under `probeMu`. `go test -race` on `./runtime/ ./cmd/ ./api/` passes.
- Tests (`runtime/module_cache_test.go`): loader runs once across ten `ExecRule` calls with two `use`s of one module; 16 concurrent first uses load once; a load cancelled by its caller is retried by the next; Close is idempotent and rejects later calls. `TestGetModuleBindingCachesBindings` now asserts the same module pointer and one load.
- The #102 and #105 sequential regression tests (`TestUseExportRestrictionIsPerUseStatement`, `TestModuleStateDoesNotLeakAcrossEvaluations`) are now meaningful: with caching on they still pass, which is the point of landing #102 and #105 first.
- Graph: `runtime.executor` no longer claims "never actually cached" and documents the map, singleflight, `Close`, and the no-eviction gotcha; `runtime`, `cmd` and `ext.dependencies` updated. This corrects the node text #140 (`140-perch-entry-count-caches`) wrote about bindings never being retained.
- Fixed a stale catalog line in `overview.md` on `105-isolate-pooled-vms` and restacked this branch onto it.
- _Model: claude-opus-5-5_
