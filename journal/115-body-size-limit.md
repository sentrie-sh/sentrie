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

### 2026-11-11 10:20 - Test the body limit
- Extended `api/body_limit_test.go`: trailing data within the limit is still ignored (PROVISIONAL decision), one byte over a declared length is 413, a chunked body exactly at the limit passes, a negative `WithMaxBodyBytes` clamps to unlimited, an oversized unparseable body is 413 not 400.
- `cmd/serve_test.go`: help lists the flag with default 1048576, `validateMaxBodyBytes` table, negative flag rejected before pack load (the pack location does not exist and the error still names the flag). New `cmd/serve_body_limit_test.go` drives the real binary: default 1 MiB, `--http-max-body-bytes 100`, and `0` (2 MiB body accepted). `startServe` now takes optional extra args.
- Bite proofs: drain disabled, `limitBody` removed, default removed, clamp removed, validator disabled, `WithMaxBodyBytes` dropped from serve, default changed in the flag: each fails the matching tests. Production `handle_decision.go` from `main` fails five tests. Disabling only the `bodyOverLimit` pre-read check changes nothing visible, since the reader yields the same 413; it is an early-out, not a decision.
- _Model: claude-sonnet-5-5_

### 2026-11-11 10:23 - Docs for #115 (sentrie-sh/website branch 115-body-size-limit)
- Documented the 1 MiB default body limit, `--http-max-body-bytes` (plain integer bytes, 0 unlimited, negative rejected at startup) and the 413 problem document on `cli-reference/serve.md`, `running-sentrie/serving-policies.md`, `guides/deployment-and-security.md`, plus a 413 mention in `deployment-operations/running-as-service.md` found by grep. Trailing data within the limit is stated as still ignored (PROVISIONAL).
- Verified against the branch binary with a limit of 100: small body 200, larger body 413 with the exact document quoted on the page; `-1` prints the startup error quoted on the page. No `sentrie` snippets added or edited.
- Pre-existing broken snippet recorded in the website PR description: the "Complete Example" pack on the serve page (old `schema_version`/`[engines]` keys, undeclared `user` fact).
- Website build passes (`/tmp/website-115-body-size-limit-build.log`). Four signed commits on the website branch.
- _Model: claude-sonnet-5-5_
