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
