---
id: runtime.imports
type: Function / Endpoint
language: Go
file_path: runtime/imports.go
tags: cross-policy, composition, sandboxing, recursion
---

# Node: runtime.ImportDecision (Cross-Policy Decision Import)

## 1. Architectural Role & Intent
Implements `import decision <rule> from <ns/policy> with <fact> as <expr>`: it resolves the target policy, verifies the rule is exported, evaluates each `with` expression **in the importing policy's context**, marshals those values across the boundary, and re-enters the executor to run the target rule. The result is flattened into a dictionary envelope so the importing policy can read `state`, `value`, and any attachments as ordinary fields.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.imports` | `CALLS` | [[runtime.executor]] | Re-enters `ExecRule` with a freshly built fact map — a full nested execution. |
| `runtime.imports` | `CALLS` | [[runtime.eval]] | Evaluates each `with` expression in the **importing** policy's context. |
| `runtime.imports` | `DEPENDS_ON` | [[index]] | `ResolvePolicy` and `VerifyRuleExported` gate the import. |
| `runtime.imports` | `DEPENDS_ON` | [[box]] | `TryToBoundaryAny` marshals values out; the envelope is a `box.Dict`. |
| `runtime.imports` | `DEPENDS_ON` | [[trinary]] | Supplies `Unknown` for the empty envelope. |
| `runtime.imports` | `CALLS` | [[runtime.trace]] | Opens an `import` node and attaches the callee's whole rule trace. |
| [[runtime.eval]] | `CALLS` | [[runtime.imports]] | An `ast.ImportClause` used as a rule body dispatches here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ImportDecision(ctx, exec, ec, p, t *ast.ImportClause) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Requires a fully qualified `namespace/policy` source. Evaluates only those `with` clauses whose name matches a declared fact in the **target** policy — unmatched clauses are skipped without evaluation. Marshals each value through the boundary codec, then calls `ExecRule` on the target and wraps the output.
  - **Side Effects:** A complete nested rule execution, including the callee's own JS module binding and fact validation.
  - **Exceptions:** `import from must specify namespace/policy`; policy resolution and export-verification failures; `with` evaluation errors; `import with fact %q: …` marshalling errors; anything the nested `ExecRule` raises.

- **Signature:** `executorOutputEnvelope(output *ExecutorOutput) -> box.Value`
  - **Behavior:** Builds a dictionary of the attachments, then writes `state` and `value` **last** so an attachment named `state` or `value` cannot shadow them. A nil output yields `Unknown`/`Undefined`.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per call, but each invocation spins up an entirely separate execution context for the target policy.
- **Performance/Scale Notes:** An import is a **full nested execution** — target facts are re-validated, target `use` modules are re-bound (cache-hit after the first time), and a complete sub-trace is built and attached. A chain of imports multiplies this cost at every level.
- **Dependencies Risk:**
  - **`output.RuleNode` is dereferenced before the error check.** `ExecRule` returns `nil` output on several early failure paths — unresolvable policy, unexported rule, missing required fact — and this function calls `n.Attach(output.RuleNode)` **before** testing `err`. Any of those failures in an imported policy is a **nil-pointer panic**, not an error. This is the most reachable crash in the package: a missing required fact in the target policy is ordinary user error.
  - **Unmatched `with` clauses are silently dropped.** If the target policy renames or removes a fact, the corresponding `with` expression is never evaluated and no warning is produced — the import quietly runs with a default or fails as "required fact missing" instead of pointing at the stale clause.
  - **`with` names are matched against the target's alias-keyed `Facts` map**, so a `with` clause must use the target's **alias**, not its declared name.
  - **Recursion protection crosses the boundary via the fact map, not the context.** The nested `ExecRule` builds a brand-new execution context with an empty `refStack`, so import cycles are caught by [[index.validate]]'s static rule DAG rather than at runtime. A cycle that evades static detection would recurse until stack exhaustion.
  - **Dead branch:** the function rejects paths with fewer than two parts, then still tests `len(Parts) == 1` when computing the namespace. That branch is unreachable.
  - **The `with` expressions are evaluated with `ec.policy`, not the shadowed local `p`.** That is correct — values must come from the importer's scope — but the inner block shadows `p` with the *target* policy, which makes the code easy to misread.
