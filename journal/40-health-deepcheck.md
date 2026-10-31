# 40-health-deepcheck

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #40
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: `serve` has only a bare `GET /health`; the issue specifies a healthcheck response (`check_id`, `status`, `time`) and a deepcheck with per-component status.
- Decision (default proposed to the user on 2026-10-31, not overridden): extend the existing health endpoint to the issue's healthcheck schema and add a deepcheck endpoint reporting the components the runtime can actually observe; fields the runtime cannot measure are left out rather than faked.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:35 - Implement healthcheck and deepcheck
- Done: `GET /health` now returns `{check_id, status, time}` (`check_id` is a fresh `crypto/rand.Text()` per response; `status` and `time` unchanged so existing callers and `cmd/serve_signal_test.go` keep working). New `GET /health/deep` reports `index`, `executor`, `js_runtime` and `pack` components plus an `errors` array. Handlers moved to `api/health.go`.
- Supporting changes outside `api/`, needed to observe state without changing it: `index.Status()` (new `index/status.go`), `Executor.BuiltinModules()` and `js.Registry.BuiltinModules()`. `Validate` and `Commit` now write the validated/committed flags and errors under a new `stateMu` so `Status` is race-free during a run (`go test -race` clean).
- Decision, path: `/health/deep`, because it sits beside the existing `/health` and keeps probes under one prefix; `/deepcheck` would add an unrelated top-level route.
- Decision, status code: 503 for an unhealthy deepcheck (200 when healthy), the usual contract for load balancers; `/health` stays 200 (liveness).
- Decision, `/health` stays unconditional: it is the liveness probe, and callers already rely on 200.
- Fields left out because the runtime cannot observe them (a made-up value would defeat the check): `executor.caches` (call memoize is an entry-count perch with no reservation or failure state, module binding is not a cache), `js_runtime.instance_pools` (the per-module VM limit is a constant with no availability state), `pack.location`, and the whole `operational` component (can_resolve_policy / can_create_context / can_load_modules would be restated index state or fabricated). `degraded` is never emitted.
- Left out for safety: `pack.location` is the operator-supplied path, often absolute, on an unauthenticated route; `pack.accessible` (an `os.Stat` of the pack directory) is reported instead, and its error text names no path. Added `pack.name` and `pack.version` (not in the issue) as harmless, useful identity.
- Index semantics: `validated` is false when validation failed (as in the issue's unhealthy example); a commit failure inside `Validate` is also stored as the validation error, so it is reported once, under `commit_error`. `statistics.shapes` counts namespace and policy-local shapes. `registered` equals `len(modules)` (real count, 15 today, not the issue's 13).
- Known limit: a closed executor still reports healthy because `Executor` exposes no closed flag.
- Problems: `make lint` panics in this environment ("file requires newer Go version go1.27", golangci-lint built with go1.26); not caused by this change. `go vet ./...` and gofmt are clean.
- Graph: new nodes `api.health`, `index.status`; updated `api`, `api.http`, `index`, `index.index`, `index.commit`, `runtime.executor`, `runtime.js.registry`, `overview`; `make docs-graph` passes.
- _Model: claude-sonnet-5-5, role: implementer subagent_

### 2026-10-31 14:39 - Test and CI pass
- Added tests on top of the implementer's: `api/health_test.go` (`TestDeepHealthCommitError`, `TestHealthRoutesRejectPost` for both routes, `TestDeepHealthConcurrentWithValidate`), `index/status_test.go` (`TestStatus_CommitFailure`), `cmd/serve_health_test.go` (`TestServeAnswersHealthAndDeepHealth`, real `serve` binary, checks nonce freshness, healthy deep check, no pack path in the body).
- Bite proof: reverting each production file to main (or removing it) fails the tests; reverting `index/validate.go` or `index/commit.go` alone fails the two race tests with DATA RACE.
- Bug found, not fixed (test code only here): `/health/deep` leaks the pack's absolute path through `validation_error`, `commit_error` and `errors[].message`, because index errors embed a source span that starts with the pack path (for example "cannot compose ... at /abs/pack/policy.sentrie:7:1-19"). The regression test is not committed; it needs a production fix (strip or relativize the pack location in the reported messages).
- Gate: make test, lint (CI-pinned v2.4.0, 0 issues), gofmt, docs-graph, race on api, index, runtime/js, cmd, em dash check all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-10-31 14:43 - Round 1 review fixes
- Fixed the path leak (Warning): `validation_error`, `commit_error` and the index `errors` entries are now the fixed strings `index validation failed` / `index commit failed`; the full error is logged at Debug through the api logger (`deepHealth` now takes the request context). No path rewriting: `Pack.Location` can be relative while spans are absolute, and `/var` vs `/private/var` differ. Added `api/health_leak_test.go` (queued with the api commit), updated the two tests that asserted on error text, made `api.health.md` true, and noted in the PR that detail stays withheld until #115 adds auth.
- Removed the legacy Apache boilerplate left under the SPDX header in `index/index.go` (Warning).
- Suggestion 2 taken: `registerHealthRoutes(mux)` used by `Setup` and by the test `healthMux`, so tests exercise the real routes.
- Suggestion 3 taken: PR Review notes state that `runtime.Executor` gained `BuiltinModules()`.
- Suggestion 1 declined: `operational` would only restate index validated/committed state under a second name, and its other issue fields (`can_create_context`, `can_load_modules`) cannot be observed; a partial component invites the reading that the rest was checked.
- Verified: `go test ./api ./index ./cmd`, `go test -race ./api ./index`, `make docs-graph` pass.
- _Model: claude-sonnet-5-5, role: implementer subagent_
