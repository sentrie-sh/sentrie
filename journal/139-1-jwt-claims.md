# 139-1-jwt-claims

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:46 - Plan the #139 stack
- Stacks on `111-throw-builtin-errors`: the JWT rework relies on `throwErr` and the "verify answers false, everything else throws" convention from #111.
- Split per `.agentrules/stacked-prs.md`, because the whole of #139 (claims, constant-time HMAC, RS*/ES* with PEM and JWK, and their known-answer tests) runs past about 800 hand-written lines:
  - `139-1-jwt-claims` (this layer, `Updates #139`): constant-time `hmac.Equal` on decoded bytes, `exp`/`nbf` validated against the execution start by default with `leeway` and a `checkTime: false` opt-out, `none` rejected.
  - `139-2-jwt-asymmetric` (`Closes #139`): RS256/384/512 and ES256/384/512 from PEM or JWK, standard library only.
- _Model: claude-opus-5-5_
