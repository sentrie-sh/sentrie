---
id: index.typeref_walk
type: Function / Endpoint
language: Go
file_path: index/typeref_walk.go
tags: static-analysis, type-refs, traversal, shapes
---

# Node: index.forEachTypeRef (Type Ref Position Walk)

## 1. Architectural Role & Intent
The one walk over every type ref position the index holds, shared by the checks that read types: [[index.shape_refs]] (shape names, #282) and [[index.constraint_args]] (constraint arguments, #273). Each type ref arrives with the scope its shape names resolve in, so a check that needs scope and one that does not see the same set of positions, and a new position is added in one place.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.typeref_walk` | `READS_FROM` | [[index.policy]] | Facts and policy lets carry type refs; their scope is the policy. |
| `index.typeref_walk` | `READS_FROM` | [[index.shape]] | Each `ShapeStatement`'s alias type and field types, once per statement, in the shape's own scope. |
| `index.typeref_walk` | `READS_FROM` | [[index.derive]] | A derive lambda's parameter and return types, in the derive's scope. |
| `index.typeref_walk` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` (`builtin_calls.go`) visits every expression: block lets, lambda types, casts and `is` types, with the expression's scope. |
| `index.typeref_walk` | `DEPENDS_ON` | [[ast.typeref]] | Unwraps `NullableTypeRef`; recurses into `ListTypeRef`, `DictTypeRef` and `RecordTypeRef`. |
| [[index.shape_refs]] | `CALLS` | [[index.typeref_walk]] | Resolves each `ShapeTypeRef` in its site's scope. |
| [[index.constraint_args]] | `CALLS` | [[index.typeref_walk]] | Checks each type ref's own constraint arguments. |
| [[index.is_resolve]] | `CALLS` | [[index.typeref_walk]] | `checkIsTypeShapes` expands an `is` type with `forEachNestedTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).forEachTypeRef(ctx, visit func(site typeRefSite, t ast.TypeRef)) -> error`
  - **Behavior:** Visits facts, policy lets, shape aliases and fields (deduplicated by statement), derive parameter and return types, and inside expressions block lets, lambda parameter and return types, casts and `x is T`. Each top-level type ref is expanded by `forEachNestedTypeRef`. `typeRefSite` holds the `policy` (nil at namespace scope) and `ns` that `LookupShape` takes, and `isType` for the type of an `is`. Scope: a shape's types resolve in the shape's policy or namespace, a derive's signature and body in its policy or, for a namespace-level derive, its namespace; everything else in the enclosing policy.
  - **Side Effects:** None.
  - **Exceptions:** `validation cancelled` on context cancellation (from `forEachScopedExpr`).

- **Signature:** `forEachNestedTypeRef(t ast.TypeRef, visit func(ast.TypeRef))`
  - **Behavior:** Visits `t` and, in pre-order, every list and set element, dict value and record field type in it. A nullable is never visited itself; its inner type is. A nil type ref is skipped.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One pass over the policy set plus one expression walk per caller.
- **Dependencies Risk:**
  - **A new type ref position or kind needs a case here.** A position the walk misses is silently unchecked by every caller; a new composite type ref kind needs a case in `forEachNestedTypeRef` for its children.
  - **Shape field scope is the shape's, not the caller's.** Until #281 (PR #291) lands, the runtime's `validateAgainstShapeTypeRef` ([[runtime.typeref_shape]]) validates a field type in the evaluating policy's scope, so the two can disagree for a namespace shape whose field names a shape only some policy declares. #281 makes the runtime use the shape's scope too; delete this note when it lands.
