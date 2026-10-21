# 142-eval-limits

## Binaek Sarkar (binaek89@gmail.com)

### 15:05 - Start #142 on top of 154-js-boundary-values
- Stacked on `154-js-boundary-values`: deadlines rely on the per-call VMs and the synchronous `js.InterruptOnDone` from #105, and every VM is built in `loadJSInstance` (#105, #154).
- Plan:
  - `WithEvalTimeout(d)` executor option, default 10s (`DefaultEvalTimeout`); `ExecRule` and `ExecPolicy` wrap `ctx` in `context.WithTimeout`. Nested `ExecRule` (cross-policy imports) inherits the earlier deadline. The existing interrupt then stops JS.
  - `--eval-timeout` on `exec` and `serve` (cling flags in `cmd/`), passed to `NewExecutor`.
  - `vm.SetMaxCallStackSize` in `js.NewAliasRuntime` (a documented constant, about 10000 frames).
  - `serve`: a per-request `context.WithTimeout(r.Context(), evalTimeout)` in `api/handle_decision.go`, with the server `WriteTimeout` kept above it so the error response can still be written.
  - Document what is bounded (time, call depth) and what is not (memory: recommend `GOMEMLIMIT`; Go-side Sentrie evaluation only checks the deadline where it calls into JS).
  - Tests: infinite-loop and deep-recursion modules fail with a timeout or stack error, with short test timeouts.
- _Model: claude-opus-5-5_

### 15:05 - Paused at the plan limit (handoff)
- Stopped before implementing: the 5-hour plan window read 97%, and #142 touches `runtime`, `runtime/js`, `cmd` and `api` and needs a full `make test` and `make lint`, which would push it past the 98% stop line.
- Nothing is implemented on this branch yet. Left to do is the whole plan in the entry above, plus its tests, graph updates (`runtime.executor`, `runtime.modules`, `runtime.js.alias_runtime`, `cmd`, `api.http`, `api.handle_decision`) and the PR description.
- Starting points: `runtime/executor.go` (`ExecRule`, `ExecPolicy`, options), `runtime/js/alias_runtime.go` (`NewAliasRuntime`), `cmd/exec.go` and `cmd/serve.go` (cling inputs), `api/http.go` (timeouts) and `api/handle_decision.go`. `TestCancelledModuleCallDoesNotAffectTheNextCall` in `runtime/module_isolation_test.go` already shows a looping module stopped by a 50ms deadline.
- _Model: claude-opus-5-5_
