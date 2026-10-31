# 163-string-ordering

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #163
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: `<`, `<=`, `>`, `>=` accept only numbers; two strings should compare by byte order (`strings.Compare`), mixed kinds stay an error.
- No design choice left open by the issue.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:32 - Order strings with relational operators
- `runtime/eval_infix.go`: added `compareStrings`, which returns `strings.Compare` when both operands are strings; `<`, `<=`, `>`, `>=` try it before `box.MustNumbers`. The numeric lines are unchanged so the concurrent `box` numeric rework (#166) merges cleanly.
- Decision: a mixed pair (string and number, or string and unknown/bool) falls through to the numeric path and keeps its existing `left/right operand is not a number` error rather than gaining a new message, so existing error tests and messages stay valid.
- Undefined and unknown handling needs no new code: undefined short-circuits before the switch for every operator, and an unknown trinary is not a string, so it reaches `MustNumbers` and errors exactly as before.
- Static checks: index and validate do not inspect relational operand kinds (`index/validate.go` only walks infix operands for the dependency graph), so nothing rejected string operands statically and nothing needed aligning.
- Documentation: no prose operator docs live in the repo, so the byte-order (not locale-aware) rule is recorded in comments above the relational production in `grammar/grammar.ebnf` and `grammar/grammar.peg` and in the `runtime.eval_infix` graph node.
- Rewrote the legacy Apache header of `runtime/eval_infix.go` to the two-line SPDX form, per the license-header rule for touched files.
- Tests: `TestEvalInfixRelationalOrdersStringsByByte` covers all four operators, equality edges, prefix, empty string, case and digit byte order, ISO timestamps, multibyte UTF-8, mixed-kind errors both ways, string vs unknown, and undefined on either side. `./runtime/...` and `./lang_test/...` pass; `make docs-graph` passes.
- Problem: local `golangci-lint` panics (built with go1.26, module needs go1.27); `go vet ./runtime/` is clean. Environment issue, not this change.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:35 - End-to-end string ordering tests and CI gate
- Added `runtime/eval_infix_string_order_test.go` (`TestStringOrderingEndToEndThroughExec`): a temp pack run through load, index, validate and `ExecRule` covering literal rule bodies, fact values (window check, fact vs list element), `when` gate, ternary condition, a `filter` lambda, and mixed string/number failing the rule (also inside a `when` gate). The unit test from the implementer already covered operator and byte-order edges, so none were repeated.
- Bite proof: with `runtime/eval_infix.go` restored to `main`, both the unit and end-to-end tests fail; the mixed-kind subtests pass on main by design (that behaviour is unchanged).
- Gate: `make test`, pinned lint (0 issues), gofmt, `make docs-graph`, no em dash all pass. No non-runtime Go package touched, so no extra race run.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-10-31 14:37 - Address review of string ordering
- Reviewer verdict CHANGES, one Warning: the PR description Testing notes listed only the unit test. Added the end-to-end test `TestStringOrderingEndToEndThroughExec` and what it covers (rule bodies, facts, `when` gate, ternary, `filter` lambda, mixed kinds failing the rule), a line that both tests fail with `runtime/eval_infix.go` restored to `main`, and a `### Tests` area.
- Took the suggestion to drop the two `pick` rows from the "literals and facts" table in `runtime/eval_infix_string_order_test.go`: they only asserted truthiness, and the ternary subtest checks `pick` by value. Reran the test, it passes.
- Declined the suggestion to fold the four relational cases into one comparison helper. That would rewrite the numeric `MustNumbers` lines, which the concurrent `box` numeric rework (#166) is changing; keeping a separate string pre-branch per case avoids a merge conflict and keeps numeric semantics (including NaN) byte-for-byte unchanged.
- _Model: claude-opus-5-5, role: implementer subagent_
