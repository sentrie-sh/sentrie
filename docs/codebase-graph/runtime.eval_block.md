---
id: runtime.eval_block
type: Function / Endpoint
language: Go
file_path: runtime/eval_block.go
tags: scoping, lazy-evaluation, blocks, lambda-bodies
---

# Node: runtime.evalBlock (Block Expression Evaluation)

## 1. Architectural Role & Intent
Evaluates `{ let a = … let b = … yield expr }`. It creates an attached child scope, registers every `let` **without evaluating it**, then evaluates the yield expression - so bindings are lazy and only those actually reached by the yield ever run. It is also the entry point for every lambda and derive body, since both are always block expressions.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_block` | `CALLS` | [[runtime.exec_ctx]] | `AttachedChildContext` for the scope, `InjectLet` per binding. |
| `runtime.eval_block` | `CALLS` | [[runtime.eval]] | Evaluates the yield expression. |
| `runtime.eval_block` | `CALLS` | [[runtime.trace]] | Opens an unnamed node; unsupported statements become `trace.IgnoredStmt` markers. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_block]] | `ast.BlockExpression` nodes dispatch here. |
| [[runtime.callable]] | `CALLS` | [[runtime.eval_block]] | Lambda bodies. |
| [[runtime.derive_invoke]] | `CALLS` | [[runtime.eval_block]] | Derive bodies. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalBlock(ctx, ec, exec, p, block *ast.BlockExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Creates an attached child context, walks the statement list registering `VarDeclaration` nodes as lets, skips comments, and records anything else as an ignored-statement trace marker. Then evaluates `block.Yield` in the child scope and returns its value.
  - **Side Effects:** Mutates the new child context; the yield's evaluation can do anything.
  - **Exceptions:** `xerr.ErrConflict` from `InjectLet` on a duplicate binding name; anything the yield raises.

## 4. Operational Context & Gotchas
- **Statefulness:** Creates and discards a scope per invocation. The child is attached, so it inherits the parent chain - and therefore inherits `evalDerive`, which is what keeps purity enforcement intact inside derive bodies.
- **Performance/Scale Notes:** Registration is O(statements) and allocation-light because initializers are not evaluated. The real cost is deferred to [[runtime.eval_ident]], which evaluates a let on first read and caches it. **An unreferenced `let` costs nothing.**
- **Dependencies Risk:**
  - **Lets are order-independent within a block.** All bindings are registered before the yield runs, so `let a = b` followed by `let b = 1` works - forward references are legal, unlike in most languages. Cycles are caught at read time by `PushRefStack`, not here.
  - **Unsupported statements are silently ignored.** The `default` branch attaches a `trace.IgnoredStmt` marker and continues. Nothing errors, and nothing surfaces unless someone inspects the trace - so a statement kind the parser allows in a block but this function does not handle just vanishes.
  - **`ec` is shadowed by the child.** The parameter is reassigned to the child context, so every reference after that line operates on the child. Readable, but easy to misread when scanning for which scope a call targets.
  - **The trace node is opened with an empty name**, unlike every other evaluator, so blocks appear as unlabelled nodes in the decision tree.
  - **`Dispose()` on the child is a no-op**, so the deferred call documents intent rather than releasing anything.
