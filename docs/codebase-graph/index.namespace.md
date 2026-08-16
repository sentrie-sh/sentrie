---
id: index.namespace
type: Class
language: Go
file_path: index/namespace.go
tags: scoping, hierarchy, name-conflicts, visibility
---

# Node: index.Namespace (Scope and Name Registry)

## 1. Architectural Role & Intent
`Namespace` is the scope container holding a namespace's policies, shapes, derives, exports, and its parent/child links. Its defining responsibility is **cross-kind name uniqueness**: `checkNameAvailable` enforces that a single identifier cannot simultaneously name a policy, a shape, a child namespace, and a derive within one namespace — the invariant that makes identifier resolution in [[runtime]] unambiguous without a kind qualifier.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.namespace` | `DEPENDS_ON` | [[ast]] | Wraps `ast.NamespaceStatement` and uses `ast.FQN` hierarchy predicates. |
| `index.namespace` | `DEPENDS_ON` | [[xerr]] | Every collision is reported as `xerr.ErrConflict` carrying **both** conflicting spans. |
| [[index.index]] | `CALLS` | [[index.namespace]] | `ensureNamespace` calls `createNamespace` and `addChild`. |
| [[index.policy]] | `CALLS` | [[index.namespace]] | `addPolicy` registers each policy and checks name availability. |
| [[index.shape]] | `CALLS` | [[index.namespace]] | `addShape` registers namespace-level shapes. |
| [[index.derive]] | `CALLS` | [[index.namespace]] | `addDerive` and `addDeriveExport` register namespace derives. |
| [[index.resolve]] | `READS_FROM` | [[index.namespace]] | `VerifyShapeExported` / `VerifyDeriveExported` are declared on this type. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Namespace` struct — `{ Statement, FQN, Parent, Children, Policies, Shapes, Derives, ShapeExports, DeriveExports }`
  - **Behavior:** All collections are maps keyed by simple (unqualified) name, except `Children` which is a slice. Exports are separate maps from declarations, so a name may be declared without being exported.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `createNamespace(node *ast.NamespaceStatement) -> *Namespace`
  - **Behavior:** Allocates all maps with a nil parent and empty children.
  - **Side Effects:** Allocation only.
  - **Exceptions:** None.

- **Signature:** `(*Namespace).checkNameAvailable(name string) -> error`
  - **Behavior:** The shared uniqueness gate, checked in order against policies, shapes, child namespaces, and derives.
  - **Side Effects:** None.
  - **Exceptions:** `xerr.ErrConflict` naming the offending kind. Note the first span reported is the **namespace's own** span, not the incoming declaration's — see below.

- **Signature:** `(*Namespace).addChild(child *Namespace) -> error` / `IsChildOf(other) -> bool` / `IsParentOf(other) -> bool`
  - **Behavior:** Links a child after checking its base name, setting `child.Parent`. The predicates delegate to `ast.FQN`, which defines "child" as **exactly one segment deeper** — not any descendant.
  - **Side Effects:** `addChild` mutates both namespaces.
  - **Exceptions:** Conflict when the child's base name is already taken.

- **Signature:** `(*Namespace).addPolicy(p) -> error` / `addShape(s) -> error` / `addShapeExport(e) -> error` / `addDeriveExport(e) -> error`
  - **Behavior:** Register-with-conflict-check. `addPolicy` and `addShape` check `checkNameAvailable` **and** their own map; the export adders check only their own map.
  - **Side Effects:** Mutate the respective maps.
  - **Exceptions:** `xerr.ErrConflict` for `policy declaration`, `shape declaration`, `shape export`, `derive export`.

## 4. Operational Context & Gotchas
- **Statefulness:** Mutable during population, read-only afterwards. Not goroutine-safe; relies on the index's write lock.
- **Performance/Scale Notes:** `checkNameAvailable` iterates `Children` linearly while the other checks are map probes — irrelevant at realistic namespace widths.
- **Dependencies Risk:**
  - **Conflict spans are misleading.** `checkNameAvailable` passes `ns.Statement.Span()` as the first span, so a policy/shape name collision points at the **namespace declaration** rather than the offending declaration. The second span is correct. `addPolicy`/`addShape` re-check their own maps afterwards with correct spans, which is why some collisions read better than others.
  - **Child linking is one level only.** `IsChildOf` requires exactly one extra segment, so declaring `a/b/c` without `a/b` leaves `a/b/c` parentless — a hole in the tree that resolution does not repair.
  - **Order-dependent tree construction.** Parent/child wiring happens in `ensureNamespace` against namespaces known *at that moment*; a namespace created later still gets linked because the scan runs in both directions, but the base-name conflict check therefore also depends on insertion order.
  - **Export maps are unvalidated against declarations here.** `addShapeExport` accepts any name; only derive exports are checked against declarations, and that check lives in [[index.index]], not here.
