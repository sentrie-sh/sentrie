# 139-2-jwt-asymmetric

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:48 - Start the asymmetric JWT layer
- Layer 2 of the #139 stack, cut from `139-1-jwt-claims` (claims, constant-time HMAC, `none` rejection). This layer adds RS256/384/512 and ES256/384/512 from a PEM or JWK public key, standard library only, and closes #139.
- _Model: claude-opus-5-5_
