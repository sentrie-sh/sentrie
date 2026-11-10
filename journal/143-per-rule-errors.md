## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #143
- One failing rule fails the whole policy, and output and error order is nondeterministic (runtime/executor.go:232-278).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A failing rule carries a per-rule `error` field and every other rule still evaluates.
  - `serve` returns 207 Multi-Status when any rule failed (including all of them), and 200 otherwise.
  - Output and error order stays nondeterministic: the user declined a total order, so that part of the issue is closed as won't fix.
  - `exec` exits non-zero when any rule errored; its output still carries every rule's result.
- Docs PR: sentrie-sh/website#79 (branch `143-per-rule-errors` on `next`). Sentrie PR: #319.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 18:29 - Per-rule errors, 207 and exec exit code
- Reproduced on the branch (stacked on 141-lambda-output): a policy with two healthy and two failing rules printed only the joined errors, in varying order, and exit 1; the healthy rules were lost.
- Shape: `ExecutorOutput` gains `Error *RuleError` (JSON `error`, omitted on success), next to the decision, which is then unknown. `RuleError` is `{"title": "Evaluation Failed", "status": <int>, "detail": "<message>"}`, problem-document style, and unwraps to the real error. `status` comes from the new `runtime.ErrorStatus` (400 invalid invocation, 503 deadline, 500 otherwise), which `api` now uses for its problem documents too, so there is one mapping. New file `runtime/rule_error.go`.
- `ExecPolicy` returns an output for every rule; when any rule failed its error is a `RuleErrors` (the rule errors, message one per line, unwrapping to each) next to the complete outputs. Chose a typed error over "nil error, look at outputs" so callers that treat any error as fatal (bench, the existing runtime tests) keep working and `errors.Is`/`errors.As` still see the rule errors. A non-`RuleErrors` error means the policy did not resolve, outputs nil.
- `ExecRule` now always returns an output on error (a rule failing before evaluation gets an unknown output), with `Error` set; `execRuleFrom` is unchanged for `import decision`.
- Data flow across goroutines: each rule's goroutine builds its own output, then takes `theLock` once to append the output and its `RuleError`; a recovered panic goes through the same report. Both slices are read only after `wg.Wait`. `facts`, `namespace`, `policy` are read-only shared, as before. `withEvalDeadline` now returns `(ctx, cancel, annotate)`: `ExecPolicy` applies the policy deadline's `annotate` to each rule's error (so each timed-out rule says "evaluation exceeded its <d> deadline") and cancels only after `wg.Wait`. In `execRuleFrom` the annotate defer runs before cancel, as the old `done` did. `go test -race` on runtime, api and cmd passes.
- `serve`: a policy request answers 207 with every rule when any rule failed (also when all failed), 200 otherwise. A named-rule request that fails and an unresolvable policy are still problem documents (400/503/500). So a missing required fact or a deadline on a policy request is now a 207 whose rules carry status 400 or 503; updated `TestHandleDecisionMissingRequiredFact` and `TestHandleDecisionEvaluationInternalError` accordingly.
- `exec`: for a policy it prints every rule (JSON `error` objects; table gets an `Errors:` section) and then fails with `<n> of <m> rules failed`, exit 1. A named rule's failure prints the error only, as before.
- Found that `main.go` printed `Error: ...` on stdout, which would have appended a non-JSON line to `exec --output json`. Moved it to stderr (separate commit) since #143 is what makes JSON and an error come together.
- Ordering stays nondeterministic (declined by the user). Errors inside a `default` expression being dropped is #148's, left alone.
- Note: a ctx deadline hit during an attachment of a lambda-valued rule reports `ErrCallableDecision`, not the timeout (from #141, deliberate, fails closed). It is now also what the rule's `error` shows.
- Tests: runtime `TestExecPolicyReportsEachRuleErrorAndKeepsTheOthers`, `TestRuleErrorJSONShape`, `TestExecPolicyMissingFactIsAnErrorOnEveryRule`, `TestExecPolicyDeadlineIsPerRuleError`, `TestExecRuleFailureCarriesTheRuleError`; updated `TestContributeRuntimeErrorFailsTheRuleAndIsReported` (it asserted the failing rules were dropped). api `TestDecisionPolicyWithAFailingRuleIsMultiStatus`; cmd `TestExecPolicyPrintsEveryRuleAndFailsOnARuleError`, `TestExecPolicyTableShowsRuleErrors`.
- Graph: runtime, runtime.executor, api.handle_decision, cmd, main nodes; `make docs-graph` passes. `make test`, `make lint` pass (`/tmp/sentrie-143-per-rule-errors-test.log`).
- Open for the user: whether a named-rule request (`/decision/ns/p/rule`, `exec ns/p/rule`) should also answer 207 / print the output with its error; implemented as unchanged problem document / error only.
- _Model: claude-opus-5-5_
