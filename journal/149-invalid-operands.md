## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #149
- Invalid operands give silent results: `[10,20,30]["foo"]` is 10, `[10,20,30][1.7]` is 20, `count(5)` is undefined, and `"a" + 1` is `"a1"`.
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A string or non-integer list index is an error.
  - A missing dict key or an out-of-range integer index stays undefined.
  - `count` on anything but a list, dict or string is an error; `count(undefined)` stays undefined.
  - Only `count` is fixed here; a follow-up issue is filed to audit the other builtins for the same problem.
  - `+` with a string follows JavaScript for scalars: string + number or bool concatenates in either order (`"a" + 1` is `"a1"`); undefined propagates as undefined; a list or dict with a string is an error.
- Docs PR: sentrie-sh/website#78 (branch `149-invalid-operands` on `next`). Sentrie PR: #318.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 15:54 - Implement strict list indexes, strict count and JavaScript-style concatenation
- Re-reproduced all four symptoms on the branch with a scratch pack and `sentrie exec`, plus one more: an undefined list index (`[10,20,30][x.missing]`) also read element zero.
- `runtime/eval_access.go`: a new `listIndex` helper serves both the boxed list and the boundary `[]any` paths. An integer or integral float selects the element; out of range (including beyond int64 and the infinities) stays undefined; an undefined index propagates as undefined; a string, bool, null or container index and a fractional or NaN index error with `list index must be an integer, got ...`. Dict access is unchanged. The legacy licence header was rewritten to the SPDX form.
- `builtins/scalar.go`: `count` no longer declares `MismatchUndefined`; it carries `KindError` instead, so the precheck errors at runtime and `index/builtin_check.go` now rejects a statically known wrong kind such as `count(5)` at load. `implCount` returns undefined for undefined and errors for null, since the precheck passes both through. `MismatchUndefined` itself is kept (now unused) for the follow-up builtin audit.
- `runtime/eval_infix.go`: `+` with a string goes through `concatStrings`. Numbers render with a new `jsNumberString` that follows JavaScript's `Number.prototype.toString` (Go's `%v` gave `1e+20`, `1e-06` and `1.2345675e+06` where JavaScript gives `100000000000000000000`, `0.000001` and `1234567.5`); bools, null and an unknown trinary keep their text; a list, dict, document or callable beside a string errors with `+ cannot join a string with a <kind>`.
- Null and an unknown trinary beside a string were not covered by the recorded decisions. Null concatenates as `null` (what JavaScript does and what main did); unknown concatenates as `unknown` (main's behaviour, no JavaScript analogue). Both flagged for the orchestrator rather than decided here.
- Adjusted four existing tests that encoded the old `count` behaviour: `TestCount_OtherType`, `TestCount_Bool`, `TestGoldenBehavior/count wrong kind undefined` (renamed `count wrong kind errors`), and `TestBuiltinCheckCountMismatchUndefinedAccept` (now `TestBuiltinCheckCountWrongKindReject`). No existing test covered the list index or the container concatenation.
- Updated graph nodes `runtime.eval_access`, `runtime.eval_infix`, `runtime.eval_cast`, `builtins` and `index.builtin_check`. `go test ./...` and `make docs-graph` pass.
- Commit signing was waiting on the Secretive prompt for the first commit, so the implementation commits are queued for the orchestrator.
- _Model: claude-opus-5-5_

### 2026-11-10 16:36 - String + with undefined or unknown gives unknown
- The user decided the two open questions: `"a" + null` stays `"anull"`, and a string `+` with undefined or unknown, in either order, gives the trinary unknown. This supersedes the pickup bullet "undefined propagates as undefined" for string `+` only; arithmetic, comparison, access and list indexes still propagate undefined.
- `evalInfix` now runs `concatStrings` before the blanket undefined check; `concatStrings` returns `box.Trinary(trinary.Unknown)` when either operand is undefined or `ValueTrinary`. `1 + undefined` and `undefined + unknown` (no string) stay undefined.
- Updated `runtime.eval_infix` and the PR description. `go test ./runtime/... ./builtins/ ./index/ ./lang_test/` passes.
- Signing works again (4af9fdb landed), so the queued commits are made now.
- _Model: claude-opus-5-5_

### 2026-11-10 16:40 - Regression tests for #149
- Added `runtime/eval_invalid_operands_test.go` (end-to-end through a temp pack: list index on boxed and boundary `[]any` lists, count at runtime, string `+`), `builtins/scalar_test.go` (count null errors, count undefined stays undefined) and `index/builtin_check_test.go` (count rejected at load for bool and typed facts, countable kinds accepted).
- Findings: `count(null)` is not rejected at load (only at runtime); `count(u)` for a typed optional number fact is rejected at load, so undefined-through-count is tested via a lambda parameter. `lang_test` fixtures are parse fixtures, so language-visible behaviour went into the runtime suite instead.
- Bite proof: swapping `runtime/eval_access.go`, `runtime/eval_infix.go`, `builtins/scalar.go` and `index/builtin_check.go` back to main fails the new tests; the "stays the same" rows pass on both by design.
- _Model: claude-sonnet-5-5_

### 2026-11-10 16:45 - Docs for #149 (website#78)
- Pages: `reference/arithmetic-operations.md` (string concatenation section), `reference/types-and-values.md` and `language-concepts/type-system-shapes.md` (list index rules), `reference/built-in-functions.md` (`count`), `reference/errors.md`, `troubleshooting/common-errors.md`. Breaking changes are stated on the pages.
- Ran all new snippets against the branch binary: indexes, count, ten `+` examples, errors and string+undefined (unknown) all match. Records follow the list rules.
- Found, not fixed (candidates): `10 / count(xs)` parses as a call of `(10 / count)`; policy-level `let t = reduce(...)` yields a callable; the arithmetic page's `total / count scores` is a parse error; docs use `bool` where the binary wants `boolean`.
- Website build passes.
- _Model: claude-sonnet-5-5_

### 2026-11-10 18:01 - Test round 2: strengthen the unknown rows
- Reviewer found the rows "string plus undefined" and "undefined plus string" passed on main because `DecisionOf(undefined)` already sets State Unknown, so the earlier bite-proof claim for those rows was wrong. The `wantUnknown` branch now also asserts `Value.Kind() == box.ValueTrinary`.
- Bite proof: with `runtime/eval_infix.go` swapped to main, all four unknown rows (string/undefined, undefined/string, string/unknown, unknown/string) fail; they pass on HEAD.
- Added a callable row (`"a" + ((x) => { yield x })`) confirming a callable beside a string errors; documents hit the same default branch in `concatOperand`.
- _Model: claude-sonnet-5-5_
