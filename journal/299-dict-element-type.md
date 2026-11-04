## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #299
- A `dict[T]` type never checks its values against `T` at runtime, so `{"a": "not a number"} is dict[number]` is `true`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: validate every value of a `dict[T]` against `T`, with an error naming the failing key as `list[T]` names an index, in every type position.
- _Model: claude-opus-5-5_

### 2026-11-03 21:11 - Check dict values against the element type
- `validateAgainstDictTypeRef` now validates every value against `ValueType` through the shared `validateValueAgainstTypeRef`, so facts, typed lets, casts, `is`, shape fields, derive parameters and returns, lambda returns and nested types all pick it up with no per-site change.
- Keys are walked in `slices.Sorted(maps.Keys(...))` order so the reported key is stable; Go map order would name a different key per run.
- Error is `value for key '<k>' is not valid at <pos>: <inner>`, modelled on the list wrapper `item is not valid at <pos>: <inner>` plus the key. It wraps with `%w`, so `isTypeMismatch` sees the value's own marker and `x is dict[T]` with a bad value is `false`, not an error.
- `list[T]` already validates its elements; its error names no index (the `TODO` in typeref_list.go). Left alone as out of scope.
- The fix broke three #136 envelope tests that used `dict[trinary]` (and `dict[bool]`, where `bool` is not a type and resolved as a missing shape only once values were checked) as a loose dict type. An envelope's `why` attachment is a string, so it does not conform to `dict[trinary]`. Switched the derive parameters and the `rl`/`rn` facts to `document`, which keeps their intent, and turned `rb` into a `dict[trinary]` fact that now asserts the binding fails.
- `example_pack` declared `let u7: dict[number @min(2)] = { "one": 1, ... }`, which only passed because values were never checked; changed the bound to `@min(1)` and added the SPDX header the file lacked.
- Noticed: an unknown shape inside a fact's `dict[...]` (`dict[bool]`) is not rejected when the index is built; it only fails at runtime once a value is checked. Not fixed here.
- _Model: claude-opus-5-5_

### 2026-11-03 21:15 - Test pass for dict element validation
- Added `TestDictElementTypeStructuredValues` in `runtime/typeref_dict_test.go`: `dict[Item]` with a value missing a shape field and with a wrong field type, `dict[record[...]]` with a bad element, a lambda (via `collect`) returning a bad `dict[number]`, a fact dict of shapes naming key `z` stably over 10 runs, a fact dict of records naming key `b`, and empty dicts (fact and `is`) conforming.
- Proved each new case bites by running the dict tests against `main`'s `typeref_dict.go`: every new bad-value case failed, then the file was restored and its diff checked.
- Ruled out: dict constraint ordering (`DictContraintCheckers` is empty, so a dict cannot fail both a value check and a constraint today; values are checked first, which is deterministic); `dict[A | B]` (the language has no union type refs).
- Checked the example pack: `sentrie exec sh/sentrie/example/var_test/test --pack-location example_pack` returns `{one:1, two:2, three:3}` and `sentrie validate` loads the pack.
- _Model: claude-sonnet-5-5_

### 2026-11-03 21:19 - Review fixes
- The key in the per-value error is now formatted with `%q` (`value for key "b" ...`), since fact keys are untrusted input and could hold quotes or control characters; tests and the graph node follow.
- Folded the `rl` envelope policy into `rd`: once `rl` became `document` the two differed only by the `nested` rule.
- Added a TypeScript module test: `let d: dict[number] = m.bad()` where the module returns `{ a: 1, b: "two" }` fails on key `b`, pinning the JS-object path.
- `runtime.typeref_document` claimed the dict and document validators were structurally the same; reworded as the dict validator without the per-value check.
- Dated the 2026-11-03 headings, as the branch spans two days.
- _Model: claude-opus-5-5_
