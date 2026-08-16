---
id: index.validate
type: Function / Endpoint
language: Go
file_path: index/validate.go
tags: validation, cycle-detection, dependency-graph, lifecycle
---

# Node: index.Validate (Validation Pipeline)

## 1. Architectural Role & Intent
The gate every policy pack must pass before execution. `Validate` runs six checks in a fixed order — identifier reference cycles, rule cycles, shape cycles, derive cycles, derive purity, and builtin call checking — then stores the rule and shape DAGs on the index and triggers [[index.commit]]. It is `sync.Once`-guarded, so an index is validated exactly once and the verdict is cached for its lifetime.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.validate` | `DEPENDS_ON` | [[dag]] | Builds three graphs: a per-policy `dag.G[String]` of identifiers, a global `dag.G[*Rule]`, and a global `dag.G[*Shape]`. |
| `index.validate` | `DEPENDS_ON` | [[ast]] | `addNodes` type-switches over expression forms to harvest identifier references. |
| `index.validate` | `DEPENDS_ON` | [[xerr]] | Emits `ErrInfiniteRecursion` and `ErrIndex`-wrapped cycle messages. |
| `index.validate` | `CALLS` | [[index.derive_cycle]] | Fourth check in the pipeline. |
| `index.validate` | `CALLS` | [[index.derive_purity]] | Fifth check. |
| `index.validate` | `CALLS` | [[index.builtin_check]] | Sixth check. |
| `index.validate` | `CALLS` | [[index.commit]] | Runs immediately after a successful validation, inside the same `Once`. |
| `index.validate` | `CALLS` | [[index.resolve]] | Resolves imported policies and base shapes while building edges. |
| `index.validate` | `MUTATES` | [[index.index]] | Writes `ruleDag`, `shapeDag`, `validated`, and `validationError`. |
| [[cmd]] | `CALLS` | [[index.validate]] | The `validate` command's whole purpose; `exec` calls it before evaluating. |
| [[runtime]] | `CALLS` | [[index.validate]] | Refuses to evaluate an index that has not validated. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).Validate(ctx) -> error`
  - **Behavior:** Runs `validate` once under `validationOnce`, wraps any failure as `validation error: …`, and on success proceeds to `Commit`. A commit failure is folded back into `validationError`, so a caller sees one error channel for both phases.
  - **Side Effects:** Populates the DAGs; hydrates shapes via commit.
  - **Exceptions:** Any check's failure, or a commit failure.

- **Signature:** `(*Index).IsValid(ctx) -> error`
  - **Behavior:** A pure alias for `Validate` — despite the predicate-style name it **performs** validation rather than reporting a cached boolean, and returns an error, not a bool.
  - **Side Effects:** Identical to `Validate` on first call.
  - **Exceptions:** Same.

- **Signature:** `(*Index).detectReferenceCycle(ctx) -> error`
  - **Behavior:** Per policy, builds a graph whose nodes are **bare identifier names** — rules and lets — and adds an edge for every identifier appearing in a rule's default, guard, or body, or in a let's initializer. Catches mutual recursion within a single policy.
  - **Side Effects:** None.
  - **Exceptions:** `xerr.ErrInfiniteRecursion` with the cycle path; `validation cancelled` on context cancellation.

- **Signature:** `addNodes(g, nodes, referedBy, policy)`
  - **Behavior:** The recursive harvester behind reference-cycle detection. Its `default` case **ignores** unrecognised nodes silently.
  - **Side Effects:** Mutates the graph.
  - **Exceptions:** None; edge-add errors are discarded.

- **Signature:** `(*Index).detectRuleCycle(ctx) -> (dag.G[*Rule], error)`
  - **Behavior:** Adds every rule as a node, then adds an edge only where a rule's body is an `ast.ImportClause`, resolving the target policy (defaulting to the current namespace for a single-segment path). Returns the graph for storage on the index.
  - **Side Effects:** None directly.
  - **Exceptions:** Policy resolution failures; `detected cyclic dependency in rules: a -> b -> a`.

- **Signature:** `(*Index).detectShapeCycle(ctx) -> (dag.G[*Shape], error)`
  - **Behavior:** Adds namespace and policy shapes as nodes, then adds a `with` edge for each. Namespace shapes resolve their base via `ResolveShape` (falling back to the current namespace when the reference has no parent); policy shapes look **only** in the enclosing namespace's map, keyed by the full FQN string.
  - **Side Effects:** None directly.
  - **Exceptions:** Shape resolution failures; `shape not found`; `detected cyclic dependencies in shapes: …`.

- **Signature:** `String` — a `string` alias with a `String()` method
  - **Behavior:** Adapts plain identifier names to the `dag` package's node constraint.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Exactly-once and sticky.** A failed validation is cached permanently; the index cannot be repaired and re-validated. Build a fresh one.
- **Performance/Scale Notes:** Six full traversals of the policy set, with a fresh graph allocated per policy in `detectReferenceCycle`. Context cancellation is checked at every nesting level, so a long validation aborts promptly. Every cycle detector reports only the **first** cycle found.
- **Dependencies Risk:**
  - **Reference-cycle detection works on unqualified names.** A rule and a fact in different policies sharing a name are distinct graph nodes only because the graph is rebuilt per policy — but within one policy, a let and a rule with related names can produce edges that do not correspond to real references, and edge-add errors are discarded rather than surfaced.
  - **The two shape-edge paths behave differently.** Namespace shapes resolve by parent-FQN plus last segment; policy shapes index `ns.Shapes` by the **whole `WithFQN` string**, which only matches when the reference is written as a bare name. A policy shape composing with a qualified path fails with `shape not found` even when the target exists.
  - **`IsValid` is a misleading name.** It mutates the index on first call and can take a long time; it is not a cheap status probe.
  - **Commit is inside the validation `Once`.** There is no way to validate without committing, and a commit failure is reported as a validation error.
  - **A large block of commented-out `containsSelfReference` code** remains in the file, superseded by `addNodes`. It is dead and should not be read as an alternative implementation.
