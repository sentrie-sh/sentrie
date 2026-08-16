---
id: index.derive_expr_walk
type: Function / Endpoint
language: Go
file_path: index/derive_expr_walk.go
tags: traversal, visitor, ast-walking, shared-utility
---

# Node: index.DeriveExprWalk (Shared Expression Traversal)

## 1. Architectural Role & Intent
The package's single generic AST expression traversal: `forEachDeriveExprChild` yields the direct children of any expression, and `walkDeriveExprDFS` builds a depth-first visitor on top of it. Despite the `derive` prefix it is the shared walker for all three analysis passes — derive cycles, derive purity, and builtin kind checking — and its exhaustive switch is the de-facto registry of which expression forms the index understands.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.derive_expr_walk` | `DEPENDS_ON` | [[ast]] | Switches over every concrete expression type and both comment wrapper forms. |
| [[index.derive_cycle]] | `CALLS` | [[index.derive_expr_walk]] | `walkDeriveExprDFS` to collect call targets. |
| [[index.derive_purity]] | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` for the generic recursion case. |
| [[index.builtin_check]] | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` as the default branch of the kind-check walk. |

## 3. Interface Contracts & Public Surface

- **Signature:** `walkDeriveExprDFS(root ast.Expression, visit func(ast.Expression) error) -> error`
  - **Behavior:** Depth-first pre-order: visits a node, then recurses into its children. Nil nodes are skipped. A visitor error aborts the whole walk.
  - **Side Effects:** None.
  - **Exceptions:** Propagates the visitor's error or the child enumerator's.

- **Signature:** `forEachDeriveExprChild(e ast.Expression, yield func(ast.Expression) error) -> error`
  - **Behavior:** Yields direct children per type. Leaves (identifiers, all literals, pipeline holes) yield nothing. Notable shapes: a ternary yields **only condition and else** when `Elvis` is set, matching how the parser aliases the abbreviated form; a block yields each let's value then the yield expression; map literals yield **both key and value**; comment wrappers yield their wrapped expression, making them transparent.
  - **Side Effects:** None.
  - **Exceptions:** `derive walk: unsupported statement %T` for a non-`VarDeclaration` inside a block; `derive walk: unsupported expression %T` for an unrecognised type.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless and reentrant.
- **Performance/Scale Notes:** Allocation-free apart from closures; cost is linear in AST size. It is invoked repeatedly by three passes over the same bodies, so a body is walked several times per validation.
- **Dependencies Risk:**
  - **The default case fails closed.** Any new `ast.Expression` type is an immediate error in every pass that uses this walker until it is added here. That is a deliberate safety property for [[index.derive_purity]] — unknown constructs are never silently permitted — but it means AST additions require a coordinated change.
  - **Blocks only tolerate `let`.** The statement switch mirrors the purity rule, so this walker cannot traverse a block containing any other statement kind even in contexts where that would be legal.
  - **No cycle protection.** `walkDeriveExprDFS` has no `seen` set; callers that might encounter a shared or cyclic AST must supply their own, as `walkDeriveExprSeen` does in [[index.derive_purity]].
  - **The `derive` prefix understates its reach.** Changing behaviour here to suit derive analysis silently changes rule-body builtin checking too.
