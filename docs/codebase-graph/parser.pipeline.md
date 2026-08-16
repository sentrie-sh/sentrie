---
id: parser.pipeline
type: Module / File
language: Go
file_path: parser/pipeline.go
tags: pipeline-operator, desugaring, ast-rewriting, hole-substitution
---

# Node: parser.pipeline (Pipeline Operator and Hole Substitution)

## 1. Architectural Role & Intent
`parser/pipeline.go` implements `|>` — the largest piece of **AST rewriting** in the front-end. A pipeline stage is not represented as a node; instead the left-hand value is folded into the right-hand call, either substituted at every explicit `#` hole or prepended as the first argument when no hole appears. The result is an ordinary `CallExpression`, so [[runtime.eval_call]] never learns that pipelines exist. The file also enforces a deliberately narrow right-hand-side grammar to keep that rewrite unambiguous.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.pipeline` | `CALLS` | [[parser.expression]] | Parses the right-hand side at `PIPELINE` precedence. |
| `parser.pipeline` | `CALLS` | [[ast]] | Reconstructs nodes during substitution and emits the final `CallExpression`. |
| `parser.pipeline` | `DEPENDS_ON` | [[parser.primary]] | Consumes the `PipelineHoleExpression` nodes that `#` produces. |
| [[parser.lookups]] | `CALLS` | [[parser.pipeline]] | Registered as the infix handler for `TokenPipeForward`. |
| [[parser.precedence]] | `DEPENDS_ON` | [[parser.pipeline]] | `PIPELINE` is the loosest real binding power, so a stage absorbs everything to its right. |
| [[runtime.eval_call]] | `DEPENDS_ON` | [[ast]] | Evaluates the rewritten call with no knowledge of the original pipeline syntax. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parsePipelineExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `|>`, records whether the right-hand side starts with `(`, parses it, then validates: the RHS must be a `CallExpression` whose callee has an identifier root (a bare identifier or a chain of field accesses over one), and grouped expressions are rejected outright. If any argument subtree contains a `#`, **every** hole is replaced by the left-hand expression; otherwise the left-hand expression is prepended as the first argument. The emitted call preserves the RHS's memoization flag and TTL, and spans from the left operand to the right.
  - **Side Effects:** Consumes tokens; allocates a rewritten argument tree.
  - **Exceptions:** `invalid pipeline target: missing left-hand side expression`; `invalid pipeline target: grouped expressions are not allowed on the right-hand side`; `invalid pipeline target: right-hand side must be a call on an identifier or module-qualified field access` (raised for both a non-call RHS and a non-identifier-rooted callee).

- **Signature:** `hasIdentifierRoot(expr: ast.Expression) -> bool`
  - **Behavior:** Recursively unwraps `FieldAccessExpression` to check that a callee bottoms out in an `Identifier`. This is what permits `x |> mod.fn()` while rejecting `x |> arr[0]()`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `containsPipelineHole(expr) -> bool` / `containsPipelineHoleInExprs(exprs) -> bool`
  - **Behavior:** Exhaustive structural search for `#` across every composite expression kind — calls, accesses, list/map literals, infix, unary, ternary, cast, `is defined`/`is empty`, transform, and both comment wrappers.
  - **Side Effects:** None.
  - **Exceptions:** None; an unhandled node type returns `false`.

- **Signature:** `substitutePipelineHoles(expr, replacement) -> ast.Expression`
  - **Behavior:** The mirror of the search: rebuilds the tree with each `#` replaced. Every branch **reconstructs** the node via its `ast.New*` constructor rather than mutating, preserving the original spans.
  - **Side Effects:** Allocates new nodes; the replacement expression is **shared** by every hole it fills.
  - **Exceptions:** None; unhandled node types are returned unchanged.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless, but the rewrite means the AST no longer reflects the source shape.
- **Performance/Scale Notes:** Two full traversals of the argument subtree per stage (detect, then substitute), plus reallocation of every node on the substitution path. Long pipelines with large literal arguments pay this per stage.
- **Dependencies Risk:**
  - **The two helper switches must stay in lockstep.** `containsPipelineHole` and `substitutePipelineHoles` enumerate node kinds independently. Adding an expression type to one and not the other yields a silent bug: a hole that is detected but never substituted (leaving a bare `#` for the runtime) or, worse, one that is never detected and so triggers the prepend path instead. Any new `ast.Expression` requires editing both.
  - **A multi-hole stage duplicates the left expression by reference.** `x() |> f(#, #)` puts the *same* node in both argument positions, so the left-hand side is evaluated twice unless the runtime memoizes — a real semantic consequence of the desugaring.
  - **The grouped-RHS check runs after parsing, not before.** The flag is captured from the head token but the error is raised only after the whole RHS is parsed, so the diagnostic's position is past the offending parenthesis.
  - **Pipelines are invisible after parsing.** Nothing downstream can tell `f(x)` from `x |> f()`, so error messages about argument counts refer to the rewritten call and may not match what the author wrote.
  - **`|>` binds loosest**, so `x |> f(#) or y` puts the `or` *inside* the stage. See [[parser.precedence]].
