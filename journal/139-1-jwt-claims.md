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
