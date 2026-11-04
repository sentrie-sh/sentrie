## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #300
- A `derive` that is the first body statement of a policy is rejected as a late header (`'derive' must appear before rules, exports, lets, and shapes`), so a policy-level derive is only accepted after a `let`, `rule` or `shape`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: a policy-level `derive` starts the body like a `let` does; facts and uses after it are still late headers.
- _Model: claude-opus-5-5_
