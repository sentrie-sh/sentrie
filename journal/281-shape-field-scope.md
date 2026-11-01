## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #281
- A shape's alias and field types resolve in the calling policy's scope, so an exported shape whose field names a sibling shape fails at `exec` from another namespace while `sentrie validate` passes.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `271-derive-shape-scope`, which is not merged yet; the PR targets that branch.
- Plan: resolve a shape's alias and field types in the shape's own policy or namespace, however the shape was reached, using the same scope idea as `Derive.ShapeScope` from #271.
- _Model: claude-opus-5-5_
