# 160-1-contribute-rules

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #160
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Each rule has exactly one definition, so results cannot be built up from independent conditions.
- Decision (user, 2026-10-31): a new `contribute` keyword, e.g. `contribute deny when c { yield ... }`, adding to a declared rule. Result ordered by declaration, undefined contributions skipped, boolean rules OR their contributions. Contributions combine in a deterministic order (#143).
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:10 - Design choices beyond the recorded decision
- Rules had no declared type at all, so "a non-list/non-bool rule" and "a yield incompatible with the rule's `list[T]`" could not be index errors. Chose to add an optional declared type to `rule` (`rule deny: list[string] = []`), parsed exactly like `let name: T`. This is the issue's own proposed syntax and keeps contribution checks static. Rejected the alternative of inferring the kind from the base rule's runtime value: it gives no index-time errors and is ambiguous when the base guard does not hold.
- A declared rule type is enforced at runtime on the rule's body or default value (undefined and a guard that does not hold are not checked), so the annotation is never decorative.
- Only `list[T]`, `boolean` and `trinary` rules (nullable unwrapped) take contributions. Contributing to an undeclared rule, an untyped rule, another type, or a rule whose body is `import decision` is an index error. A contribution must follow its rule in the policy, matching how `export decision of` already requires a declared rule.
- `contribute` without `when` is allowed (unconditional; still skipped when the body yields undefined), matching `rule` where `when` is optional. No `default` and no import body on contributions.
- Literal yields are checked statically against the yield type; computed values are checked at runtime against the element type, and the combined list against the full declared type (so list constraints apply to the combined value).
- List rule: base list followed by applied contributions; a base that is not a list (undefined, or guard false with no default) starts empty. Boolean rule: Kleene OR of base state and applied contributions, so a skipped base stays `Unknown` - write `rule allow: boolean = false` for a false identity. All contributions are evaluated, no short-circuit, so errors and traces are complete; order is declaration order (#143).
- `contribute` is a real reserved keyword like every other keyword. No identifier, field or namespace segment in fixtures or `example_pack` uses it, but user packs that do would break (the lang_test fixture namespace had to avoid `contribute` as a path segment).

### 2026-10-31 14:41 - Implement contribute across lexer, parser, index and runtime
- `tokens`: `KeywordContribute`. `ast`: `ContributeStatement`, `RuleStatement.Type`. `parser`: `parseContributeStatement` (policy scope), optional `: T` on `rule`.
- `index/contribute.go`: `Contribution`, `ContributionKindOf`, `addContribution` with the static checks; `Rule.Contributions`; contributions feed the reference-cycle graph and the builtin-call walk.
- `runtime/contribute.go`: `validateRuleValue`, `applyContributions`, `evalContribution` with `rule-contribution` / `contribution-when` / `contribution-body` trace nodes recording `index` and `skipped`. `evaluateRuleOutcome` split into base plus contributions, so references, exports, attachments and imports see the combined value.
- Grammar (`grammar.ebnf`, `grammar.peg`), new graph nodes `parser.contribute`, `index.contribute`, `runtime.contribute`, and updates to grammar, parser.rule, parser.policy, parser.lookups, ast, index.rule, index.policy, index.validate, index.builtin_check, runtime.executor, lang_test nodes.
- Found while testing: an attachment cannot read its own exported rule (self-recursion through the ref stack); pre-existing, recorded in the runtime.contribute node, tests use an attachment on another export.
- Tests: parser, index and runtime suites plus `lang_test/0050-contribute-rules.sentrie`. Touched package trees and `make docs-graph` pass. Local `golangci-lint` panics (binary built with go1.26, toolchain go1.27), so lint was not run locally.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:51 - Test and CI pass
- Renamed the fixture to the branch number: `lang_test/0050-contribute-rules.sentrie` is now `0052-...`; queued commit script updated. Added `0052b-contribute-rule-types.sentrie` (trinary and nullable rules, contributions reading rules declared before and after, an importing policy).
- Added `runtime/contribute_gaps_test.go` (stable order under repeated and concurrent ExecRule/ExecPolicy, runtime error in a guard or body fails the rule and is reported by ExecPolicy, own-rule read errors instead of hanging, three-valued OR cases, skipped base, nullable elements, `not_empty` on the combined list, importing policy sees the combined list and boolean), `cmd/contribute_test.go` (validate reports each index error with a span, exec JSON carries the combined value and attachment, capabilities and nondeterminism checks reach contributions), and index tests (cycle through contributions, contribution reading another rule, trinary/nullable targets, reserved word, builtin-arity check inside contributions, policy statement kind).
- Bite proof: every touched production file was reverted to main (new files deleted) and the tests failed; two files initially survived (`index/policy_stmt.go`, `index/builtin_calls.go`), tests were added and now fail without them.
- Findings, not fixed (non-test code is out of scope): (1) a contribution that reads its own target rule passes `Validate` and fails only at runtime with "references itself" (same as plain `rule x = count(x)` on main, indirect cycles are caught at index time); (2) when a contribution errors, `ExecRule` returns the error together with an output whose decision is the base value (state False for an empty list or false base) instead of Unknown, so a caller that ignores the error sees a decided result; (3) `4 / user.two` fails with "right operand is not a number" (the slash-vs-path ambiguity, unrelated to this branch).
- Gates: make test, lint (0 issues), gofmt, docs-graph, -race on non-runtime packages and on the contribute runtime tests, no em dash all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-11-01 02:32 - Review round 1 fixes
- W1: every contribution and declared-type error path now returns a nil decision (`applyContributions` three paths, `evaluateRuleBase` default and body type checks), so `ExecRule` reports `Unknown` instead of the base value or `DecisionOf(val)`. The pre-existing body evaluation error path is unchanged. Tests assert `Unknown` on contribution evaluation errors, contribution type errors, and a bad base or default.
- W2: the literal yield check maps `record[...]` to list, since records are fixed-length lists at runtime; `dict` stays separate. Index and runtime tests cover `list[record[string, number]]` (a list literal passes, a dict literal is rejected).
- S1: a contribution whose guard or body names its own rule is now an index error (`detectContributionSelfReference`, run in `Validate` after the reference-cycle check). It uses the scope-aware walker, so a lambda parameter or block `let` of the same name shadows the rule, matching runtime resolution (locals first). The runtime self-reference test fixtures moved to index tests.
- S2: the null base becoming `[]` is named in the `runtime.contribute` node, the code comment and the PR, with a runtime test for `list[string]? = null`.
- S3 and W3: PR Review notes mention the `user["contribute"]` workaround; Testing notes name fixtures 0052 and 0052b, `cmd/contribute_test.go` and `runtime/contribute_gaps_test.go`.
- Each new test was shown to fail with its fix reverted. While doing that, a backup name clash overwrote `runtime/contribute.go`; it was rewritten from the reviewed version plus the W1 fix, and the full runtime, index and cmd tests pass on it.
- No new commit lines: none of the queued commits have run, every touched path is already in one, so the fixes fold into those commits (separate lines would be empty commits).
- Tests: `go test ./runtime/... ./index/... ./cmd/... ./parser/... ./lang_test/... ./internal/docsgraph/...` pass; `GOTOOLCHAIN=go1.25.0 make lint` with the go1.25 golangci-lint reports 0 issues.
- _Model: claude-opus-5-5, role: implementer subagent_
