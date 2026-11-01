## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #281
- A shape's alias and field types resolve in the calling policy's scope, so an exported shape whose field names a sibling shape fails at `exec` from another namespace while `sentrie validate` passes.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `271-derive-shape-scope`, which is not merged yet; the PR targets that branch.
- Plan: resolve a shape's alias and field types in the shape's own policy or namespace, however the shape was reached, using the same scope idea as `Derive.ShapeScope` from #271.
- _Model: claude-opus-5-5_

### 12:58 - Resolve shape alias and field types in the shape's scope
- `index.Shape` gets an unexported `shapeScope` and `ShapeScope()`, mirroring `Derive.ShapeScope` from #271: the shape's policy, or a `Policy` holding only its namespace. `createShape` builds it at index time; the getter only falls back (allocating, never writing) for hand-built shapes, and returns nil when a hand-built shape has neither policy nor namespace, so the runtime then keeps the caller's `p`.
- Race-free for the same reason as #271: the scope is written once in `createShape` while the index is being built, before any executor exists, and never written again; evaluations only read the pointer and the maps it holds, which commit has finished mutating. No lazy cache, so there is no first-use write to race on.
- `ShapeModelField.DeclaredIn` records the declaring shape. `resolveDependency` copies field pointers into a `with` child, so a composed field keeps its base shape and `TypeScope` resolves it where it was written. Without this, `shape Line with com/example/catalog/Base` would check Base's `price: Price` in the shop's scope, the same bug one level down; checked by disabling `DeclaredIn` and watching the `composed` case fail.
- `validateAgainstShapeTypeRef` resolves the reference itself in `p`, then recurses into the alias and each field in the shape's scope. The constraints on the reference stay in `p`; constraints on a field type now evaluate in the shape's scope, which is where they were written.
- Static side: `index/builtin_kind.go` had the same bug. `lookupShapeKind`'s alias recursion and the field-access hops (`lookupShapeFieldTypeReadOnly`, `fieldTypeRefHop`, `resolveFieldAccessChain`) now carry the declaring shape's scope, so a caller's same-named shape no longer gives a false "filter: first argument must be a list". `index/is_resolve.go` needed nothing: it checks only the names written at the `is` site and never descends into a shape's fields.
- Left alone and noted in the graph: `resolveShapeReadOnly` matches a qualified reference's last segment against the site's local shapes first and skips the export check on its cross-namespace branch, so it can still pick a different shape than the runtime for a qualified name.
- Tests: runtime cross-namespace field, alias and `with` cases, with and without a caller `Price`; a cast error naming the sibling's missing field rather than "not found"; a policy-scoped shape still preferring its policy's shapes; an index test for the static kind. The runtime and index tests fail with the fix reverted (the policy-scoped one is a guard and passes either way).
- _Model: claude-opus-5-5_
