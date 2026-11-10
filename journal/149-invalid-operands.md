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
