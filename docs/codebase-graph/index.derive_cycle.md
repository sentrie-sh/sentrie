---
id: index.derive_cycle
type: Function / Endpoint
language: Go
file_path: index/derive_cycle.go
tags: cycle-detection, dependency-graph, validation, termination
---

# Node: index.detectDeriveCycle (Derive Recursion Guard)

## 1. Architectural Role & Intent
Builds a graph over every derive in `Index.DerivesByFQN`, adds an edge for each call one derive makes to another, and rejects any cycle. Because derives are pure functions that [[runtime]] inlines and evaluates eagerly, mutual or indirect recursion would not terminate — this check is what guarantees derive evaluation halts.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.derive_cycle` | `DEPENDS_ON` | [[dag]] | Uses `dag.New[*Derive]` and `DetectFirstCycle` to find the offending path. |
| `index.derive_cycle` | `DEPENDS_ON` | [[ast]] | Inspects `ast.CallExpression` callees and uses `ast.SlashCalleeFQNS` to flatten a slash-path callee into a string. |
| `index.derive_cycle` | `CALLS` | [[index.derive_expr_walk]] | `walkDeriveExprDFS` enumerates every expression in the derive body. |
| `index.derive_cycle` | `READS_FROM` | [[index.derive]] | Resolves callees against `DefineShort`, `DefineFQN`, then the global `DerivesByFQN`. |
| [[index.validate]] | `CALLS` | [[index.derive_cycle]] | Runs as the fourth check in the validation pipeline. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).detectDeriveCycle(ctx) -> error`
  - **Behavior:** Adds every derive as a node, then for each derive adds an edge to each resolved callee, skipping **self-edges**. Runs `DetectFirstCycle` and renders the path as FQNs joined by ` -> `.
  - **Side Effects:** None on the index; the graph is local and discarded.
  - **Exceptions:** `validation cancelled` on context cancellation; `derive dependency: …` when the graph rejects an edge; `cyclic derive dependency: a -> b -> a`, wrapped in `xerr.ErrIndex`.

- **Signature:** `deriveCallees(idx *Index, d *Derive) -> []*Derive`
  - **Behavior:** DFS over the lambda body collecting call targets by two routes — a bare identifier resolved through `DefineShort`, and a slash-qualified callee resolved through `DefineFQN` with a fallback to the index-wide `DerivesByFQN`. Unresolvable callees are silently omitted.
  - **Side Effects:** None.
  - **Exceptions:** None; the walk's error is discarded.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; builds and discards a throwaway graph per invocation.
- **Performance/Scale Notes:** One full body walk per derive plus graph construction — linear in total derive AST size. `DetectFirstCycle` stops at the first cycle, so error reporting is not exhaustive.
- **Dependencies Risk:**
  - **Self-recursion is silently allowed through this check.** The `callee == d` branch skips self-edges, so a directly self-recursive derive produces no cycle here. It is caught instead by [[index.derive_purity]] (via the derive's own name not being in scope) — meaning the diagnostic a user sees for `f = () => f()` is a scoping message, not a recursion message.
  - **Both resolution routes must agree with purity checking.** `deriveCallees` and `checkDeriveCall` in [[index.derive_purity]] independently re-implement callee resolution. If they drift, a call could be purity-approved but invisible to cycle detection, reintroducing non-termination.
  - **Unresolvable callees are dropped, not reported.** Cycle detection assumes purity checking already rejected unknown derives; run out of order, cycles through unknown names would be missed.
  - **The walk error is discarded**, so an unsupported expression type inside a derive body yields an incomplete callee list here rather than a failure.
