---
id: index.validate
type: Function / Endpoint
language: Go
file_path: index/validate.go
tags: validation, cycle-detection, dependency-graph, lifecycle
---

# Node: index.Validate (Validation Pipeline)

## 1. Architectural Role & Intent
The gate every policy pack must pass before execution. `Validate` first settles the right-hand side of every `x is Name` ([[index.is_resolve]]), then resolves every other shape name in a type ([[index.shape_refs]]), then runs thirteen checks in a fixed order - identifier reference cycles, self-references ([[index.self_reference]]), undefined identifiers ([[index.ident_check]]), contributions that read their own rule ([[index.contribute]]), rule cycles, unknown `with` fact names on imports, pack data names, shape cycles, derive cycles, derive purity, call targets, builtin call checking, and constraint arguments - then stores the rule and shape DAGs on the index and triggers [[index.commit]]. It is `sync.Once`-guarded, so an index is validated exactly once and the verdict is cached for its lifetime.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.validate` | `DEPENDS_ON` | [[dag]] | Builds three graphs: a per-policy `dag.G[String]` of identifiers, a global `dag.G[*Rule]`, and a global `dag.G[*Shape]`. |
| `index.validate` | `DEPENDS_ON` | [[ast]] | `addNodes` harvests identifier references; for an `IsExpression` it follows `Right` only when the index resolved it to a value. |
| `index.validate` | `DEPENDS_ON` | [[xerr]] | Emits `ErrInfiniteRecursion` and `ErrIndex`-wrapped cycle messages. |
| `index.validate` | `CALLS` | [[index.is_resolve]] | Runs before every check, so none of them reads a shape name on the right of `is` as a value reference. |
| `index.validate` | `CALLS` | [[index.expr_walk]] | `addNodes` descends into every expression kind through `forEachExprChild`, so a cycle cannot hide behind a kind the harvester forgot. |
| `index.validate` | `CALLS` | [[index.shape]] | `detectShapeCycle` takes each `with` edge from `resolveWithBase`, the resolution hydration uses. |
| `index.validate` | `CALLS` | [[index.shape_refs]] | Right after `is` resolution: every shape named in a fact, let, shape, derive or lambda signature, or cast must be visible from its site. |
| `index.validate` | `CALLS` | [[index.self_reference]] | Second check, right after reference cycles: a rule or let that reads itself, resolved by scope. |
| `index.validate` | `CALLS` | [[index.ident_check]] | Third check: every identifier evaluated as a value must name something at its site. |
| `index.validate` | `CALLS` | [[index.import_withs]] | Sixth check, right after `detectRuleCycle` has resolved every imported policy. |
| `index.validate` | `CALLS` | [[index.pack_data]] | Seventh check: every `data` declaration names a loaded `[data]` entry. |
| `index.validate` | `CALLS` | [[index.derive_cycle]] | Ninth check in the pipeline. |
| `index.validate` | `CALLS` | [[index.derive_purity]] | Tenth check. |
| `index.validate` | `CALLS` | [[index.call_check]] | Eleventh check: unknown callees, calls to non-callable values, let-lambda arity. |
| `index.validate` | `CALLS` | [[index.builtin_check]] | Twelfth check. |
| `index.validate` | `CALLS` | [[index.constraint_args]] | Thirteenth and last check: every constraint's literal arguments against its definition. |
| `index.validate` | `CALLS` | [[index.commit]] | Runs immediately after a successful validation, inside the same `Once`. |
| `index.validate` | `CALLS` | [[index.resolve]] | Resolves imported policies while building rule edges. |
| `index.validate` | `CALLS` | [[index.contribute]] | `detectContributionSelfReference` rejects a contribution that names its own rule; the fourth check, right after undefined identifiers. |
| `index.validate` | `MUTATES` | [[index.index]] | Writes `ruleDag`, `shapeDag`, `validated`, and `validationError`. |
| [[cmd]] | `CALLS` | [[index.validate]] | The `validate` command's whole purpose; `exec` calls it before evaluating. |
| [[runtime]] | `CALLS` | [[index.validate]] | Refuses to evaluate an index that has not validated. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).Validate(ctx) -> error`
  - **Behavior:** Runs `validate` once under `validationOnce`, wraps any failure as `validation error: …`, and on success proceeds to `Commit`. A commit failure is folded back into `validationError`, so a caller sees one error channel for both phases.
  - **Side Effects:** Populates the DAGs; hydrates shapes via commit.
  - **Exceptions:** Any check's failure, or a commit failure.

- **Signature:** `(*Index).IsValid(ctx) -> error`
  - **Behavior:** A pure alias for `Validate` - despite the predicate-style name it **performs** validation rather than reporting a cached boolean, and returns an error, not a bool.
  - **Side Effects:** Identical to `Validate` on first call.
  - **Exceptions:** Same.

- **Signature:** `(*Index).detectReferenceCycle(ctx) -> error`
  - **Behavior:** Per policy, builds a graph whose nodes are **bare identifier names** - rules and lets - and adds an edge for every identifier appearing in a rule's default, guard, or body, in the guard or body of any of the rule's contributions, or in a let's initializer, including inside index access, casts, transforms, comment wrappers, map keys and `import decision` `with` values and the body of a lambda passed to a call, but not the body of a lambda that is a let's whole value (`letValueNodes`), which is not evaluated when the let is. Catches mutual recursion between values within a single policy; recursion between let lambdas is allowed (#161) and bounded at runtime by the reference stack and the call and eval depth limits ([[runtime.exec_ctx]]). The graph ignores scope, so a block let still shares a node with a rule or let of the same name.
  - **Side Effects:** None.
  - **Exceptions:** `xerr.ErrInfiniteRecursion` with the cycle path; `validation cancelled` on context cancellation; an `xerr.ErrIndex` error if the walker meets an AST node kind it does not know.

- **Signature:** `addNodes(g, nodes, referedBy, policy) -> error`
  - **Behavior:** The recursive harvester behind reference-cycle detection. An identifier adds an edge; a rule or let declared in a body adds its node and is walked with itself as referrer, except that a let bound directly to a lambda contributes no edges; every other node, including a lambda passed to a call, is descended into through [[index.expr_walk]], which fails closed on an unknown kind. A shape operand of `is` is not a reference.
  - **Side Effects:** Mutates the graph.
  - **Exceptions:** The walker's unsupported-node error; edge-add errors are discarded, including `dag.ErrSelfLoop`: the graph ignores scope, so a self edge may come from a block let or parameter sharing the name. Real self-references are [[index.self_reference]]'s.

- **Signature:** `(*Index).detectRuleCycle(ctx) -> (dag.G[*Rule], error)`
  - **Behavior:** Adds every rule as a node, then adds an edge only where a rule's body is an `ast.ImportClause`, resolving the target policy (defaulting to the current namespace for a single-segment path). Returns the graph for storage on the index.
  - **Side Effects:** None directly.
  - **Exceptions:** Policy resolution failures; `detected cyclic dependency in rules: a -> b -> a`.

- **Signature:** `(*Index).detectShapeCycle(ctx) -> (dag.G[*Shape], error)`
  - **Behavior:** Adds namespace and policy shapes as nodes, then adds a `with` edge for each. Namespace and policy shapes alike get their edge from `addShapeBaseEdge`, which resolves the base with `resolveWithBase` ([[index.shape]]), the same resolution hydration uses.
  - **Side Effects:** None directly.
  - **Exceptions:** Any `resolveWithBase` error (not found, not exported, must be qualified); `error adding edge` (a shape composing with itself); `detected cyclic dependencies in shapes: …`.

- **Signature:** `String` - a `string` alias with a `String()` method
  - **Behavior:** Adapts plain identifier names to the `dag` package's node constraint.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Exactly-once and sticky.** A failed validation is cached permanently; the index cannot be repaired and re-validated. Build a fresh one.
- **Performance/Scale Notes:** Fifteen full traversals of the policy set (the `is` resolution pass, the shape name pass, then thirteen checks), with a fresh graph allocated per policy in `detectReferenceCycle`. Context cancellation is checked at every nesting level, so a long validation aborts promptly. Every cycle detector reports only the **first** cycle found.
- **Dependencies Risk:**
  - **Reference-cycle detection works on unqualified names.** A rule and a fact in different policies sharing a name are distinct graph nodes only because the graph is rebuilt per policy - but within one policy, a let and a rule with related names can produce edges that do not correspond to real references, and edge-add errors are discarded rather than surfaced. A self edge is the one case covered elsewhere, by the scoped [[index.self_reference]].
  - **`IsValid` is a misleading name.** It mutates the index on first call and can take a long time; it is not a cheap status probe.
  - **Commit is inside the validation `Once`.** There is no way to validate without committing, and a commit failure is reported as a validation error.
