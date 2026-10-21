# 191-cache-module-bindings

## Binaek Sarkar (binaek89@gmail.com)

### 14:52 - Start #191 on top of 105-isolate-pooled-vms
- Stacked on `105-isolate-pooled-vms`: #191 says caching bindings for real must wait for #102 and #105, because it would turn both into cross-evaluation bugs. With both below this branch, caching is safe.
- Plan: hold loaded modules for the executor's lifetime in a plain map with singleflight on first load (no evicting cache), add `Executor.Close()` called from `exec`, `validate` and `serve`, test that the loader runs once across repeated `ExecRule` calls, and correct the `runtime.executor` node.
- Read `journal/140-perch-entry-count-caches.md`: #140 sized the perch caches by entry count and deliberately left `ttl = -1` alone.
- _Model: claude-opus-5-5_
