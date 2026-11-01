## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #283
- The builtin kind check passes `d.Policy` (nil for a namespace derive) when resolving parameter types, so a shape-typed parameter of a namespace derive is unknown and the check is skipped.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: resolve a namespace derive's parameter shapes in `d.Namespace`, matching the runtime's derive scope.
- _Model: claude-opus-5-5_
