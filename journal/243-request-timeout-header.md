## Binaek Sarkar (binaek89@gmail.com)

### 23:30 - Pick up #243
- Branch cut from `225-evaluator-deadline`: it builds on #142's deadline plumbing in `serve`, and stacking on #225 avoids a conflict in the README "Evaluation Limits" section both edit.
- Design agreed with the maintainer in conversation:
  - header `Sentrie-Timeout` taking an RFC 3339 duration;
  - `--eval-max-timeout` as a ceiling that a request may go up to;
  - `--eval-timeout` stays the default.
- Calls made for the maintainer to confirm in review (listed in the issue):
  - a decimal fraction on seconds, which RFC 3339's grammar doesn't allow;
  - clamping over-ceiling requests rather than rejecting them;
  - startup errors for a ceiling below the default, and for a ceiling with an unbounded default.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 23:34 - Spec change to Go durations, and the plan
- Maintainer changed the header format before any code was written: `Sentrie-Timeout` takes a Go duration (`time.ParseDuration`: `250ms`, `2s`, `1m30s`), not an RFC 3339 duration. The RFC 3339 parser and the fraction-on-seconds extension are dropped and no custom parser is written. `Sentrie-Timeout-Applied` echoes `time.Duration.String()`. This supersedes the RFC 3339 wording in the issue and in the 23:30 entry above. Nothing parser-related was ever committed, so there is nothing to remove.
- Everything else stands: length limit before parsing, 400 for unparseable, zero or negative, clamping to `--eval-max-timeout`, the flag table, 503 on deadline, write timeout from the ceiling.
- One deadline per request. #142's review fix (ed2d94a on `142-eval-limits`, not yet under this stack) dropped the API's own `context.WithTimeout` so only the executor's deadline applies. This branch needs the same, because a request asking for more than the default would otherwise hit the API's default first. So:
  - `runtime.ContextWithEvalTimeout(ctx, d)` carries a per-evaluation deadline that `withEvalDeadline` uses instead of `--eval-timeout`. The executor still owns the deadline, the error still names the value used, and the `Executor` interface does not change.
  - `handleDecision` stops adding its own deadline (the same removal as ed2d94a) and puts the applied value on the context instead.
- Restack note: `api/handle_decision.go`, the `WithRequestTimeout` comment in `api/http.go` and `api/request_timeout_test.go` will conflict with ed2d94a and 65b594e when this stack is rebased onto the fixed #142. Take this branch's side; it already has ed2d94a's behaviour.
- _Model: claude-opus-5-5_

### 23:39 - Second spec change: reject over the ceiling, cap an unbounded default
- Maintainer decisions received mid-implementation, superseding the issue text and the 23:30 entry:
  - A `Sentrie-Timeout` above `--eval-max-timeout` is rejected with 400 (the detail names the ceiling) before evaluation. Nothing is clamped anywhere.
  - `--eval-timeout 0` with `--eval-max-timeout M` is valid, and a request without the header runs under M. This replaces the startup error for that row.
  - Still a startup error: a ceiling below a non-zero default. `Sentrie-Timeout-Applied` echoes the deadline used, default or requested.
- The first clamping version of `requestDeadline` was never committed; the committed code rejects from the start.

### 23:39 - Implementation
- `runtime`: `ContextWithEvalTimeout(ctx, d)`; `withEvalDeadline` uses that value in place of `e.evalTimeout` (shorter, longer, or when the executor's is disabled) and names it in the deadline error. The executor does not cap it, by design: the API is the one place that knows the ceiling. A context value rather than a new `Executor` parameter, so the interface and every caller stay as they are.
- `api/request_timeout.go`:
  - `parseTimeoutHeader` is the network boundary. Absent is fine; more than one header line is a 400; the value is length-checked (32 bytes; the longest Go duration string is 24) before `time.ParseDuration`, which itself rejects overflow; zero and negative are rejected. The value is quoted in an error only after the length check, and it is never logged.
  - `requestDeadline` returns the default (`defaultTimeout`: `--eval-timeout`, capped to the ceiling, so an unbounded default with a ceiling runs under it) or the header value, and a 400 error when the header is above `maxTimeout` (the ceiling, or else the default).
  - `handleDecision` drops its own `context.WithTimeout` (the same change as ed2d94a on `142-eval-limits`), checks the header after the method check and before path resolution, and sets the deadline with `ContextWithEvalTimeout` plus `Sentrie-Timeout-Applied` whenever one applies. The header is used for nothing else.
  - CORS: `Sentrie-Timeout` is added to `Access-Control-Allow-Headers` and `Sentrie-Timeout-Applied` to `Access-Control-Expose-Headers`, or a browser caller could not send or read them.
  - `writeTimeout` is `max(30s, ceiling + 5s)`, saturating so an absurd ceiling cannot overflow into a negative timeout.
- `cmd`: `--eval-max-timeout` (serve only, empty default meaning unset, Go duration like `--eval-timeout`). `parseEvalMaxTimeout` rejects garbage, an explicit zero or negative (unset is how you say "no ceiling"), and a ceiling below a non-zero default.
- Tests: table-driven `TestParseTimeoutHeader` and `TestRequestDeadlineCombinations` (every flag-table row, with and without the header) in `APITestSuite`; HTTP tests on the looping pack for shorter, longer-within-ceiling, over-ceiling 400, invalid 400s, the echoed default, the capped unbounded default, and CORS; `TestParseEvalMaxTimeout` in `CmdTestSuite`; `TestContextEvalTimeoutReplacesTheExecutorDeadline` in `RuntimeTestSuite`. `loopingPackAPI` now delegates to `loopingPackAPIWith(opts...)`.
- Docs: README "Evaluation Limits" (row, example, flag table, behaviour) and the `api`, `api.http`, `api.handle_decision`, `api.problem_details`, `cmd` and `runtime.executor` nodes. `make docs-graph` passes.
- Checks: `go test -race ./api/ ./cmd/`, `make test` and `make lint` (go1.25.0 golangci-lint) pass.
- _Model: claude-opus-5-5_
