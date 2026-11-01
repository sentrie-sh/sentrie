---
id: index.derive_expr_walk
type: Function / Endpoint
language: Go
file_path: index/derive_expr_walk.go
tags: traversal, visitor, ast-walking, shared-utility
---

# Node: index.DeriveExprWalk (Shared Expression Traversal)

## 1. Architectural Role & Intent
The derive-analysis AST expression traversal (the reference cycle check uses the general [[index.expr_walk]] instead): `forEachDeriveExprChild` yields the direct children of any expression, and `walkDeriveExprDFS` builds a depth-first visitor on top of it. Despite the `derive` prefix it is the shared walker for three analysis passes - derive cycles, derive purity, and builtin kind checking. It owns no AST shape of its own: it delegates to [[index.expr_walk]] and adds only the derive restrictions (rule-body clauses rejected, a block may hold only lets, derive-named errors), so the set of expression forms it understands is exactly the general walker's.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.derive_expr_walk` | `DEPENDS_ON` | [[ast]] | Type-switches on the rejected clause kinds and on a block's statements. |
| `index.derive_expr_walk` | `CALLS` | [[index.expr_walk]] | `forEachExprChild` for the children of every expression kind. |
| [[index.derive_cycle]] | `CALLS` | [[index.derive_expr_walk]] | `walkDeriveExprDFS` to collect call targets. |
| [[index.derive_purity]] | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` for the generic recursion case. |
| [[index.builtin_check]] | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` as the default branch of the kind-check walk. |

## 3. Interface Contracts & Public Surface

- **Signature:** `walkDeriveExprDFS(root ast.Expression, visit func(ast.Expression) error) -> error`
  - **Behavior:** Depth-first pre-order: visits a node, then recurses into its children. Nil nodes are skipped. A visitor error aborts the whole walk.
  - **Side Effects:** None.
  - **Exceptions:** Propagates the visitor's error or the child enumerator's.

- **Signature:** `forEachDeriveExprChild(e ast.Expression, yield func(ast.Expression) error) -> error`
  - **Behavior:** Delegates to `forEachExprChild`, so children, order and nil handling (nil children are skipped) are the general walker's: a ternary yields only condition and else when `Elvis` is set; a block yields each let's value then the yield expression; map literals yield both key and value; comment wrappers yield their wrapped expression, making them transparent. Three derive-specific overrides sit on top: import, `with` and attachment clauses are rejected; a block's `let` is yielded as its value (the `VarDeclaration` itself is not an expression); any other block statement is rejected.
  - **Side Effects:** None.
  - **Exceptions:** `derive walk: unsupported statement %T` for a non-`VarDeclaration` inside a block (including a line comment); `derive walk: unsupported expression %T` for an import, `with` or attachment clause or an unrecognised type. An error from `yield` is returned unchanged.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless and reentrant.
- **Performance/Scale Notes:** Allocation-free apart from closures; cost is linear in AST size. It is invoked repeatedly by three passes over the same bodies, so a body is walked several times per validation.
- **Dependencies Risk:**
  - **Fails closed.** Any new `ast.Expression` type is an immediate error in every pass that uses this walker until [[index.expr_walk]] gains a case for it (a single place; `TestForEachDeriveExprChildCoversEveryASTNode` checks both walkers agree on every kind). That is a deliberate safety property for [[index.derive_purity]] - unknown constructs are never silently permitted.
  - **Blocks only tolerate `let`.** The override mirrors the purity rule, so this walker cannot traverse a block containing any other statement kind even in contexts where that would be legal. A line comment before a `let` or the yield is one such statement: [[index.derive_purity]] rejects a derive body containing one (known defect, left as is by the delegation change).
  - **No cycle protection.** `walkDeriveExprDFS` has no `seen` set; callers that might encounter a shared or cyclic AST must supply their own, as `walkDeriveExprSeen` does in [[index.derive_purity]].
  - **The `derive` prefix understates its reach.** Changing behaviour here to suit derive analysis silently changes rule-body builtin checking too.
