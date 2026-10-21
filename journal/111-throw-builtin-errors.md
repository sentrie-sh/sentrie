# 111-throw-builtin-errors

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:38 - Plan the throw conversion for #111
- Confirmed the issue on `main@b35fe4b`: 141 `return vm.NewGoError(` sites across 12 `runtime/js/builtin_*.go` files, none thrown. `builtin_crypto.go`, `builtin_math.go` and `builtin_string.go` have no sites, which keeps this branch out of the files #138 and #158 own.
- Helper shape: `throwErr(vm, err) goja.Value` panics with `vm.NewGoError(err)`. It is typed to return a `goja.Value` only so every call site keeps its `return` and the control flow (and the compiler's view of it) stays unchanged: `return throwErr(vm, err)`.
- Conventions, applied to every function:
  - Argument-count and argument-type errors throw.
  - Predicates (`*IsValid`, `is*`, `jwt.verify`) answer `false` for input they cannot parse; they are questions about the input, so "not valid" is an answer.
  - A genuine "no result" is `null` (`net.parseIP`, `regex.find`, `semver.prerelease`/`metadata`); `map_get` keeps `undefined`/default for a missing key.
  - Everything else throws.
- Behaviour changes beyond the mechanical swap, made to honour "type errors throw":
  - `collection.list_includes`, `list_indexOf`, `list_lastIndexOf` threw nothing on a non-array and answered `false`/`-1`; they now throw. `map_has` and `map_get` on a non-dict now throw instead of answering `false`/default.
  - `net.cidrMerge` silently dropped non-string array items; it now throws.
  - `jwt.verify` answered `false` for an unsupported expected `algorithm` argument; that is a caller error, so it now throws. A malformed or mismatched token still answers `false`.
- Kept out: `@sentrie/time.now()` keeps its wall-clock fallback here; #153 removes it on a later branch. `BuiltinBase64Go` is never registered with the executor, so it gets the mechanical swap only.
- The module guide is not in this repo: `README.md` links to sentrie.sh, whose source is the separate `website` repo (`src/content/docs/reference/typescript_modules/`). This branch documents the conventions in the graph node and the `.d.ts` declarations; the website needs its own change.
- _Model: claude-opus-5-5_
