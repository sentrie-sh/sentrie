## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #115
- The decision endpoint reads request bodies without a size limit. Scope is the body size limit only; authentication, CORS, TLS, logging and traces moved to #315 (out of scope; the issue was rescoped to the limit), so the PR says `Closes #115`.
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Wrap the body in `http.MaxBytesReader`. Default limit 1 MiB, overridable on `serve` with `--http-max-body-bytes` (plain integer bytes); 0 means unlimited (user overrode the recommendation).
  - A body over the limit returns 413 with a problem document.
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#88 (draft, base `next`); Sentrie PR #328.
- _Model: claude-opus-5-5_

### 2026-11-10 22:33 - Implement the request body limit
- Reproduced on `main`: `handleDecision` decodes `r.Body` with a bare `json.NewDecoder`, no limit anywhere in `api/`. It is the only handler that reads a body; `/health` and `/health/deep` are GET and never touch it.
- `api/body_limit.go` (new): `DefaultMaxBodyBytes` (1 MiB), `WithMaxBodyBytes` (negative clamps to 0, 0 = no limit), `limitBody` (wraps `r.Body` in `http.MaxBytesReader`), plus `bodyOverLimit`, `bodyTooLarge`, `drainBody` and `writeBodyTooLarge` (413 through the existing `writeErrorResponse`, title `Request Body Too Large`, detail names the limit in bytes).
- `api/http.go`: route registration moved out of `Setup` into `routes()`, which returns the mux behind `limitBody`. The limiter sits in front of the whole mux so every route, including later ones, is covered, rather than inside the one handler. It cannot write the 413 itself because `writeErrorResponse` needs the request ID that `RequestIDMiddleware` adds per route (`GetRequestIDFromRequest` panics without it), so the reading handler maps the error.
- `api/handle_decision.go`: a declared `Content-Length` over the limit is a 413 before any read; a `*http.MaxBytesError` from `Decode` is a 413, not the old 400 `Invalid JSON`. After a successful decode the rest of the body is drained through the limiter: `json.Decoder` stops at the end of the first value, so without the drain a chunked body whose first value fits but whose total does not would be evaluated. Trailing data within the limit is still ignored, as before. With no limit nothing is drained.
- `cmd/serve.go`: `--http-max-body-bytes` (cling int input, default 1048576); a negative value is rejected by `validateMaxBodyBytes` before the pack loads.
- Tests: `api/handle_decision_test.go` `decisionHandler` now returns `api.routes()` (the served handler, limit included) instead of a hand-built mux, so the existing decision tests run behind the limit. Added `api/body_limit_test.go` (default, declared length, chunked, trailing data, exactly at limit, 0 unlimited) to drive the change; the trailing-data test was checked to fail with the drain disabled. The test agent owns the fuller regression set (serve flag parsing, negative value, end-to-end over a real listener).
- Codebase graph: `api.http`, `api.handle_decision`, `api.md`, `cmd.md` updated; the auth/CORS notes now point at #315.
- Gate: `make test` passed (`/tmp/sentrie-115-body-size-limit-impl-test.log`), `make docs-graph` passed. Manual: `serve --http-max-body-bytes=-1` errors; `serve --help` lists the flag with default 1048576.
- Signing down: nothing committed; queued script at `worktrees/v0.0.5-queued/115-body-size-limit.sh`.
- _Model: claude-opus-5-5_
