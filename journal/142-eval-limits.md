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

### 17:27 - Resume after the plan window reset: deadline, stack limit, request deadline
- `runtime/executor.go`: `DefaultEvalTimeout = 10s` and `WithEvalTimeout`. `ExecRule` and `ExecPolicy` go through `withEvalDeadline`, which wraps `ctx` in `context.WithTimeout`. It adds `evaluation exceeded its <d> deadline` to the error only when this deadline, and not the caller's, ended the evaluation (`parent.Err() == nil`), so a nested `ExecRule` does not wrap twice. `d <= 0` disables it.
  - 10s default: long enough for any reasonable decision, short enough that a runaway module frees its core quickly, and it matches the order of the old 30s HTTP timeouts.
- `runtime/js/alias_runtime.go`: `MaxCallStackSize = 10_000`, set in `NewAliasRuntime`, so every VM (probe and per call) has it. goja's default is effectively unlimited. 10,000 frames is well above real module code and close to what browsers allow.
- `cmd`: `--eval-timeout` on `exec` and `serve` (`cmd/eval_timeout.go`). It is a string flag because cling hydrates `time.Duration` as a plain int, so `parseEvalTimeout` parses it; `0` disables, negatives are rejected.
- `api`: `NewHTTPAPI(exec, WithRequestTimeout(d))`, defaulting to `DefaultEvalTimeout`. `handleDecision` puts that deadline on the request context, because the 30s `WriteTimeout` only closed the connection and never cancelled the handler. `WriteTimeout` is now `max(30s, requestTimeout + 5s)`, so the error can still be written. A deadline error maps to 503 (choice flagged for review; 504 implies an upstream).
- Bounded vs not (README "Evaluation Limits", `runtime.executor` gotcha): time and JS call depth are bounded. Memory is not, so the README recommends `GOMEMLIMIT` plus a container limit. Go-side Sentrie expression evaluation never checks `ctx`, so it stops only at its next module call; noted as a known gap, not fixed here.
- Tests: an infinite loop fails with `DeadlineExceeded` in about 100ms; deep recursion fails with `*goja.StackOverflowError` under a 5s deadline; `ExecPolicy` is bounded as a whole; a shorter caller deadline does not get the extra prefix; every runtime has the stack limit; an HTTP request whose executor deadline is disabled still returns 503 at its 100ms request deadline; flag parsing. Every blocking test runs under a 5s guard so a regression fails instead of hanging.
- `make test`, `make lint` and `-race` on `api`, `cmd` and the runtime limit tests pass.
- Commits could not be signed earlier (`agent refused operation`); see the next entry.
- _Model: claude-opus-5-5_

### 17:28 - Commits pending signing (handoff)
- The #142 work is complete, and `make test`, `make lint` and the race runs pass, but nothing after 4d789d4 is committed: commit signing fails with `Couldn't sign message (signer): agent refused operation?`, and only the maintainer can unlock the signer. Signing was not disabled.
- Intended commit split, in order (`PR_DESCRIPTION_142-eval-limits.md` stays uncommitted):
  1. "Bound each evaluation with a deadline and every JS runtime with a call-depth limit": `runtime/executor.go`, `runtime/js/alias_runtime.go`, `runtime/module_limits_test.go`.
  2. "Add --eval-timeout to exec and serve and a request deadline to the HTTP API": `cmd/eval_timeout.go`, `cmd/eval_timeout_test.go`, `cmd/exec.go`, `cmd/serve.go`, `api/http.go`, `api/handle_decision.go`, `api/request_timeout_test.go`.
  3. "Document evaluation limits in the README and the graph": `README.md`, `docs/codebase-graph/{runtime.executor,runtime.js.alias_runtime,cmd,api,api.http,api.handle_decision}.md`, `journal/142-eval-limits.md`.
- _Model: claude-opus-5-5_

### 22:10 - Review fixes (PR #219)
- Finding: three Go-side touches ran module code outside any interrupt, so the deadline did not bound them. They were `newJSInstance`'s `exports.Keys()` (a Proxy `ownKeys` trap), `checkExports`'s `probe.exports.Get` (an export getter, which also held `probeMu` and so blocked `Close`), and `lookupExport`, which ran before `Call` installed its interrupt.
- Fix in `runtime/modules.go`: new `runGuardedJS(ctx, vm, f)` (`InterruptOnDone` plus `recoverUncaughtJS`, which turns goja throws, interrupts and stack overflows into errors).
  - `loadJSInstance` captures the export keys under it.
  - `checkExports` takes `ctx`, runs under it, and always clears the probe's interrupt afterwards (`probeMu` is held and the guard's stop has returned).
  - `Call` installs `InterruptOnDone` before `lookupExport` and defers `recoverUncaughtJS`, so the lookup and the return conversion are covered too.
- Duplication, by request: `recoverUncaughtJS` duplicates `isGojaThrow` plus the recover that `154-js-boundary-values` added to `jsToSentrie` (PR #215 fix), because this branch was cut from 154 before that fix and branches must not be merged. The names differ so the restack after #215 merges does not clash. Once both are in `main`, `jsToSentrie` can drop its own recover in favour of `Call`'s (follow-up).
- `serve` keeps one deadline, the executor's. `handleDecision` no longer adds its own `context.WithTimeout` (it always fired first, so the "evaluation exceeded its ... deadline" message never reached clients, and it also charged body decoding to the budget). `WithRequestTimeout` now only sizes the write timeout. With `--eval-timeout 0` the 30s write timeout still caps a response; this is documented and pinned by a test.
- The deadline prefix is now added only when the error itself wraps `context.DeadlineExceeded`, so a fact-validation error is not labelled a timeout. The `DefaultEvalTimeout` comment now says it bounds JavaScript only (#225).
- README and the `runtime.executor`, `api.http`, `api.handle_decision` and `cmd` nodes are corrected to say where the deadline applies.
- Tests: `TestDeadlineBoundsGoSideJSTouches` (a looping getter export and a looping Proxy `ownKeys`, both fail with `DeadlineExceeded` within the 5s guard, and `Close` returns afterwards). The API test now runs serve's configuration and asserts the deadline message. `TestWriteTimeoutCoversTheRequestDeadline` pins the `0` case.
- _Model: claude-opus-5-5_

### 22:26 - Do not cache a module load whose context ended (300-run flake check)
- `TestFailedModuleLoadIsRetried` with `-count=300` failed on this branch before the change: 1 of 300 on 191, 5 on 154, 1 on 142. On 191 and 154 these were the #194 panic ("context canceled at get" in `checkExports`). On 142, whose export check already clears the probe, it was `moduleLoads == 1`: the load sometimes finished before the pre-cancelled context's interrupt landed, so the module was cached anyway.
- Fix in `loadModule`: a load that succeeds after its own context has ended is treated as failed with `ctx.Err()`. It is not cached, and a waiter with a live context retries (the existing `isContextError` path). Whether such a load finished, and whether its probe kept an interrupt, was down to timing; now the outcome is deterministic. Same change as on `191-cache-module-bindings` (a separate, identical commit, because branches are not merged). It applies cleanly when GitHub restacks this branch, since both sides make the same edit.
- After the change: `TestFailedModuleLoadIsRetried` passed `-count=300` (0 failures). The lifecycle tests (`TestWaiterRetriesAfterLoaderIsCancelled`, `TestCloseDuringLoad`, `TestCloseDoesNotWaitForABusyExportCheck`, `TestConcurrentFirstUseLoadsTheModuleOnce`, `TestCancelledModuleCallDoesNotAffectTheNextCall`) passed `-race -count=30`.
- _Model: claude-opus-5-5_
