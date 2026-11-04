## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #298
- A shape composed with `with` two levels deep loses the top-level parent's fields, so `{"y":1,"z":1} is C` is `true` although `x` is required through `B with A`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: hydrate shapes parents first. `TopoSort` reverses a post-order that already lists parents first; fix the order (and check other `TopoSort` callers), with runtime tests for two- and three-deep chains.
- _Model: claude-opus-5-5_
