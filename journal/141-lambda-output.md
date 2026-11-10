## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #141
- A decision whose output contains a lambda fails with "cannot marshal callable value to JSON". The issue's other two claims are already fixed on main.
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A lambda nested in output encodes as `{"kind":"callable"}`.
  - A rule whose whole value is a lambda is that rule's error.
  - Any other JSON encode failure: `exec --output json` reports it and exits non-zero; `serve` buffers the body and returns a 500 problem document instead of an empty 200.
- Docs PR: sentrie-sh/website#81 (branch `141-lambda-output` on `next`). Sentrie PR: #321.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 15:53 - Encode callables and fail lambda-valued rules
- Re-reproduced on the branch: `exec t/p/lam --output json` (a `filter` rule) exited 1 with "encoding output: ... cannot marshal callable value to JSON". The encode-error handling from the issue is already on main, so only the callable encoding was left. The lambda lives in the trace (the `filter` argument node's `result`), not in the decision value.
- `box.Value.MarshalJSON` now encodes through an unexported `jsonAny`, which is `Any` except that a callable becomes `{"kind":"callable"}`. Chose a single any-tree walk over marshalling `[]Value` recursively, so nested values are not re-compacted at every depth.
- `execRule` (`runtime/executor.go`) fails a rule whose reported value is a callable with an error wrapping the new `ErrCallableDecision` sentinel, returns no decision (reported as `unknown`) and skips attachments. This is the "rule's error" for now; #143 adds the per-rule `error` field on top of the same error. A lambda-valued rule reached through `import decision` fails the import the same way.
- An attachment that is a lambda is nested output and encodes as the placeholder, not an error.
- Adjusted `TestCallableRenderingAndMarshalBehavior` in `box/box_internal_test.go`, which asserted the old marshal error.
- Updated `box.value` and `runtime.executor` graph nodes; `make docs-graph` passes.
- `make build` installs to `~/.local/bin/sentrie`; built repro binaries with `go build -o` into the scratchpad instead after the first run.
- _Model: claude-opus-5-5_

### 2026-11-10 16:40 - Regression tests for #141
- Added tests for each decision: nested lambda in list, dict, deep dict and attachment encodes as the placeholder (box, runtime); a lambda only in the trace (a filter argument) no longer fails exec or serve; a rule whose whole value is a lambda errors in runtime (ErrCallableDecision), exec (non-zero, names the rule) and serve (500 problem document, detail names the rule); import decision of a lambda rule fails.
- The non-callable encode failure (number overflow) is already covered on main by TestExecJSONOutputSurfacesEncodingErrors and TestDecisionEncodingFailureIsAnError, so no new test there.
- Bite proof: with main's box/value.go the box, runtime, cmd and api placeholder tests fail; with main's runtime/executor.go (plus the sentinel var so it compiles) the whole-value and import tests fail in runtime, cmd and api.
- Dropped lang_test fixtures: those are parser-only fixtures, and runtime behaviour is covered by the suites.
- _Model: claude-sonnet-5-5_

### 2026-11-10 16:43 - Docs for #141
- Website branch 141-lambda-output (sentrie-sh/website#81): new "Lambdas in decision output" section in `reference/index.md`, plus notes in `cli-reference/exec.md` and `running-sentrie/serving-policies.md`. One commit, "Document lambdas in decision output".
- Checked against the branch binary: nested lambda value and filter-argument trace encode as the placeholder, whole-value lambda errors in exec (exit 1, single rule and whole policy), serve (500 problem+json "Evaluation Failed") and import decision. All validating snippets on reference/index.md pass; the calling example yields 10 and 11 for n=5. `npm run build` passes.
- Found: the website's policy-composition page writes `import decision of X from ...`, which the parser rejects (valid form is `import decision X from ...`). Not fixed here.
- _Model: claude-sonnet-5-5_

### 2026-11-10 17:59 - Review round 1 fix: callable check at the reporting boundary
- The round-1 reviewer found the whole-value callable check in `execRule`, which also serves in-policy rule references (`runtime/eval_ident.go`), so an unreported helper `rule helper = (x) => { yield x > 1 }` used as `filter([1, 2, 3], helper)` failed.
- Moved the check into `execRuleFrom`, after `execRule` returns: it covers `ExecRule` (and so `ExecPolicy`) and `import decision`. The output reports `unknown` with no attachments and the error on the rule node. Attachments are now evaluated before the check and then dropped; the decision only said the rule reports no attachments.
- Tests: `TestLambdaHelperRuleIsCallableFromAnotherRule` (fails with the round-1 check, passes now); `TestRuleWholeValueLambdaIsRuleError` also covers a lambda from a `default` branch and asserts decision `unknown`, no attachments and `RuleNode.Err`.
- Updated `runtime.executor` graph node (check moved from `execRule` to `ExecRule`/`execRuleFrom`); `make docs-graph` passes. Refreshed the PR description (tests by package, journal entries, review note on the placement).
- Website: `reference/index.md` says a rule whose whole value is a lambda is its error without saying this is only when the rule is reported; an unreported helper rule now works as a callback. Left for the docs writer.
- _Model: claude-opus-5-5_
