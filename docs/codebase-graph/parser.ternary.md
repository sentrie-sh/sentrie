---
id: parser.ternary
type: Function / Endpoint
language: Go
file_path: parser/ternary.go
tags: conditional, elvis-operator, infix-handler, branching
---

# Node: parser.parseTernaryExpression (Conditional and Elvis)

## 1. Architectural Role & Intent
Parses the conditional operator in three shapes: the full `cond ? a : b`, the **elvis** form `cond ?: fallback` (yield the condition when it is truthy, otherwise the fallback), and an abbreviated `cond ? : b` where an omitted true-branch defaults to the condition itself. It is Sentrie's only expression-level branching construct, which makes it the primary way policies express "use this value unless it is missing" without a statement-level conditional.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.ternary` | `CALLS` | [[parser.expression]] | Parses branches; the elvis fallback is parsed at `COMPARISON`, the ordinary branches at the inherited precedence. |
| `parser.ternary` | `CALLS` | [[ast]] | Emits `ast.NewTernaryExpression` or `ast.NewTernaryElvis`. |
| `parser.ternary` | `CALLS` | [[parser.parser]] | Uses `expect(TokenQuestion)`, `canExpect(PunctColon)`, `advance`. |
| [[parser.lookups]] | `CALLS` | [[parser.ternary]] | Registered as the infix handler for `TokenQuestion` at `TERNARY` precedence. |
| [[runtime.eval_ternary]] | `DEPENDS_ON` | [[parser.ternary]] | Evaluates the condition through [[trinary]] and selects a branch. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseTernaryExpression(ctx, p, condition: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `?`, then branches:
    - **Elvis** — if `:` follows immediately, consumes it and parses one expression at `COMPARISON` precedence, emitting `NewTernaryElvis(condition, fallback)`.
    - **Full/abbreviated** — otherwise parses the true branch (defaulting to the condition itself when the next token is `:`), requires `:`, and parses the false branch, emitting `NewTernaryExpression`.

    The span starts at the condition and extends to the last branch — one of the few productions here with a fully correct range.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `?` or `:`, or when either branch fails to parse.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Asymmetric branch precedence.** The elvis fallback is parsed at `COMPARISON` while the ordinary branches use the inherited precedence. That means `a ?: b or c` and `a ? b : c or d` group their right-hand sides differently — the elvis fallback stops before `or`, the false branch does not.
  - **The abbreviated form aliases the condition node.** When the true branch is omitted, the **same** `ast.Expression` pointer is used for both `Condition` and `Then`. Any consumer that mutates or memoizes per-node will see one node in two positions, and the condition is evaluated once but referenced twice.
  - **Elvis and abbreviated forms are nearly identical in source but produce different nodes** (`NewTernaryElvis` vs `NewTernaryExpression`), so tooling must handle both.
  - **`?` is context-overloaded** — the conditional here, the optional-fact marker in [[parser.fact]], and the optional-field marker in [[parser.shape]].
