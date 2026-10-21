# 158-string-math-modules

## Binaek Sarkar (binaek89@gmail.com)

### 14:39 - Pick up #158: implement @sentrie/string and @sentrie/math
- Issue: both modules are registered (`runtime/executor.go`) but their providers export nothing, and `use { ... } from @sentrie/string` does not parse because `string` lexes as `KeywordString` while `parseUseStatement` wants `Ident` path segments. Taking the issue's option 1: implement them and accept keywords as module path segments.
- Branched off `main` at b35fe4b (matches `origin/main` per `gh api`). Worktree `worktrees/158-string-math-modules`.
- Parallel work: #111 is converting every builtin to throw instead of returning `Error` objects. New code here throws from the start (`panic(vm.NewGoError(...))`) and stays out of the other `builtin_*.go` files, `registry.go`, `executor.go`, `modules.go` and `tscompile.go`. Registration already exists, so no registry edit is needed.
- Found `ts_src/sentrie/string.d.ts` and `math.d.ts` already declaring a different, larger API (`toLowerCase`, `trimLeft`, `sin`, `random`, ...) that was never implemented. The declarations will be rewritten to the implemented surface, since a declaration with no implementation behind it is what made the empty modules confusing in the first place.
- Module guide lives on the website repo (`sentrie.sh/typescript-modules/`), not in this repo, so in-repo documentation is the `.d.ts` reference plus the graph. The website gap is flagged for the parent.
- _Model: claude-opus-5-5_

### 14:46 - Parser, string module, math module
- Parser (d24ed96): `advanceModulePathSegment` in `parser/use.go` accepts an `Ident` or any keyword token as an `@` path segment. Keyword detection is `tokens.IsKeyword(tok.Value)` plus a kind match, since the lexer keeps the literal text as the token value; no change to `tokens` or the lexer, so keywords parse exactly as before everywhere else.
  - A keyword last segment with no `as` is now a parse error (`default alias "string" is a keyword; add 'as <alias>'`). The default alias would otherwise be `string`, and `string.lower(x)` cannot parse at the call site because `string` lexes as the keyword. A clear error at the `use` beats an unrelated expression error later. Flagged for review: `use { ... } from @sentrie/string` without `as` does not work.
  - Grammar: `execSource` now reads `'@' modulePathSegment ('/' modulePathSegment)*` (it previously claimed exactly two segments, which the parser never enforced), with `modulePathSegment ::= IDENT | KEYWORD` in both `.ebnf` and `.peg`.
- `@sentrie/string` (91b1781): lower, upper, trim, trimStart, trimEnd, startsWith, endsWith, includes, split, join, replace, replaceAll, equalFold, format.
  - Strict types: every string argument must be a JS string; no coercion of `null`/`undefined`/numbers. `join` takes an array (JS array or the goja-wrapped Go slice a policy list becomes) of strings.
  - Unicode: Go `strings` works on UTF-8 code points; `split(s, "")` splits into code points, trims strip Unicode White_Space, case mapping is Go's simple one-to-one mapping (`upper("ß")` stays `ß`). Documented rather than reimplementing full special casing.
  - `format` is a hand-written closed verb set (`%s %q %d %f %.Nf %v %%`) instead of `fmt.Sprintf`, so Go's `%!d(string=...)` noise never reaches policy output and the argument count must match exactly. `%.Nf` uses `strconv.FormatFloat`, so ties round to even (`%.0f` of 2.5 is 2); documented and tested.
- `@sentrie/math` (90381d8): abs, floor, ceil, round, trunc, min, max, pow, sqrt, clamp.
  - Decision: no non-finite value in or out. NaN/Infinity arguments throw, NaN/Infinity results throw (sqrt of a negative, overflowing pow, pow(0, -1)), and negative zero is returned as 0. Reason: NaN compares false with everything, which silently flips policy comparisons, and Infinity does not JSON-encode in a decision. IEEE semantics stay available through `@sentrie/js`. Flagged for review.
  - `round` is half away from zero (Go `math.Round`), unlike JS `Math.round` for negative halves. Chosen as the symmetric arithmetic convention; documented and flagged.
  - `min`/`max` also take a single array, because a policy cannot spread a list into arguments.
- Errors: every failure is `panic(vm.NewGoError(...))`, matching where #111 is taking the builtins. Helpers are module-prefixed (`stringThrow`, `mathThrow`, ...) so they cannot collide with whatever shared helper #111 introduces; `math` reuses `stringTypeName`.
- `.d.ts`: `ts_src/sentrie/string.d.ts` and `math.d.ts` rewritten to the implemented surface. The old ones declared a different, never-implemented API (`toLowerCase`, `trimLeft`, `substring`, `padStart`, `sin`, `random`, constants, ...).
- Tests: `JSTestSuite` cases per function in `builtin_string_test.go` / `builtin_math_test.go`, parser cases in `parser/use_test.go`, a `RuntimeTestSuite` rule-evaluation test in `runtime/js_string_math_test.go` (policy facts must precede `use` statements, which the index enforces), and fixture `lang_test/0048-string-math-modules.sentrie`. Ran the fixture's rules through the executor once with a throwaway test (deleted) to confirm every imported export binds and evaluates.
- _Model: claude-opus-5-5_

### 14:47 - Graph, tests, verification
- Graph: `runtime.js.builtins` gains contract entries for both modules and notes that they throw, unlike the rest of the family until #111 lands; `runtime.js` gets the same exception, `lang_test.lang_test` now counts forty-four fixtures. `make docs-graph` passes.
- Committed the runtime test and fixture (24fae02) and the graph update.
- `make test` passes. `make lint` passes (0 issues) with golangci-lint v2.4.0 built on go1.25.0, per the coordinator's recipe, because the host binary panics on the newer local Go.
- Not done here: the website user guide (`sentrie.sh/typescript-modules/`) has no page for these modules; it lives in a separate repo.
- _Model: claude-opus-5-5_

### 2026-10-21 22:26 - Review fixes from pr-199
- Warning, allocation: `min`/`max` (array form) and `join` sized their buffers with `make(..., n)` from the JS-controlled `length` before checking any element. `a.length = 2**32 - 1` is free in JS and asked Go for 34 GB (68 GB for `join`), which on Linux overcommit settings is an unrecoverable `fatal error: out of memory`. Fix: named limits `mathMaxArrayLength` and `stringMaxListLength` (both 2^20, about 8 MB of float64) checked before allocating, and buffers grown with `append` from a capacity of `min(n, 64)` while validating, so a bad element throws before anything large exists. 2^20 is far above any realistic policy list while keeping the worst case small. `split` and `format` size from the input string, which goja already holds, so they need no limit.
- Warning, `%f` negative zero: `format("%f", -0)` printed `-0.000000` while `%d` and `%v` printed `0`. Now normalized the same way. A negative value that rounds to zero (`%.0f` of -0.4 gives `-0`) keeps its sign, matching `toFixed`.
- Warning, merge with #195 (`111-throw-builtin-errors`): #195 adds `throwErr(vm, err)` as the single throw helper. `stringThrow` and `mathThrow` should collapse into it in whichever PR lands second, and the graph conflict in `runtime.js.builtins.md` / `runtime.js.md` resolves to #195's text plus this PR's string/math signature blocks. Recorded in the PR Review notes; no code change here, since `throwErr` is not on this branch.
- Suggestions applied (all small): an unsupported multi-byte verb is reported as its rune (`%é`, not `%Ã`); `stringTypeName` names `bigint` and `symbol` instead of Go types; the keyword default-alias parse error now points at the path segment rather than the token after the statement; `string.d.ts` documents the `toFixed` difference with examples and that `trim` strips NEL but keeps a BOM; `use { f } from @sentrie/ as x` is pinned by a test so it cannot start parsing as module `@sentrie/as`. The `pow` ULP note needed no action.
- The first `@` path segment now also extends the statement's range, so a one-segment `@x` path no longer ends its span at the `@`.
- `make test`, `make lint` (go1.25.0 golangci-lint) and `go test -race ./runtime/js/ ./parser/` pass.
- Commit signing was refused (`agent refused operation`), so all of this is left uncommitted in the worktree for the parent, per the dispatch rules. Intended split: (1) bound min/max (math files, `math.d.ts`, graph node); (2) bound join, `%f` negative zero, string error names (string files, `string.d.ts`); (3) keyword alias error position (`parser/`); (4) this journal entry.
- _Model: claude-opus-5-5_
