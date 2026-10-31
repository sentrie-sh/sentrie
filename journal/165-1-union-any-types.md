# 165-1-union-any-types

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #165
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no union type and no top type other than `document`, which is dict-shaped.
- Decision (default proposed to the user on 2026-10-31, not overridden): `T1 | T2` in any typeref position, a value validates if it matches any member, constraints bind to the member they follow; `any` accepts every value including scalars. Discriminated unions are out of scope. When no member matches, report the closest member's failure.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:42 - Implement union types and `any`
- Lexer: a bare `|` is now `tokens.TokenPipe` rather than an error token. The parser registers an infix handler for it at PIPELINE precedence that reports "only '|>' is supported in an expression", so the old helpful message survives. The right of `is` parses at EQUALITY, so that handler never fires there. The first version did not consume the token and the Pratt loop spun forever; it now advances.
- Precedence: `|` binds looser than a member's `?` and `@constraint(...)` suffixes and than `[...]` brackets. `string | number?` is `string | (number?)`, `string | number@min(1)` constrains only the number, `list[a | b]` is a list of a union. No grouping parentheses: a union with a nullable member accepts null whichever member carries the `?`, so grouping adds nothing. The union node owns no constraints.
- `any`: not a lexer keyword, which keeps the `any(...)` quantifier and any other identifier named `any` working. In a type position a bare single-segment `any` is `ast.AnyTypeRef` (`com/ex/any` stays a shape reference). `shape any` is rejected at parse time so a declared shape cannot be silently shadowed. On the right of `is` a bare `any` is always the type; equality with a value named `any` needs `==`.
- `any` rejects undefined, like every type: undefined means no value, and `x is T` already returns false for undefined before validating.
- Closest-member rule (decision said to define it deterministically): if any member matches, pass. Otherwise, if any member's error is not a mismatch (unknown shape, a constraint with invalid arguments), return the first such error in member order, so `is not` cannot fail open. Otherwise the closest member is the first in source order whose outer kind check accepts the value (`outerKindAccepts`: read through shape aliases, `T?` also accepts null); its error is wrapped `closest member <T>: ...`. If no member's kind accepts it, report the union itself. Rejected alternative: scoring by how many nested fields matched, which is harder to predict and unnecessary without discriminated unions.
- Shape resolution pulled out of `validateAgainstShapeTypeRef` into `resolveShapeTypeRef` (same behaviour) so `outerKindAccepts` can resolve shape members.
- Fact injection's null gate now uses `ast.AcceptsNull` (`T?`, `any`, or a union with such a member) instead of `IsNullableTypeRef`.
- Index: `checkIsTypeShapes` walks union members; `typeRefKind` returns unknown for `any` and unions.
- Left alone: `bindTruthValue` (decision envelope into state) only for a bare trinary type, not for a union with a trinary member. Recorded in the new graph node.
- Grammar and graph nodes updated; new node `runtime.typeref_union`. `make docs-graph` passes. `make lint` panics locally (golangci-lint built with Go 1.26, toolchain newer), unrelated to this change.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:50 - Tests and CI gate for union types and `any`
- Renamed the implementer's fixture `lang_test/0050-union-any-types.sentrie` to `0055-union-any-types.sentrie` (this branch's reserved number) and fixed the queued commit script.
- Added tests for the gaps: unions and `any` through let annotations, `as` casts, derive returns and parameters (including optional), shape fields, alias members, nested `list[list[number] | string]`, `dict[...]`, JSON-decoded facts (float64), required and optional union facts, null handling, `is not`, closest-member error text and its determinism over 25 runs, a broken member (unresolved shape, bad constraint argument) failing closed for both `is` and `is not`, `typeRefKind` unknown for union and `any`, `sentrie validate` accepting unions and rejecting an unknown shape in a member, and the bare `|` error with `|>` still lowering.
- Finding: `dict[T]` never validates its values (`runtime/typeref_dict.go` only checks the dict kind and constraints), so `{ "a": true } is dict[number | string]` passes. Pre-existing and unrelated to unions; tests assert only acceptance for dict members.
- Finding: `index/builtin_kind.go` listing `AnyTypeRef` and `UnionTypeRef` explicitly is behaviour-neutral (the fall-through already returned unknown), so no test can fail without it.
- Gate: make test, lint (0 issues, CI-pinned linter), gofmt, docs-graph, race on the touched non-runtime packages, no em dash all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_
