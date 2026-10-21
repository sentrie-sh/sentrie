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

### 2026-10-21 22:26 - Record the test work behind the first commits
- Written after the fact, as review asked: the first entry covered only the plan.
- Added `TestBuiltinFailuresThrow` (per module: a failure is caught by `try`/`catch` with the expected message, and surfaces as a `*goja.Exception` when uncaught), `TestBuiltinPredicatesAndNoResult`, and the policy-level `TestJSBuiltinFailuresFailClosed` for the issue's reproduction.
- Checked the first, line-based guard by planting `return vm.NewGoError(nil)` in a scratch file; it failed at that line.
- _Model: claude-opus-5-5_

### 2026-10-21 22:26 - Review fixes for PR #195
- The guard matched only the spelling `return vm.NewGoError(`. Review showed `e := vm.NewGoError(err); return e`, `return (vm.NewGoError(err))`, `vm.ToValue(vm.NewGoError(err))` and `vm.ToValue(errors.New(...))` all got through. I inverted it: `errorObjectViolations` walks the AST and allows a `NewGoError`/`NewTypeError` call only as the argument of `panic`. It also flags `ToValue` wrapping `errors.New`, `errors.Join`, `fmt.Errorf` or `err`. Today it allows exactly `builtin_errors.go` (`throwErr`) and `alias_runtime.go:125`.
- Moved the guard to `builtin_errors_guard_test.go` and added `TestErrorObjectGuardCatchesEverySpelling`, which runs the matcher over in-memory snippets: nine bad spellings are each flagged once, and four throwing or plain forms are not.
- `semver.satisfies("1.0.0", "~~bad")` checked `Contains(got, "")`, which always passes; it now pins `improper constraint: ~~bad`.
- _Model: claude-opus-5-5_

### 2026-10-21 22:33 - Review fixes committed
- Signing worked again on a later attempt, so the review fixes above are committed as 8edeea6 ("Address review findings"). That supersedes the "left uncommitted" lines in the entry above. It is one commit rather than one per finding, because splitting it now would need an amend, and the review-fix rules forbid that.
- _Model: claude-opus-5-5_
