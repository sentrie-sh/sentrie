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
