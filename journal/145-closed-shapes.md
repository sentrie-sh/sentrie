## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 22:54 - Pick up #145
- Shapes accept fields they don't declare, `dict[T]` values are unchecked, a shape can declare the same field twice, and duplicate dict literal keys overwrite each other silently.
- Picked up in batch 2 of the v0.0.5 milestone (#148, #145, #144, #155, #147), locally while the user is AFK: SSH signing is down, so this stub and its push are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Shapes are closed by default, nested shapes too. The opt-out is a `...` marker as the last entry inside the shape body (user picked this over `@open`).
  - `dict[T]` values are checked against T, including constraints, and the error names the key (check whether #299, already merged, covers this).
  - A duplicate shape field is an error that shows both spans.
  - A duplicate dict literal key keeps last-one-wins, and the docs say so (user overrode the recommendation).
- PROVISIONAL (user AFK), needs confirmation: branch stacked on `159-keyword-field-names`, which changes shape field parsing (parser/shape.go). PR base `159-keyword-field-names` until #159 merges.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5_

### 2026-11-10 23:01 - Implement closed shapes
- Re-reproduced on the base (`159-keyword-field-names`) with a scratch pack: an extra field passes a shape, and `shape S { a: string a: number }` keeps the last `a`. The `dict[T]` claims do NOT reproduce: #299 already checks each value against `T` with its constraints and names the key (`value for key "a" is not valid ...`). Nothing changed there. A duplicate dict literal key still collapses to the last value, as decided.
- Closed by default: `runtime/typeref_shape.go` rejects the first undeclared key (sorted) unless `ShapeModel.Open`. Nested shapes, list items, dict values and record fields get it for free, since they recurse through the same validator. The check runs after the declared-field walk: first attempt ran it before, which changed the message of 7 derive-scope tests in `runtime/derive_shape_scope_test.go` (they send the other scope's fields and expect "field x is required"); reporting missing or wrongly typed fields first keeps those messages and is the more useful error, so no existing test was changed.
- Open marker: `parser/shape.go` reads `...` into `ast.Cmplx.Open` / `OpenRange`; a field or a second `...` after it is a parse error naming the marker's span; EOF after it falls through to the usual missing `}` error. `index/shape.go` copies it to `ShapeModel.Open`. `format` prints it after the fields and `format.Equal` orders it with the members.
- PROVISIONAL (user AFK), needs confirmation: openness is per declaration, never inherited through `with` (an open base does not open the child; an open child of a closed base is open). Conservative: never accepts more than the shape that is named says.
- PROVISIONAL (user AFK), needs confirmation: comments may follow `...`; the "anything after it is a parse error" rule is applied to fields and markers only.
- Duplicate shape field: parse time, in `parseComplexShape`, the only place both fields exist (the AST keys fields by name). Error at the second span, naming the first. The old check in `index/shape.go` `createShape` is unreachable for parsed shapes; left alone.
- Grammar (`complexShape ::= ... '{' shapeElement* '...'? '}'`) and graph nodes (`parser.shape`, `runtime.typeref_shape`, `index.shape`, `ast`, `format`, `overview`) updated. `make test` and `make docs-graph` pass (`/tmp/sentrie-145-closed-shapes-impl-test.log`).
- _Model: claude-opus-5-5_

### 2026-11-10 23:22 - Test closed shapes
- Added `runtime/closed_shapes_test.go` (extra field on a closed shape; open shape; nested shape, list item and dict value closed; an open outer shape does not open its nested shape; openness not inherited through `with`, an open child of a closed base is open; undeclared field reported after declared ones; `dict[T]` value type and constraint pinned, with the key named; duplicate dict literal key last-one-wins pinned), `parser/shape_test.go` (marker accepted after fields, alone, with comments around it; field or second marker after it rejected; EOF after it falls through to the missing `}` error; duplicate field error names both spans, keyword names included), `format/open_shape_test.go` (round trip and stability; `Equal` tells open from closed and sees a comment moved across the marker), and `lang_test/0061-closed-shapes.sentrie`.
- Bite proofs against `159-keyword-field-names`: `parser/shape.go`, `runtime/typeref_shape.go`, `format/format.go` and `format/equal.go` each make the matching tests fail; `index/shape.go` and `ast/shape.go` carry the new fields and the base does not compile against the tests. The `dict[T]` pins and the duplicate dict key pin pass on the base too, by design (#299 covers dict values; the key behaviour is unchanged). The "declared field reported first" test also passes on the base; it guards the implementation's check order.
- Gotcha for fixtures: a rule-level typed `let` is lazy, so a rule that never reads the variable never fails the check. The rule tests read it with `v is defined`.
- Local gate: `make test`, `make lint` (0 issues), `make docs-graph`, gofmt, race on parser, format, index, ast and lang_test all pass; no em dash; no signature lines.
- _Model: claude-sonnet-5-5_

### 2026-11-10 23:36 - Docs for #145
- Website branch 145-closed-shapes (from origin/next); pages: reference/shapes.md (new "Closed and open shapes" section, syntax, composition row, Good to Know), reference/types-and-values.md (dict constraints, shapes as T, last-wins keys), language-concepts/type-system-shapes.md (syntax, table row, caution). Keyword field names (#159) deliberately not documented.
- Documented as implemented: closed by default at every depth, `...` last entry with comments allowed after it, openness per declaration and not inherited through `with` (PROVISIONAL, flagged in the website PR description), duplicate shape field parse error with both spans, duplicate dict literal keys last-wins, dict[T] value and key errors (already #299).
- Ran every added snippet against the branch binary (validate plus exec with facts); all behaved as shown. Pre-existing broken snippets (unquoted map keys, policy with no export, two-file example, fragments) recorded in the website PR description, not fixed.
- Build: npm run build passes.
- _Model: claude-sonnet-5-5_
