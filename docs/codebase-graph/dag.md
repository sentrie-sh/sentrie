---
id: dag
type: System / Package
language: Go
file_path: dag/
tags: graph-algorithms, cycle-detection, dependency-ordering, static-analysis
---

# Node: DAG (Generic Directed Acyclic Graph)

## 1. Architectural Role & Intent
`dag` is a small, dependency-free generic directed-acyclic-graph utility providing node/edge insertion, topological sorting, and cycle detection over any `fmt.Stringer` key type. It exists to serve one architectural need: [[index.package]] must order `derive` declarations by their dependencies and must reject policies whose derivations reference each other cyclically, before any evaluation is attempted. Keeping it generic and project-agnostic means the ordering logic is testable in isolation from policy semantics.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `dag` | `IMPORTS` | `std.errors`, `std.fmt`, `std.slices`, `std.strings`, `std.sync` | Zero internal or third-party dependencies; a graph sink. |
| [[index.package]] | `LAYERED_ON` | [[dag]] | Builds a graph of derive/rule dependencies and calls `TopoSort` to establish evaluation order. |
| [[index.derive]] | `CALLS` | [[dag]] | Registers each derive as a node and each referenced symbol as an edge. |
| [[index.derive_cycle]] | `CALLS` | [[dag]] | Uses `DetectFirstCycle` to produce a human-readable cyclic-dependency diagnostic. |

## 3. Interface Contracts & Public Surface

- **Signature:** `New[T fmt.Stringer]() -> G[T]`
  - **Behavior:** Constructs an empty graph. The `fmt.Stringer` bound is what allows cycle paths to be rendered as readable symbol chains in error messages.
  - **Side Effects:** Allocates internal adjacency storage.
  - **Exceptions:** None.

- **Signature:** `G.AddNode(n: T)`
  - **Behavior:** Registers a vertex. Idempotent — re-adding an existing node is a no-op, so callers need not track insertion.
  - **Side Effects:** Mutates internal graph state.
  - **Exceptions:** None.

- **Signature:** `G.AddEdge(from: T, to: T) -> error`
  - **Behavior:** Registers a directed dependency edge.
  - **Side Effects:** Mutates internal graph state.
  - **Exceptions:** Returns `ErrNodeMissing` if either endpoint was never added, and `ErrSelfLoop` if `from == to` — the latter catches the trivial "a derive references itself" case eagerly rather than deferring it to topological sort.

- **Signature:** `G.TopoSort() -> ([]T, error)`
  - **Behavior:** Returns vertices in dependency order, so that every node appears after everything it depends on. This is the evaluation order consumed by [[index.package]].
  - **Side Effects:** None (read-only traversal).
  - **Exceptions:** Returns `ErrNotADAG` when the graph contains a cycle.

- **Signature:** `G.DetectFirstCycle() -> []T`
  - **Behavior:** Returns the vertex path of the first cycle found, or nil/empty when the graph is acyclic. Used to turn a bare `ErrNotADAG` into a diagnostic naming the actual offending chain.
  - **Side Effects:** None.
  - **Exceptions:** None — reports absence of a cycle by returning an empty path rather than erroring.

- **Signature:** `ErrCycle` struct with `Path []string`
  - **Behavior:** Carries the rendered cycle path for presentation to the policy author.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Stateful and mutable.** A `G[T]` instance accumulates nodes and edges across calls and is intended to be built up once and then queried. `G` is an interface, so the concrete implementation is hidden and callers cannot inspect internals.
- **Performance/Scale Notes:** Graph sizes here are bounded by the number of derives and rules in a policy pack — typically tens to low hundreds — so asymptotic behaviour is not a practical concern. `TopoSort` and `DetectFirstCycle` are each a full traversal; calling both (the normal error path) means walking the graph twice.
- **Dependencies Risk:** No external failure domain. Two usage hazards: (1) `AddEdge` requires **both** endpoints to be pre-registered via `AddNode`, so a caller that adds edges while discovering symbols must add nodes first or handle `ErrNodeMissing` as "forward reference" rather than "unknown symbol"; (2) `TopoSort` reports only that a cycle exists — a caller that surfaces `ErrNotADAG` directly to the user gives an unactionable message, so it must follow up with `DetectFirstCycle` to name the participants.
