## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 22:54 - Pick up #144
- `bool`, `int` and `float` are not types, and `boolean` accepts `unknown` (the parser/typeref.go:88 flag is never read).
- Picked up in batch 2 of the v0.0.5 milestone (#148, #145, #144, #155, #147), locally while the user is AFK: SSH signing is down, so this stub and its push are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - `bool`, `float` and `int` are all parse errors with a "did you mean boolean/number" hint (user overrode the recommendations).
  - `boolean` is essentially a `trinary @not(unknown)`: strictly true or false, modelled as a trinary that rejects unknown.
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#85 (draft, base `next`); Sentrie PR #325.
- _Model: claude-opus-5-5_

### 2026-11-10 23:05 - Implement #144
- Reproduced on `main`: `let x: bool/int/float` fail as missing shapes, and `let d: boolean = unknown` is accepted (rule yields unknown). Did not reproduce: `example_pack` no longer uses `int`/`float` (#282 fixed it), and `validate` already reports the missing shape (it does not pass any more). The `TrinaryTypeRef.Boolean` flag was read, but only by `format`.
- Parser: `checkNotATypeName` in `parser/typeref.go` rejects a single-segment `bool`/`int`/`float` in a type position with `unknown type "bool": did you mean "boolean"?` (`number` for the other two); `parser/is.go` applies it where a suffixed or union name after `is` becomes a type. The names stay identifiers (not keywords), so `let int = 5`, fields named `bool` and `x is int` against a value keep working. The hint table is `ast.NotATypeHint` in `ast/typeref_any.go`.
- PROVISIONAL (user AFK), needs confirmation: `shape int = ...` is a parse error with the same hint, since such a shape could never be referenced. A bare `x is bool` stays an index error (it may be a value), now with the hint in `index/is_resolve.go`.
- Runtime: `validateAgainstTrinaryTypeRef` rejects unknown when `Boolean` is set (`value 'unknown' is not a boolean ... use trinary to allow unknown`), and the non-truth messages now say `trinary` or `boolean` instead of `bool`. Casts, `is`, unions, sets, facts, lets, derive params and returns and rule types all go through it, so no other site needed a change. `TrinaryTypeRef.String()` renders `boolean`. `index/builtin_kind.go` gives `boolean` the static kind `ValueBool`.
- PROVISIONAL (user AFK), needs confirmation: `runtime/contribute.go` skips the combined type check for a `boolean` rule whose OR is unknown. Only a skipped base with no true contribution can produce it (the base value and every contribution are checked against boolean), and a boolean rule without contributions also reports unknown with no error when its guard fails. Without this, `rule r: boolean = when x {...}` plus a false contribution would error.
- Tests updated for the decided behaviour: parser fixtures that wrote `int` as a type (`statement_test.go`, `program_test.go`, `comment_test.go`); `contribute_gaps_test.go` three-valued OR rules that OR `unknown` now declare `trinary`; `import_envelope_test.go` binds the unknown import to `trinary`; `set_test.go` checks the unknown-state set error through `set[trinary]`; `index/contribute_test.go` message now says `boolean`; `typeref_trinary_test.go` messages, plus one boolean-rejects-unknown subtest.
- Graph nodes updated: ast, ast.typeref, parser.typeref, parser.is, parser.shape, index.is_resolve, index.builtin_kind, runtime.typeref_trinary, runtime.contribute. `make docs-graph` and `make test` pass (`/tmp/sentrie-144-boolean-int-float-impl-test.log`).
- _Model: claude-opus-5-5_

### 2026-11-10 23:25 - Test and CI for #144
- Added tests for every decision bullet: `parser/typeref_not_types_test.go` (17 positions x bool/int/float give `unknown type "x": did you mean ...`; the three names stay usable as lets, fields, parameters; a shape named bool/int/float is a parse error, PROVISIONAL), `index/is_resolve_test.go` `TestResolveIsBareNotATypeName` (bare `x is bool` index error with hint, PROVISIONAL; `let int` still resolves), `index/boolean_kind_test.go`, `ast/typeref_boolean_test.go`, `runtime/boolean_unknown_test.go` (boolean accepts true/false and rejects unknown for facts, rule returns and casts; trinary accepts unknown; a contribution yielding unknown to a boolean rule errors; a skipped base with only a false contribution is unknown without error, PROVISIONAL; a true contribution rescues a skipped base), and fixture `lang_test/0060-boolean-not-bool-int-float.sentrie`.
- Bite proofs by copy/restore against `main`'s production code (or a surgical disable where the main file would not compile): parser/is.go, parser/shape.go, index/is_resolve.go, runtime/contribute.go, ast/typeref_trinary.go, index/builtin_kind.go, the checkNotATypeName call in parser/typeref.go and the Boolean unknown check in runtime/typeref_trinary.go all make the new tests fail. The lang_test fixture is a still-works guard and does not fail on main.
- Local CI gate: make test, make lint (0 issues), make docs-graph, gofmt, race on ast/index/parser, em dash and attribution checks all pass.
- _Model: claude-sonnet-5-5_

### 2026-11-10 23:30 - Docs for #144
- Website branch 144-boolean-int-float (base next), uncommitted while signing is down. `types-and-values.md` gets a breaking-change caution (bool/int/float are parse errors with a hint, still valid identifiers, a shape named so is a parse error, PROVISIONAL), the `boolean` row (strictly true/false, rejects unknown) and a note that casting unknown to boolean errors. `trinary.md` says unknown is only valid in trinary. `: bool` became `: boolean` across 8 reference pages, `int` became `number` in example-pack, and type lists/tables on index, identifiers, constraints, rules and type-system-shapes were fixed.
- Snippets: let.md, facts.md and example-pack.md snippets pass `validate` (facts also exec'd, null boolean fact rejected); the 35 other touched fragments (no namespace/policy) were switched to `text` fences. Hint errors, `let int`, `rule r: boolean = unknown` and `unknown as boolean` checked by hand on the branch binary. `npm run build` passes.
- Follow-up: make the `text` fragments runnable again. Website PR description is PR_DESCRIPTION_144-boolean-int-float.md. #149 (website#78) may conflict on types-and-values.md.
- _Model: claude-sonnet-5-5_

### 2026-11-10 23:34 - Round 2 review fixes for #144
- Reviewer (round 1): `shape S with int { ... }` had no hint and failed later with "shape com/a/int not found". `parseComplexShape` in `parser/shape.go` now calls `checkNotATypeName` on the `with` FQN. Added a "shape with" case to `TestRejectNotATypeNameInEveryTypePosition` (bites: 3 subtests fail with the check removed). `parser.shape` node and the PR description mention the `with` base.
- Swept the other `parseFQN` callers: `namespace` and `import ... from` are not type positions; every type position goes through `parseTypeRef` or the suffixed `is` path, and the bare `x is int` case gets its hint in `index/is_resolve.go`. No other gap.
- `make test` PASS (/tmp/sentrie-144-boolean-int-float-impl2-test.log), `make lint` 0 issues, `make docs-graph` PASS.
- _Model: claude-opus-5-5_

### 2026-11-11 04:00 - Docs for #144 (round 2)
- Website branch 144-boolean-int-float, changes staged but uncommitted (signing refused twice). `trinary.md` now says an `unknown` value yielded by a boolean rule, let, fact or cast is an error, while a boolean rule with no value (failed `when` and no `default`, or a skipped base with no `true` contribution) still reports unknown. `rules.md` contribution row states both cases. `built-in-functions.md`: `any` and `all` examples are now full `sentrie` policies.
- Verified on the branch binary: failed-when boolean rule gives unknown; skipped base with a false contribution gives unknown, with a true one gives true; `yield unknown`, `default unknown` and an unknown contribution on a boolean rule error. any/all policies exec'd (true, true, true, false). `exec --output json` cannot marshal a `let` holding lambdas (table output works): possible Sentrie follow-up. No page talks about `with` bases and type names, so no `with` mention. `npm run build` passes (/tmp/website-144-boolean-int-float-build.log).
- _Model: claude-sonnet-5-5_
