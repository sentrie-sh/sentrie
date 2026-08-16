---
id: runtime.eval_transform
type: Function / Endpoint
language: Go
file_path: runtime/eval_transform.go
tags: unimplemented, jq, data-transformation, stub
---

# Node: runtime.evalTransform (Transform Expression - Unimplemented)

## 1. Architectural Role & Intent
Placeholder for `transform <expr> with "<jq program>"`, the language's intended data-reshaping construct. **The runtime does not implement it.** The function opens a trace node and returns `xerr.ErrNotImplemented` without evaluating its argument, so the entire feature is present in the syntax and absent from the engine.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_transform` | `DEPENDS_ON` | [[xerr]] | Returns the shared `ErrNotImplemented` sentinel. |
| `runtime.eval_transform` | `CALLS` | [[runtime.trace]] | Opens a `transform` node recording the argument, then fails. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_transform]] | `ast.TransformExpression` nodes dispatch here and always fail. |
| [[parser.transform]] | `DEPENDS_ON` | [[runtime.eval_transform]] | The parser produces nodes this function cannot evaluate. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalTransform(ctx, _ *ExecutionContext, _ *executorImpl, _ *index.Policy, t *ast.TransformExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Always returns `(box.Undefined(), node, xerr.ErrNotImplemented)`.
  - **Side Effects:** Only the trace node.
  - **Exceptions:** `xerr.ErrNotImplemented`, unconditionally.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** None - it does no work.
- **Dependencies Risk:**
  - **The failure surfaces at the worst possible moment.** The lexer, [[parser.transform]], [[ast]], and [[index]] all accept the construct, so `sentrie validate` reports such a policy as **valid**. The error appears only when a decision is actually requested - in production, not at authoring or deploy time. Filed as [#109](https://github.com/sentrie-sh/sentrie/issues/109), proposing either implementation or an early rejection in the index.
  - **The signature discards the context, executor, and policy**, so the argument expression is never evaluated either. Any side effect the argument would have does not happen.
  - **The reference grammar advertises the feature.** Both `grammar/grammar.ebnf` and `grammar/grammar.peg` document `transform`, so the specification and the engine disagree.
  - **The transformer string is never validated.** It is stored as an opaque string literal by the parser and inspected by nothing, so a malformed jq program is indistinguishable from a well-formed one at every stage.
  - **Any agent or tool inferring language capability from the grammar or AST will conclude `transform` works.** This node exists primarily to record that it does not.
