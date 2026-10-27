# 139-2-jwt-asymmetric

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:48 - Start the asymmetric JWT layer
- Layer 2 of the #139 stack, cut from `139-1-jwt-claims` (claims, constant-time HMAC, `none` rejection). This layer adds RS256/384/512 and ES256/384/512 from a PEM or JWK public key, standard library only, and closes #139.
- _Model: claude-opus-5-5_

### 2026-10-21 14:50 - RS* and ES* verification from PEM or JWK
- Replaced layer 1's HMAC-only table with `jwtAlgorithms` (family, hash, curve) covering HS/RS/ES 256/384/512. PS* is left out: the issue asks for RS256 and ES256 at minimum, and RS*/ES* come for free from the same table.
- New `builtin_jwt_keys.go`: `parseJWTKey(alg, key)` returns a `[]byte` secret, `*rsa.PublicKey` or `*ecdsa.PublicKey`. Standard library only (`crypto/x509`, `encoding/pem`, `ecdsa.ParseUncompressedPublicKey` from Go 1.25, which also rejects off-curve JWK points).
  - PEM blocks: `PUBLIC KEY`, `RSA PUBLIC KEY`, `CERTIFICATE`. Private key blocks are refused with a message saying to pass the public key.
  - JWK: object or JSON string; RSA (`n`, `e`) and EC (`crv`, `x`, `y`); an optional `alg` must match and an optional `use` must be `sig`.
  - RSA moduli under 2048 bits are rejected (RFC 7518 section 3.3).
- Algorithm confusion: `decode` takes the algorithm from the attacker-controlled header, so the key's form must match the family. An HMAC secret that looks like a PEM block or a JWK (`-----BEGIN`, `{`) throws, and a plain string is never accepted as a public key. Tested with the classic forgery: HS256 signed using the RSA public key's PEM text as the secret.
- ECDSA signatures are JWS fixed-width R || S; a wrong length is just an invalid signature.
- `verify` validates the key against `algorithm` before looking at the token, so a key that does not suit the algorithm throws (argument error) rather than answering `false`.
- Tests generate RSA 2048 and P-256/384/521 keys per run and sign tokens in Go; no fixed asymmetric vectors are checked in.
- _Model: claude-opus-5-5_

### 2026-10-21 22:31 - Review fixes for PR #202
- Warning: the confusion guard was a prefix check after `TrimSpace`, and a PEM saved behind a byte-order mark got through: `decode` accepted HS256 tokens signed with that PEM text. I fixed the root cause, as review suggested. `decode` no longer lets the header choose the family: it accepts only `options.algorithms` when given, otherwise the key's own family (`jwtKeyAlgorithms`: HS* for a secret, RS* for an RSA key, the single ES* of an EC key's curve). `verify` already pins through its `algorithm` argument and rejects `options.algorithms`.
- I hardened detection too, so the two detectors cannot drift. `keyText` strips whitespace and a leading BOM for both `looksLikePublicKey`/`jwtPublicKey` and `hmacSecret`. `hmacSecret` refuses any secret that contains `-----BEGIN` anywhere, starts with `{`, or parses as a JSON object. A genuine RS256 token now verifies against the BOM-prefixed PEM.
- Suggestions taken (each trivial): a JWK with `d` is refused; the RSA modulus is capped at 8192 bits; any `crit` header fails verification (RFC 7515 4.1.11). Not taken: caching parsed keys (unmeasured) and the boxed-`String` message.
- Tests: the BOM forgery against `decode` and `verify`, a PEM embedded mid-secret, `algorithms` pinning, allowing and rejecting `none`, the EC curve default, a CERTIFICATE PEM round trip, an ASN.1 DER ES256 signature (`false`), `crit`, a private JWK and an over-size modulus.
- Carried from #197: `decode(rs256, undefined)` is fixed on `139-1-jwt-claims` (uncommitted there, like this). This layer does not have that change yet, so the RS-flavoured undefined-key test waits until the stack restacks.
- Signing refused the commit, so these changes are left uncommitted.
- _Model: claude-opus-5-5_

### 2026-10-21 22:33 - Review fixes committed
- Signing worked again on a later attempt, so the review fixes above are committed as be7814f ("Address review findings"). That supersedes the "left uncommitted" lines in the entry above. It is one commit rather than one per finding, because splitting it now would need an amend, and the review-fix rules forbid that.
- _Model: claude-opus-5-5_

### 13:59 - Restack onto main after #194, #197, #227 and #196 merged
- Rebased onto `main` (cfe249a). One conflict, in the module-set list of `docs/codebase-graph/runtime.js.builtins.md`: kept `main`'s pure-computation line (#208 folded base64 into `encoding`) and this layer's cryptographic line (HMAC, RSA and ECDSA).
- Expect a conflict with `154-js-boundary-values` in `builtin_jwt.go` and `builtin_clock.go` whichever lands second: 154 now reads the JWT clock through `executionStartSeconds`.
- _Model: claude-opus-5-5_
