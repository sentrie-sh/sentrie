# 105-isolate-pooled-vms

## Binaek Sarkar (binaek89@gmail.com)

### 14:46 - Start #105 on top of 102-binding-export-key
- Stacked on `102-binding-export-key` because both reshape `runtime/modules.go` and `getModule` in `runtime/executor.go`; this branch builds on the shared `jsModule` introduced there.
- Plan (issue option 2): a fresh `goja.Runtime` per acquisition that reuses the module's compiled `goja.Program`, plus a fix for the interrupt watchdog racing with release.
- Reproduced with `runtime/module_isolation_test.go`: a 16-rule policy over a module-level counter under `ExecPolicy` yields 2, 3, 4 instead of 1, and two calls in one rule body share state. The issue's `record` export is renamed `track` in the tests because `record` is a Sentrie keyword.
- Baseline: a pooled `ModuleBinding.Call` costs about 1.0 to 1.1 us and 21 allocations (throwaway benchmark, not committed).
- _Model: claude-opus-5-5_

### 14:49 - Fresh runtime per acquisition, synchronous watchdog
- Replaced the `puddle` pool with a per-module design in `runtime/modules.go`:
  - `jsModule.acquire` loads the module into a **fresh** `goja.Runtime` for every call, through the registry's cached `goja.Program`, and drops the VM afterwards. Module-level state therefore lives for exactly one call, independent of ordering or concurrency.
  - A buffered channel of 10 slots per module keeps the old `MaxSize: 10` bound on live VMs; waiting honours `ctx`.
  - Dropped `puddle` rather than keeping it with destroy-after-use: puddle runs constructors on a context detached from the acquirer's cancellation (`newValueCancelCtx(ctx, baseAcquireCtx)`), so a module whose top level loops would never be interrupted. Loading inline in `Call` keeps the caller's context, which #142 needs. `go mod tidy` also drops `golang.org/x/sync`, which only puddle used.
  - Export checks at bind time read a **probe** instance loaded once per module and guarded by a mutex; it never serves a call, so no call can taint it. Reading an export may run an esbuild getter, hence the mutex rather than a lock-free read.
  - The execution-start global is set before any module code runs (in `loadJSInstance`), so top-level code and the call see the same pinned time. Previously the first load saw the wall clock.
- Semantics decision: isolation is per call, not per evaluation. Two calls in one rule body each see a freshly loaded module (`TestModuleStateDoesNotLeakBetweenCallsInOneRule`). This is the only granularity that is deterministic under `ExecPolicy`'s concurrency without threading a VM through the execution context. Flagged for review.
- Watchdog race: the old goroutine could `ClearInterrupt()` after `close(done)` while the VM was already back in the pool. New `js.InterruptOnDone` uses `context.AfterFunc`; its stop waits for an interrupt already in flight, so nothing touches the VM after stop returns. It no longer clears an interrupt that fired: the context is over, and a nested `require` must not clear the interrupt its parent frame relies on (the `alias_runtime` gotcha). `AliasRuntime.installInterrupt` now delegates to it. With fresh VMs a stale interrupt cannot reach another call anyway.
- Measurement (throwaway benchmark of `ModuleBinding.Call` on a small module that requires `@sentrie/hash`):
  - pooled reuse on `102-binding-export-key`: about 1.0 us, 808 B, 21 allocs per call.
  - first cut of fresh VMs: about 70 us per call. A CPU profile showed two thirds of it in `os.Lstat` from `confineToPackRoot` re-resolving the module path on every load.
  - added a resolution memo to `Registry.LoadRequire` (successful resolutions only, keyed by `fromDir` and spec): about 32 us, 55 KB, 749 allocs per call. What remains is goja runtime construction and GC.
  - The memo is safe for containment because a module's content is fixed by its first compile anyway (`sync.Once` in `programFor`).
  - `Registry.CompiledPrograms()` counts compiles; `TestFreshRuntimesReuseTheCompiledProgram` asserts exactly one across 20 rule evaluations.
- _Model: claude-opus-5-5_

### 22:05 - Review fix: clear the probe's interrupt after load (PR #194)
- Review finding (pr-194): if the first caller's context ends between the module factory returning and `Require`'s deferred stop, the interrupt fires and, by `InterruptOnDone`'s design, stays set. The load still succeeds and the probe is kept (from #191 on, for the executor's lifetime). Every later `checkExports` then runs an esbuild export getter on an interrupted VM and panics with `*goja.InterruptedError` from plain Go code, permanently.
- Fix: `newJSModule` calls `probe.rt.ClearInterrupt()` after a successful load. This is safe because the load's stop has returned and nothing else holds the VM. I chose clearing over failing the load because the module did load correctly.
- Test: `TestProbeInterruptIsClearedAfterLoad` hands `newJSModule` an already-interrupted instance with a getter export. It panics without the fix (checked by commenting the line out) and passes with it.
- Applied the trivial suggestion: inlined the one-caller `installInterrupt` wrapper into `Require`, and updated the `runtime.js.alias_runtime` node.
- _Model: claude-opus-5-5_
