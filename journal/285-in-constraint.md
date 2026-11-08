## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #285
- The `@in` number constraint is defined but cannot be written, because `in` is a keyword and the parser expects an identifier after `@`.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer: the parser accepts a keyword token as a constraint name right after `@`, where it cannot be ambiguous, so `@in` works as documented and nothing is renamed.
- _Model: claude-opus-5-5_

### 12:57 - Accept a keyword as a constraint name
- `parseTypeRefConstraint` now reads the name with `advanceConstraintName`: an `Ident` or any keyword token, the same check `advanceModulePathSegment` uses for `@scope/name` segments.
- Rule chosen: any keyword, not only keywords that are constraint names. It needs no second list to keep in sync with the generated tables, and a keyword that is not a constraint of the type (`@policy(1)`) now goes through `AddConstraint` and gets the existing unknown-constraint error (`cannot add constraint policy: not found`) instead of `expected Ident, got policy`. Ruled out a keyword allow-list: it would have kept the token error for other keywords, which reads as a syntax problem rather than an unknown name.
- Keyword constraint names checked against every generated table in `ast/typeref_constraint_args_gen.go`: `in` (number) is the only one. `not_in`, `one_of`, `not_one_of` and `has_substring` are identifiers.
- Spans: the constraint still spans `@` through `)`; the name has no span of its own on `TypeRefConstraint`. `@` before a non-name token (`@ = 1`, `@)`) is still `expected Ident`; `@in` with no parentheses is `expected 'LeftParen'`.
- Grammar files gain `constraintName ::= IDENT | KEYWORD`; the "only position where a keyword is a name" notes in the grammar, `parser/use.go` and the graph now name both positions.
- Tests: parser suite (`@in`, `@not_in`, two constraints, unknown keyword name, missing arguments) and one runtime test that `number @in([1, 2])` accepts 1 and rejects 3 in a let, a fact and `is`. The parser tests fail on main.
- Rewrote the legacy header of `runtime/typeref_number_test.go` on touch.
- _Model: claude-opus-5-5_

### 13:10 - Test `@in` in every type position
- Added `TestNumberInConstraintInEveryTypePosition` (runtime): `@in` accepted and enforced in a list element type, shape field, shape alias, policy-local shape, derive parameter, fact, and `is` / `is not`, with a value in and out of the set; rejections assert the "is not in the set" message so a different error cannot satisfy them.
- Added two `cmd` validate tests: a pack using `@in` validates (the #285 reproduction), and `@policy(1)` is an unknown-constraint error, not `expected Ident`.
- Added fixture `lang_test/0050-in-constraint.sentrie` (fixtures are the house style for language features: parse, index, validate). Its namespace is `constraints/inset` because a namespace segment `in` is still a keyword error; that is outside this issue.
- Proved the tests bite: with `parser/typeref.go` restored from main, the new runtime, cmd, fixture and parser tests all fail.
- Ruled out: parser-level duplicates per position (the runtime test already parses each one), and a test for `@in` on `string` beyond the existing unknown-name case.
- Bumped the fixture count in the `lang_test` graph node (45 to 46 fixtures; it said forty-four).
- _Model: claude-sonnet-5-5_

### 13:13 - Share the name-or-keyword helper (review)
- Review: `advanceConstraintName` was a copy of `advanceModulePathSegment`. Both are now one helper, `advanceNameOrKeyword(p, expected)` in `parser/parser.go` next to `advanceExpected`, whose doc comment names the two positions where a keyword is a name.
- Took the error-text suggestion: the helper takes the expectation text. Module paths pass `tokens.Ident.String()`, so their errors are unchanged (`expected Ident, got <kind>`); constraints pass `constraint name after '@'`, so `number @ = 1` now reports `expected constraint name after '@', got Assign` instead of `expected Ident, got Assign`.
- Graph: `parser.parser` gains the helper; `parser.typeref` and `parser.use` name it.
- Parser, cmd and lang_test tests, lint and docs-graph pass.
- _Model: claude-opus-5-5_

### 2026-11-08 14:25 - Rebase onto main after #269
- #269 (merged as #274) appended a cast cycle test to `cmd/validate_test.go` where this branch appends its `@in` validate tests. Kept both; `go test ./cmd/` passes.
- _Model: claude-opus-5-5_
