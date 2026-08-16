---
id: index.commit
type: Function / Endpoint
language: Go
file_path: index/commit.go
tags: lifecycle, hydration, topological-order, finalization
---

# Node: index.Commit (Shape Hydration Finalization)

## 1. Architectural Role & Intent
The finalization step that flattens shape composition. It topologically sorts the shape DAG built by [[index.validate]] and calls `resolveDependency` on each shape in dependency order, so every base shape is fully hydrated before anything composes with it. After commit, a shape's field map is complete and [[runtime]] can treat it as a flat record.

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
  - **Behavior:** `shapeDag.TopoSort()` then a linear pass calling `shape.resolveDependency(idx, nil)`, checking context cancellation before each shape. Aborts on the first error, leaving earlier shapes already hydrated.
  - **Side Effects:** In-place field-map mutation via [[index.shape]].
  - **Exceptions:** As above.

## 4. Operational Context & Gotchas
- **Statefulness:** **Exactly-once, and partially destructive on failure.** Because it mutates shapes as it goes and aborts on the first error, a failed commit leaves the index in a half-hydrated state that cannot be retried - `commitOnce` has already fired.
- **Performance/Scale Notes:** One topological sort plus one pass per shape. The real cost is inside `resolveDependency`, whose cross-namespace fallback scans every namespace.
- **Dependencies Risk:**
  - **`resolveDependency` is called with a nil policy.** Commit passes `nil` for `inPolicy`, so the policy-local shape lookup inside `resolveDependency` is **never exercised from this path**. Policy-scoped shapes composing with another policy-scoped shape rely entirely on the namespace and cross-namespace lookups.
  - **It depends on `shapeDag` having been populated.** The graph is only assigned at the end of a successful `validate`, so calling `Commit` directly on an unvalidated index topologically sorts a nil graph rather than doing the right thing.
  - **Ordering correctness is entirely delegated to `TopoSort`.** [[index.shape]] assumes its base is already hydrated and does not verify it, so any change to the sort's semantics silently produces partially-flattened shapes rather than an error.
  - **Commit is not separately reachable in the normal flow.** [[index.validate]] always calls it, and its error is folded into `validationError`, so a "validation error" reported to a user may in fact be a shape-composition failure.
