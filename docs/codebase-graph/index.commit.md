---
id: index.commit
type: Function / Endpoint
language: Go
file_path: index/commit.go
tags: lifecycle, hydration, topological-order, finalization
---

# Node: index.Commit (Shape Hydration Finalization)

## 1. Architectural Role & Intent
The finalization step that flattens shape composition. It topologically sorts the shape DAG built by [[index.validate]] and calls `resolveDependency` on each shape in reverse sorted order (the graph's edges run from a shape to its `with` parent, so the reverse puts parents first), so every base shape is fully hydrated before anything composes with it. After commit, a shape's field map is complete and [[runtime]] can treat it as a flat record.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.commit` | `DEPENDS_ON` | [[dag]] | Calls `TopoSort` on the stored shape graph to derive the hydration order. |
| `index.commit` | `CALLS` | [[index.shape]] | Invokes `resolveDependency` per shape, which copies base fields into derived ones. |
| `index.commit` | `MUTATES` | [[index.index]] | Writes `committed` and `commitError`. |
| [[index.validate]] | `CALLS` | [[index.commit]] | Invoked automatically after a successful validation, inside the validation `Once`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).Commit(ctx) -> error`
  - **Behavior:** Runs `commit` once under `commitOnce`, wrapping failures as `commit error: …` and setting the `committed` flag. Idempotent by construction; the cached result is returned on every later call.
  - **Side Effects:** Mutates shape models throughout the index.
  - **Exceptions:** Topological sort failure; any `resolveDependency` error; `ctx.Err()`.

- **Signature:** `(*Index).commit(ctx) -> error`
  - **Behavior:** `shapeDag.TopoSort()` then a linear pass over the result **in reverse** (`slices.Backward`) calling `shape.resolveDependency(idx, nil)`, so a parent is hydrated before any shape that composes with it at any depth; checking context cancellation before each shape. Aborts on the first error, leaving earlier shapes already hydrated.
  - **Side Effects:** In-place field-map mutation via [[index.shape]].
  - **Exceptions:** As above.

## 4. Operational Context & Gotchas
- **Statefulness:** **Exactly-once, and partially destructive on failure.** Because it mutates shapes as it goes and aborts on the first error, a failed commit leaves the index in a half-hydrated state that cannot be retried - `commitOnce` has already fired.
- **Performance/Scale Notes:** One topological sort plus one pass per shape. The real cost is inside `resolveDependency`, whose cross-namespace fallback scans every namespace.
- **Dependencies Risk:**
  - **`resolveDependency` is called with a nil policy.** Commit passes `nil` for `inPolicy`, so the policy-local shape lookup inside `resolveDependency` is **never exercised from this path**. Policy-scoped shapes composing with another policy-scoped shape rely entirely on the namespace and cross-namespace lookups.
  - **It depends on `shapeDag` having been populated.** The graph is only assigned at the end of a successful `validate`, so calling `Commit` directly on an unvalidated index topologically sorts a nil graph rather than doing the right thing.
  - **Ordering correctness depends on the edge direction and on iterating `TopoSort` backwards.** `TopoSort` lists an edge's source first and the shape graph's edges point at the parent, so iterating forwards hydrates a child before its parent and a chain two `with` links deep loses the top parent's fields (fixed in #298). [[index.shape]] assumes its base is already hydrated and does not verify it, so any change to the sort's semantics silently produces partially-flattened shapes rather than an error.
  - **Commit is not separately reachable in the normal flow.** [[index.validate]] always calls it, and its error is folded into `validationError`, so a "validation error" reported to a user may in fact be a shape-composition failure.
