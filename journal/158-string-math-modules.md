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
