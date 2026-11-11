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

### 2026-11-10 18:34 - Fact errors are request errors
- PROVISIONAL (user AFK) decisions from the orchestrator, to be confirmed:
  - PROVISIONAL (user AFK): a failing named rule (`/decision/ns/p/rule`, `exec ns/p/rule`) gives a problem document or an error, not a 207.
  - PROVISIONAL (user AFK): fact errors are request errors, not per-rule failures. A missing required fact, or any other fact validation error found before rules evaluate, returns a 400 problem document for a policy request, as before (consistent with #147's "wrongly typed fact -> 400"). A deadline or any other evaluation-time failure stays per rule and counts toward the 207.
- Chose classification over hoisting: fact binding and type validation stay per rule (each rule has its own execution context, and hoisting would evaluate fact defaults twice or move validation out of the rule's trace node). `execRuleFrom` marks a top-level call's error as an unexported `requestError` when it came from the fact-binding loop (`bindingFacts`) or from a `factError` that `execRule`'s type validation raised for that same context. An imported policy's fact error has a different context, so it stays the importing rule's failure. A context error is never a request error. `ExecPolicy` returns the first request error alone, with nil outputs; facts are the policy's, so every rule has the same one.
- Today the type-validation error maps to 500 (not invalid invocation) until #147 changes it; it is a problem document either way.
- Tests: `TestHandleDecisionMissingRequiredFact` restored to 400. Runtime `TestExecPolicyMissingFactIsARequestError` (replaces the per-rule version) and `TestExecPolicyFactTypeErrorIsARequestError` (also covers the import case); cmd `TestExecPolicyMissingFactIsAPlainError`.
- Commit a0b04cb. `make test`, `make lint`, `make docs-graph`, race on runtime, api and cmd pass (`/tmp/sentrie-143-per-rule-errors-test2.log`).
- _Model: claude-opus-5-5_

### 2026-11-10 18:37 - Tests for per-rule errors

- Added to the implementer's tests: runtime `TestExecPolicyAllAndNoneFailing`; api all-fail 207, none-fail 200 without an `error` field, mixed set (no order asserted), named-rule problem document, fact errors (400 request-level, imported fact error stays a rule failure), deadline as per-rule 503 inside a 207; cmd all-fail "2 of 2 rules failed", named-rule plain error, and a built binary test (exit 1, `Error:` on stderr, JSON alone on stdout).
- Bite proofs against 141-lambda-output copies: api/handle_decision.go, cmd/exec.go, main.go fail the new tests. runtime/executor.go swap fails to compile (rule_error.go needs ExecutorOutput.Error), so a mutation dropping the request-level return was used: both fact-error tests fail.
- Commit signing was refused by the agent, so the test commit and this entry are queued.
- _Model: claude-sonnet-5-5_

### 2026-11-10 18:40 - Docs for #143 (website#79)

- Documented per-rule `error` objects, 207 Multi-Status on policy requests, fact errors as 400 problem documents, exec printing every rule then exiting 1 with "<n> of <m> rules failed", the `Error:` line on stderr, and that output order is not guaranteed. Pages: reference/rules.md, reference/errors.md, running-sentrie/executing-policies.md, running-sentrie/serving-policies.md, cli-reference/exec.md, cli-reference/serve.md, deployment-operations/running-as-service.md, troubleshooting/common-errors.md.
- Ran the new `access` policy with validate, exec (table and JSON) and serve (207, 200, 400, single-rule 500) against the branch binary: all as documented. The named-rule problem document and the fact-error 400 are marked PROVISIONAL (user AFK), needs confirmation in the website PR description.
- Candidate bug: `yield 100 / user.limit` fails with "right operand is not a number" (likely parses as `(100 / user).limit`); `100 / (user.limit)` works.
- _Model: claude-sonnet-5-5_

### 2026-11-10 22:57 - Docs round 2 for #143
- Reviewer found the docs claimed a missing or invalid fact on a policy request is a 400. Checked the branch binary with `serve`: a missing required fact is a 400 problem document; a wrongly typed fact is a 500 problem document today.
- Fixed `serving-policies.md`, `reference/errors.md` and `running-as-service.md` (and the website PR description) to say missing -> 400, wrong type -> 500 for now (400 after sentrie-sh/sentrie#147). Sentrie PR description line on fact errors and the API test comment corrected the same way by the orchestrator.
- Website build passes, no em dash. Left uncommitted (signing is down), one queued website commit.
- _Model: claude-sonnet-5-5_

### 2026-11-12 02:27 - Review fix: a failed attachment leaked the decision
- Review comment on `runtime/rule_error.go:94` (P2): `withRuleError` only built an unknown decision when the output was nil, so `rule ok = true` exported with `attach boom as 1 / 0` returned `RuleErrors` and an `error` field next to `state: "true"`, `value: true`. Reproduced on ExecPolicy and ExecRule (runtime), the 207 body (api) and `sentrie exec` JSON (cmd); the partial attachment `good` leaked too.
- Fix: `withRuleError` now sets the unknown decision (`DecisionOf(box.Trinary(trinary.Unknown)).reported()`, as the nil branch did) and nil attachments on every error, existing output or not. Attachments on failure are dropped: a failed output carries identity, the unknown decision, the error and the trace. Follows #141's lambda-valued rule (no attachments) and the website example (`"attachments": null` on a failed rule). The trace is kept as diagnostics, though it records the values computed before the failure.
- Checked other paths: `withRuleError` is the only place an output's `Error` is set (ExecPolicy's `report` and ExecRule both go through it); `api/health.go` `p.Error` is a health probe, unrelated; `cmd/exec.go` sets nothing. `runtime/imports.go` discards the output on error.
- Tests: runtime `TestAttachmentFailureReportsUnknown`, api `TestDecisionAttachmentFailureIsUnknown`, cmd `TestExecPolicyAttachmentFailureIsUnknown`; each failed against the old `rule_error.go`. No existing test pinned the leak. Graph: runtime, runtime.executor, api.handle_decision nodes.
- Website (read only): reference/rules.md, reference/errors.md and serving-policies.md already state a failed rule's decision is unknown, and their examples show `"attachments": null`; none mention attachment failures, so nothing there becomes false.
- `make test`, `make lint`, `make docs-graph` pass (`/tmp/sentrie-143-per-rule-errors-c1-*.log`).
- _Model: claude-opus-5-5_

### 2026-11-12 02:29 - Tests for the attachment failure fix

- Read the implementer's three tests (runtime, api, cmd). They already cover: two attachments where the second fails (neither appears, all three layers), a failing rule next to a sibling with a good attachment (sibling keeps decision and attachment), and a named-rule HTTP request whose attachment fails (problem document, no decisions or attachments). Added nothing for those.
- Added runtime `TestBodyFailureWithAttachmentReportsUnknown`: a rule body failure with an attachment declared is still unknown with no attachments and the sibling is intact. It passes against the old `rule_error.go` too, by design: it pins unchanged behaviour. The three implementer tests fail against the old file (checked).
- _Model: claude-sonnet-5-5_
