---
id: index.expr_walk
type: Function / Endpoint
language: Go
file_path: index/expr_walk.go
tags: traversal, ast-walking, reference-cycle, fail-closed
---

# Node: index.ExprWalk (General Child Walker for Reference Analysis)

## 1. Architectural Role & Intent
`forEachExprChild` yields the direct children of any expression kind, and of the statements a block can hold (let, rule, comment). It exists so the reference cycle check in [[index.validate]] descends into every expression kind instead of keeping its own list of the kinds it remembers. Unlike [[index.derive_expr_walk]] it is not derive-specific: it descends into import clauses and their `with` values, and it accepts any statement a block holds.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.expr_walk` | `DEPENDS_ON` | [[ast]] | Switches over every expression kind plus the block statement kinds. |
| [[index.validate]] | `CALLS` | [[index.expr_walk]] | `addNodes` uses it for every node that is not an identifier, rule or let. |

## 3. Interface Contracts & Public Surface

- **Signature:** `forEachExprChild(n ast.Node, yield func(ast.Node) error) -> error`
  - **Behavior:** Yields each non-nil direct child. Leaves (identifiers, literals, pipeline holes, comment statements) yield nothing. A ternary yields only condition and else when `Elvis` is set; an `is` expression yields `Right` only when it is non-nil (a resolved shape operand leaves it nil); index access yields left and index; map literals yield both key and value; comment wrappers yield their wrapped expression; an import clause yields its `with` clauses and a `with` clause its value; a rule statement yields default, when and body. Type operands (cast target, `is` type) are not expressions and are not yielded.
  - **Side Effects:** None.
  - **Exceptions:** `reference walk: unsupported node %T` for a kind with no case; a yield error aborts the walk and is returned.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless and reentrant. It does not recurse itself; the caller recurses through `yield`, and carries any scope or referrer.
- **Fails closed.** An unknown node is an error, never a silent skip, and `TestForEachExprChildCoversEveryASTNode` reads the `ast` package source and fails when a new expression kind has no case here. A new expression kind therefore needs a case here and one in [[index.derive_expr_walk]].
- **No lambda or block scoping.** A lambda parameter or block let that shares a name with a rule still reads as a reference to that rule, as before; the walker only enumerates children.
