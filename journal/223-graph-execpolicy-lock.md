## Binaek Sarkar (binaek89@gmail.com)

### 21:55 - Pick up #223
- Branch cut from `main`. Docs only.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 21:59 - Drop the stale ExecPolicy clause from the runtime node
- Checked `runtime/executor.go`: the `recover()` branch in `ExecPolicy` takes `theLock` before joining the panic into `compositeErr`, matching `runtime.executor` (fixed in #182 / #183). The clause in `runtime.md` was stale, so it is removed.
- Checked the remaining claim against `runtime/exec_ctx.go`: `AttachedChildContext` copies `ec.modules` by reference while each context locks its own `rwmu`, so it still holds. Kept it, and added why it is latent: the only `BindModule` caller is `bindUses`, which runs on the root context in `ExecRule` before evaluation starts. This matches the `runtime.exec_ctx` gotcha.
- Pointed the bullet at `[[runtime.exec_ctx]]` only, since the executor no longer has a concurrency defect to cite.
- `make docs-graph` passes.
- _Model: claude-opus-5-5_
