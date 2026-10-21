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
