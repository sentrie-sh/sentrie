# 154-js-boundary-values

## Binaek Sarkar (binaek89@gmail.com)

### 14:59 - Start #154 on top of 191-cache-module-bindings
- Stacked on `191-cache-module-bindings` because the return conversion lives in `ModuleBinding.Call` and the internal globals are set in `loadJSInstance` and `AliasRuntime.Require`, all reshaped by the three branches below.
- Plan: (a) a JS to Sentrie conversion with `Date` as Unix seconds, string-keyed `Map` as a dict, and an error for unsupported values; document the table in the TypeScript module guide. (b) take `__require` and `__executionStartTimeUnix` out of the module-visible global scope.
- Keep `runtime/js/builtin_time.go` edits to how `now()` obtains the start time: #153 changes its fallback and units, and #111 turns its error returns into throws.
- _Model: claude-opus-5-5_
