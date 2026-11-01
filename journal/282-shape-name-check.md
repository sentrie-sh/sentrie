## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #282
- `sentrie validate` does not check shape names in casts, typed lets, or derive and lambda signatures, so a missing shape there only fails when the expression is evaluated.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `273-constraint-arg-check`, which is not merged yet; the PR targets that branch.
- Plan: resolve every shape type reference at validation, in the scope the runtime uses, reusing the type position walk #273 added for constraint arguments, and report a missing or unexported shape at its position.
- _Model: claude-opus-5-5_
