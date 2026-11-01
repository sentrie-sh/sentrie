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

### 13:12 - Test gaps for shape scope
- Added runtime tests: a Price reached through a field of a field (Order.item.price), list[Price], nullable Price? and each wrong element failing on its own; an alias of an exported shape from a third namespace; a shape that takes a catalog field in through `with` and declares a pricing field (each resolves in its own namespace, with and without same-named shop shapes); the same exported shape evaluated concurrently from two policies, one with its own Price (also run with -race, clean).
- Added an index test for the static kind of a field chain across namespaces (needs a shop shape of the same name to bite, otherwise an unresolved hop is just "unknown" and lenient).
- Each new runtime test fails with index/shape.go, index/builtin_kind.go and runtime/typeref_shape.go reverted to the base branch; the composed-field test also fails with `DeclaredIn` ignored; the index chain test fails when the hop scope is not carried. The static `with` assertion is a guard: it passes either way because the static lookup recurses into the parent with the parent's own scope.
- Ruled out as not this branch's bug, so no test kept: (1) a `with` chain two deep (C with B with A) drops A's fields at runtime, because `TopoSort` returns dependents first so C is hydrated before B; it is the same on the base branch and in one namespace. (2) a dict's element type (`dict[Price]`) is not enforced at runtime (`validateAgainstDictTypeRef` only runs constraints), so nothing about its scope can be observed.
- _Model: claude-sonnet-5-5_

### 13:15 - Paused at the plan limit
- Implementation and the test and CI pass are done and committed; the reviewer has not run yet. The batch paused at 93% of the 5-hour window to stay off paid usage, and resumes after the window resets with the opus reviewer for this branch.
- _Model: claude-opus-5-5_

### 17:40 - Review: describe the behaviour change for users
- Review (opus) found the code correct and race-free, with `with` fields resolved in their declaring shape. Its findings were all about the PR description.
- The PR description now says that packs which ran before can fail at `exec`: a namespace shape whose field names a policy-only shape, and a policy shape no longer shadowing the namespace shape a namespace shape's field names. It also says #292 (#282) adds the matching `validate` check, which already resolves shape field and alias types in the shape's own scope.
- Reworded the constraint-argument note: no result can change, since constraint arguments are literals and no checker reads the policy.
- Declined for now: a shared `(*Shape).ScopeOr` helper for the two "shape scope, else fallback" sites; a small tidy-up that can follow.
- _Model: claude-opus-5-5_
