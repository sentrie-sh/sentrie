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
