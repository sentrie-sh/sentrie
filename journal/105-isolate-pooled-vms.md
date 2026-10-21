# 105-isolate-pooled-vms

## Binaek Sarkar (binaek89@gmail.com)

### 14:46 - Start #105 on top of 102-binding-export-key
- Stacked on `102-binding-export-key` because both reshape `runtime/modules.go` and `getModule` in `runtime/executor.go`; this branch builds on the shared `jsModule` introduced there.
- Plan (issue option 2): a fresh `goja.Runtime` per acquisition that reuses the module's compiled `goja.Program`, plus a fix for the interrupt watchdog racing with release.
- Reproduced with `runtime/module_isolation_test.go`: a 16-rule policy over a module-level counter under `ExecPolicy` yields 2, 3, 4 instead of 1, and two calls in one rule body share state. The issue's `record` export is renamed `track` in the tests because `record` is a Sentrie keyword.
- Baseline: a pooled `ModuleBinding.Call` costs about 1.0 to 1.1 us and 21 allocations (throwaway benchmark, not committed).
- _Model: claude-opus-5-5_
