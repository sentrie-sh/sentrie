## Binaek Sarkar (binaek89@gmail.com)

### 01:14 - Pick up #137
- Bug (fail-open class): `evalInfix` evaluates both operands eagerly, then returns `Undefined` when either is undefined, before the `and`/`or`/`xor` cases run. So `false and user.missing` and `true or user.missing` are both Unknown instead of False and True, and `true or ("a" < 1)` errors even though the left side already decided it.
- Checked before pickup: no open or closed PR mentions #137, no `worktrees/137-*` worktree, local `main` equals origin `main` (3301d9a). Builds on #135 (one boolean form, PR #252), #134 (PR #251), #136 (import envelope state, PR #253) and #150 (PR #254); no second boolean form is introduced.
- Decision agreed with the maintainer before the implementer was spawned:
  - `and`, `or` and `xor` use Kleene three-valued logic with an undefined operand read as `unknown`: `false and X` is false and `true or X` is true even when X is undefined or unknown; `true and unknown` and `false or unknown` stay unknown.
  - `and`/`or` short-circuit: when the left side decides the result the right side is never evaluated, so its errors, JS calls and cost do not happen. When the left side is unknown or undefined the right side is evaluated, because it may still decide; an error there still fails the expression.
  - `xor` cannot be decided by one side, so it evaluates both, with undefined read as unknown.
  - Arithmetic, comparison, membership and `matches` keep propagating undefined.
  - `not <undefined>` stays `Undefined` (maintainer chose to keep `evalUnary` unchanged and #137 scoped to the binary operators). Rule results and `when` gates read it as Unknown either way.
- Interactions checked, no change needed: the rule `when` gate takes the body only on `IsTrue()`, so False and Unknown both take `default` (`runtime/executor.go` `execRule`). After the fix `when true or user.missing` takes the body instead of the default, which is the intended correction. `default` itself is an ordinary expression and gains the same operator semantics.
- Plan: implementer and reviewer on Opus (language semantics), a separate test and CI agent proving each regression test fails against `main`'s `runtime/eval_infix.go`.
- _Model: claude-opus-5-5_
