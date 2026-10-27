# 154-js-boundary-values

## Binaek Sarkar (binaek89@gmail.com)

### 14:59 - Start #154 on top of 191-cache-module-bindings
- Stacked on `191-cache-module-bindings` because the return conversion lives in `ModuleBinding.Call` and the internal globals are set in `loadJSInstance` and `AliasRuntime.Require`, all reshaped by the three branches below.
- Plan: (a) a JS to Sentrie conversion with `Date` as Unix seconds, string-keyed `Map` as a dict, and an error for unsupported values; document the table in the TypeScript module guide. (b) take `__require` and `__executionStartTimeUnix` out of the module-visible global scope.
- Keep `runtime/js/builtin_time.go` edits to how `now()` obtains the start time: #153 changes its fallback and units, and #111 turns its error returns into throws.
- _Model: claude-opus-5-5_

### 15:01 - Paused at the plan limit (handoff)
- Work paused because the 5-hour plan window reached 93%; no implementation has started on this branch.
- Remaining on this chain: implement #154 here (plan above), then cut `142-eval-limits` off this branch with a stub journal first, and implement #142 per the issue.
- Open question for the maintainer before any push of 191: origin/191-cache-module-bindings still holds the old stub commit fe76acd, while local 191 was rebased onto 105's 09aaa0b (stub now aca1b7a). Either force-push with lease pinned to fe76acd, or merge fe76acd into local 191 (additive) and then merge 191 into this branch. Until then do not rebase 191 or this branch.
- _Model: claude-opus-5-5_

### 15:05 - Boundary conversion and hidden internals
- Resumed from the parent's handoff commit cc1f777.
- `runtime/js_values.go`: `jsToSentrie` replaces the kind allowlist in `Call`. Date becomes Unix seconds (int64, like `@sentrie/time.now()` and JWT `exp`), a string-keyed Map becomes a dict (iterated with `ForOf` under `vm.Try`), nested undefined stays undefined, and functions, symbols, BigInt, Set, non-string Map keys, RegExp, Error, invalid Date and cycles are errors naming the path (`unexpected return type <what> at return value.a[2]: ...`).
  - goja reports Map, Set, Promise and WeakMap as class `Object`, so they are told apart by export type. A WeakMap still reads as an empty dict; typed arrays and Go values wrapped by builtins keep the old allowlist and struct export.
  - Unit choice: seconds rather than RFC 3339, matching the issue's first option and the rest of `@sentrie/time`. Flagged for review.
- Internals: `__require` is now passed straight to the module factory and never set as a global. The execution start time moved to a Go-side `sync.Map` keyed by `weak.Pointer[goja.Runtime]` with a `runtime.AddCleanup` delete (`runtime/js/execution_start.go`), set by `js.SetExecutionStart` in `loadJSInstance`. `builtin_time.go` only changes how `now()` obtains the value (plus its header); its fallback is left for #153. `constants.ExecutionStartTimeUnixKey` and `constants/runtime.go` are removed as unused.
- Docs: the TypeScript module guide lives in `sentrie-sh/website` (`src/content/docs/reference/typescript_modules/index.md`), which this repo cannot change. The table is in the `jsToSentrie` doc comment and the `runtime.modules` node; the proposed website section is in the PR Review notes for the parent to file there.
- Tests: `TestJSToSentrieConversionTable`, `TestInternalGlobalsAreNotReachableFromModules` (`typeof` both is `undefined`; overwriting `__executionStartTimeUnix` does not move `now()`). `make test` and `make lint` pass.
- _Model: claude-opus-5-5_

### 22:08 - Review fixes (PR #215)
- Critical finding: `obj.Get`, `obj.Keys()` and `Get("length")` run module code (getters, Proxy traps, index accessors) from Go with no goja try frame, so a JS throw, the deadline interrupt or a stack overflow during conversion panicked out of `ModuleBinding.Call`. Only `ExecPolicy` recovers panics, so `exec` crashed and `serve` dropped the connection.
- Fix: `jsToSentrie` runs under a `recover` that turns `*goja.Exception`, `*goja.InterruptedError` and `*goja.StackOverflowError` (checked by `isGojaThrow`) into `converting <path>: ...` errors and re-panics anything else. The converter tracks the current path for that message. `convertMap` drops its own `vm.Try`, which could not catch uncatchable values anyway.
- A returned `Proxy` (class `Object`, export type `goja.Proxy`) is now an error rather than `{}` through `structs.Map`.
- Arrays: `maxReturnedListLen = 1 << 20` is checked before walking, and preallocation is capped at 1,024 instead of trusting `length` (`new Array(2**32-1)` asked Go for about 64 GiB). `jsToSentrie` now takes `ctx` and checks it every 1,024 converted values, because walking plain values never re-enters the VM and so is never interrupted.
- Trivial suggestions applied: the `Number` wrapper returns `ToNumber().Export()` directly, and the accepted kinds slice is hoisted to `acceptedGoKinds`.
- To stay contained for the parallel `224-weakmap-unsupported` branch, the functions touched are `jsToSentrie`, `jsConverter`, `step`, `isGojaThrow`, `convert`, `convertMap` and `exportGoValue`.
- Tests: the table gains a throwing getter, a Proxy, a `2**32-1` sparse array, a throwing index accessor and a throwing value inside a Map. `TestConversionStopsAtTheDeadline` checks that a looping getter on a returned object fails with `DeadlineExceeded` within 100ms rather than panicking.
- _Model: claude-opus-5-5_

### 22:26 - Do not cache a module load whose context ended (300-run flake check)
- `TestFailedModuleLoadIsRetried` with `-count=300` failed on this branch before the change: 1 of 300 on 191, 5 on 154, 1 on 142. On 191 and 154 these were the #194 panic ("context canceled at get" in `checkExports`). On 142, whose export check already clears the probe, it was `moduleLoads == 1`: the load sometimes finished before the pre-cancelled context's interrupt landed, so the module was cached anyway.
- Fix in `loadModule`: a load that succeeds after its own context has ended is treated as failed with `ctx.Err()`. It is not cached, and a waiter with a live context retries (the existing `isContextError` path). Whether such a load finished, and whether its probe kept an interrupt, was down to timing; now the outcome is deterministic. Same change as on `191-cache-module-bindings` (a separate, identical commit, because branches are not merged). It applies cleanly when GitHub restacks this branch, since both sides make the same edit.
- After the change: `TestFailedModuleLoadIsRetried` passed `-count=300` (0 failures). The lifecycle tests (`TestWaiterRetriesAfterLoaderIsCancelled`, `TestCloseDuringLoad`, `TestCloseDoesNotWaitForABusyExportCheck`, `TestConcurrentFirstUseLoadsTheModuleOnce`, `TestCancelledModuleCallDoesNotAffectTheNextCall`) passed `-race -count=30`.
- _Model: claude-opus-5-5_

### 13:58 - Restack onto main after #194, #197, #227 and #196 merged
- Rebased onto `main` (cfe249a). One textual conflict, in the relationship table of `docs/codebase-graph/runtime.js.builtins.md`: dropped the `LAYERED_ON [[constants]]` row #139-1 added, since this layer removes `ExecutionStartTimeUnixKey`.
- A semantic conflict git did not flag: #139-1 (merged as #197) added `builtin_clock.go` with its own `executionStartUnix`, returning `(int64, error)` and reading the `ExecutionStartTimeUnixKey` global that this layer deletes. With both, the package does not compile. I renamed it `executionStartSeconds`, made it wrap this layer's Go-side `executionStartUnix`, and kept `errNoExecutionStart`, so a VM with no pinned start still fails every JWT time check. `jwt`'s clock calls it.
- The jwt and builtin-errors test helpers now pin the clock with `SetExecutionStart` instead of setting the global. The graph's JWT lines now name the new helper, and the "module code can overwrite the clock" risk is recorded as closed by this layer.
- _Model: claude-opus-5-5_
