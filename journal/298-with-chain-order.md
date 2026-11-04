## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #298
- A shape composed with `with` two levels deep loses the top-level parent's fields, so `{"y":1,"z":1} is C` is `true` although `x` is required through `B with A`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: hydrate shapes parents first. `TopoSort` reverses a post-order that already lists parents first; fix the order (and check other `TopoSort` callers), with runtime tests for two- and three-deep chains.
- _Model: claude-opus-5-5_

### 2026-11-03 21:09 - Hydrate shapes parents first
- Cause confirmed: `dag.TopoSort` lists an edge's source before its destination (its own test pins A before B for A -> B). The shape graph's edges run shape -> `with` parent, so the sorted list has children first, and `index.commit` copied B's fields into C before B had A's.
- Fix at the call site: `index/commit.go` iterates the sorted list with `slices.Backward`. `TopoSort` keeps its meaning. Its only non-test caller is `commit`, so no other caller is affected; derive and rule graphs do not use `TopoSort`.
- Tests in `runtime/shape_with_chain_test.go`: depth 3 and 4 chains with several declaration orders (each field removed in turn must fail the `is` check), two branches sharing a grandparent (a shape has one `with` parent, so that is the diamond), and a three-namespace chain with `with ns/Shape` in both file orders. Run with the old forward iteration, every depth 3 and 4 case, the cross-namespace case and both c1_no_a and c2_no_a fail.
- Cross-namespace `with` needs the qualified form (`with com/base/A`); a bare name only resolves in the shape's own namespace at validate time.
- Noticed, not fixed: `commit` passes a nil policy to `resolveDependency`, and the shape graph's policy-shape edge lookup uses `ns.Shapes`, so policy-local shapes composing with policy-local shapes look untested. Outside this issue.
- Graph: updated `dag`, `index.commit`, `index.shape`; rewrote `index/commit.go` header to the SPDX form.
- _Model: claude-sonnet-5-5_

### 2026-11-03 21:13 - Tests for chain gaps
- Added `runtime/shape_with_chain_types_test.go`: field types inherited through a three-level chain are enforced at every level (wrong type at grandparent, parent, own level fails), an optional grandparent field stays optional and is type checked when present, in two declaration orders; and a policy-local shape `with` the top of a namespace-level chain gets every chain field.
- Added `index/shape_with_chain_test.go`: index-level checks of the hydrated field sets after `Validate` (depth 4, three orders, policy-local shape over a chain), and duplicate-field detection through a grandparent. The duplicate case was silently missed before the fix because the child never saw the grandparent's fields.
- Ran every new test with `index/commit.go` reverted to main: all fail (the one-level duplicate control passes on both, as intended); restored and diffed afterwards.
- Ruled out / found: a policy-local shape composing with another policy-local shape (`policy { shape A..; shape B with A.. }`) fails `Validate` with "shape not found: A", before any hydration: `detectShapeCycle` looks the parent up in `ns.Shapes` only (index/validate.go). Pre-existing and independent of this order fix, so no test was added; it needs its own issue.
- _Model: claude-sonnet-5-5_

### 2026-11-03 21:18 - Address review
- Added `TestShapeWithChainInheritedFieldKindIsChecked` in index/shape_with_chain_test.go: the static kind of a field reached through the chain is checked (`count(v.xs)` validates, `filter(v.n, ...)` fails). It is a regression guard and passes on main too, because static kinds resolve before commit through the recursive `lookupShapeFieldTypeReadOnly`.
- PR title now ends with `Closes #298`; the Review notes call out that a field redeclared two or more levels up is now rejected at load ("cannot compose with duplicate shape field").
- dag.md no longer says derives and rules are ordered with `TopoSort` (only the shape graph is); index.shape.md says the duplicate check covers any ancestor.
- Ruled out here: the qualified `with com/base/A` lookup resolving by last segment, local namespace first, is a separate pre-existing bug and gets its own issue.
- _Model: claude-opus-5-5_
