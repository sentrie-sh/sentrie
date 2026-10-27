# 139-1-jwt-claims

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:46 - Plan the #139 stack
- Stacks on `111-throw-builtin-errors`: the JWT rework relies on `throwErr` and the "verify answers false, everything else throws" convention from #111.
- Split per `.agentrules/stacked-prs.md`, because the whole of #139 (claims, constant-time HMAC, RS*/ES* with PEM and JWK, and their known-answer tests) runs past about 800 hand-written lines:
  - `139-1-jwt-claims` (this layer, `Updates #139`): constant-time `hmac.Equal` on decoded bytes, `exp`/`nbf` validated against the execution start by default with `leeway` and a `checkTime: false` opt-out, `none` rejected.
  - `139-2-jwt-asymmetric` (`Closes #139`): RS256/384/512 and ES256/384/512 from PEM or JWK, standard library only.
- _Model: claude-opus-5-5_

### 2026-10-21 14:47 - Constant-time HMAC, exp/nbf validation and none rejection
- Restructured `builtin_jwt.go` around `parseJWT` (header, payload, signing input, still-encoded signature), `verifyJWTSignature` and `checkJWTTime`, so layer 2 only adds key parsing and two signature families.
- Signature check: base64url-decode the signature and compare with `hmac.Equal` against the computed MAC. An undecodable signature is just "invalid", not a distinct error.
- Clock: `executionStartUnix` in the new `builtin_clock.go` reads the `ExecutionStartTimeUnixKey` global (Unix seconds, set by `ModuleBinding.Call`) and returns `errNoExecutionStart` when it is missing. No wall-clock fallback. It lives in its own file so #153 can reuse it for `time.now()` and #154 has one read path to change.
  - The clock is read before the token is parsed, so a VM without it throws on every call, not only once a token happens to verify.
- Time semantics (RFC 7519): valid while `now < exp + leeway`, and once `now + leeway >= nbf`. `exp == now` is expired. A present but non-numeric `exp`/`nbf` is invalid.
- Options are a third (`decode`) or fourth (`verify`) argument: `{ leeway?: seconds, checkTime?: boolean }`. Unknown keys throw so a typo such as `checktime` cannot silently disable the check. `checkTime: false` is the inspect-only opt-out and needs no clock.
- `decode` validates time claims even without a secret, since the issue asks for validation by default; `getPayload`/`getHeader` stay the inspect-only functions.
- `verify` keeps "token problems answer false" (#111 convention) but throws for `none`, an unsupported `algorithm`, an empty or non-string secret, and bad options. An empty secret used to make `decode` skip verification entirely (fail-open with an unset env secret); it now throws.
- `getPayload` now also requires a decodable header, since both go through `parseJWT`.
- Tests: jwt.io's published HS256 token as the known answer, plus generated HS256/384/512 tokens for expiry, `nbf`, leeway, opt-out, tampered signature and payload, wrong alg, `none` and bad arguments. The shared builtin test helper now pins the execution start like `ModuleBinding.Call` does.
- _Model: claude-opus-5-5_

### 2026-10-21 22:29 - Review fixes for PR #197
- Critical: `decode(token, undefined)` and `decode(token, null)` skipped verification, and `undefined` is exactly what an unset env secret or a missing fact produces. `decode` now requires a key (2 or 3 arguments) and throws on an absent one. I went further than the minimum, as review suggested: `decode(token)` with no key is gone too, because an unverified decode that still checks `exp` reads like verification. `getPayload`/`getHeader` are the inspect path. The redundant `if !isAbsent(key)` guard is left in place so the lines layer 2 edits stay untouched.
- Warning: a memoized `jwt.verify(...)!300` (or `now()!60`, `time.now()!60`) was cached in the executor-wide memo cache and outlived the execution whose clock produced it, so an expired token kept verifying. Fixed in `runtime/eval_call.go` rather than at index time (#168's nondeterminism metadata would be the other home): `isClockDependentCall` names the builtin `now`, `@sentrie/time.now` and `@sentrie/jwt.decode`/`verify`, and their memo key gets the execution start appended. Memoization still works within one execution. I checked that `TestMemoizedClockCallsDoNotCrossExecutions` fails with the key suffix disabled.
- Suggestion taken: `leeway` is capped at 86400 s, so a millisecond value (`3600000`) throws instead of disabling expiry. Not taken: a `requireExp` option (new API surface, left for a follow-up).
- Tests: forged and `none` tokens with an `undefined`/`null` key throw; a policy-level `jwt.decode(tok, cfg.secret)` with `cfg.secret` missing fails the rule; fractional `exp` either side of now and fractional leeway; leeway at and over the cap. Existing one-argument `decode` cases from #111 now pass a secret.
- Signing refused the commit (agent refused operation), so these changes are left uncommitted.
- _Model: claude-opus-5-5_
- Follow-up in the same session: the memo test now asserts `now()!60 == now()` in each of two executions rather than comparing values across them. That stays true whatever unit `now()` has, so it keeps passing once #153 switches `now()` to seconds. `clockDependentBuiltins` already lists `now_ms` so #153 needs no change in `eval_call.go`.

### 2026-10-21 22:33 - Review fixes committed
- Signing worked again on a later attempt, so the review fixes above are committed as f5c1f3c ("Address review findings"). That supersedes the "left uncommitted" lines in the entry above. It is one commit rather than one per finding, because splitting it now would need an amend, and the review-fix rules forbid that.
- _Model: claude-opus-5-5_
