## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #284
- `forEachDeriveExprChild` duplicates the general `forEachExprChild` walker #269 adds, so the two can drift when an expression kind is added.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `269-cycle-check-walk`, which is not merged yet; the PR targets that branch.
- Plan: make the derive walker delegate to `forEachExprChild`, keeping only what differs for derives, and cover it with the walker coverage test.
- _Model: claude-opus-5-5_
