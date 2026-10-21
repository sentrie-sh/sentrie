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
