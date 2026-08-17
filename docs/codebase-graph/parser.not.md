---
id: parser.not
type: Function / Endpoint
language: Go
file_path: parser/not.go
tags: negated-membership, infix-handler, desugaring
---

# Node: parser.parseNotExpression (Negated Membership Infix)

## 1. Architectural Role & Intent
Handles the infix use of `not` - the negated membership and matching forms `x not in […]`, `x not contains y`, `x not matches "…"`, and `x not not y`. It desugars them rather than introducing new node types: the positive infix expression is built first and then wrapped in a `UnaryExpression`, so [[runtime.eval_infix]] and [[runtime.eval_unary]] need no special cases.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.not` | `CALLS` | [[parser.expression]] | Parses the right operand at the inherited precedence. |
| `parser.not` | `CALLS` | [[ast]] | Emits `InfixExpression` wrapped in `UnaryExpression`. |
| `parser.not` | `CALLS` | [[parser.parser]] | Uses `advance`, `canExpectAnyOf`, `head`, `errorf`. |
| [[parser.lookups]] | `CALLS` | [[parser.not]] | Registered as the **infix** handler for `KeywordNot`; the prefix registration goes to [[parser.unary]]. |
| [[parser.unary]] | `DEPENDS_ON` | [[parser.not]] | Shares the `UnaryExpression` node used as the negation wrapper. |
| [[runtime.eval_unary]] | `DEPENDS_ON` | [[parser.not]] | Evaluates the wrapper via [[trinary]] `Not()`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseNotExpression(ctx, parser, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `not`, requires one of `not`/`matches`/`contains`/`in` to follow, consumes that operator, parses the right operand at the inherited precedence, builds `Infix(left, op, right)`, and wraps it in `Unary("not", …)`. The span runs from the `not` token to the end of the right operand.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `expected 'not', 'matches', 'contains', or 'in' after 'not', got %s` when the follower is not one of the four; `nil` when the right operand fails.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Regression history.** A stray third `advance()` between the operator token and `parseExpression` swallowed the first token of the right operand, making negated membership forms unparseable (`9 not in [1, 2, 3]` failed with `unexpected token 'Comma'`). It is fixed and covered by `parser/not_test.go`; do not reintroduce a bare `advance()` between the operator and the operand parse.
  - **The `lang_test/` fixtures do not run.** `lang_test/0014-matches-contains-in.sentrie` exercises all three forms, but no Go test harness references that directory, so nothing executes those fixtures in CI (see #103). They provide no regression protection for this or any other production.
  - **Right-operand precedence is inherited, not raised.** The operand is parsed at the caller's precedence rather than the operator's, which differs from [[parser.infix]] and may produce different associativity for chained forms.
  - **The desugaring is invisible in the AST.** `x not in y` and `not (x in y)` produce identical trees, so tooling cannot recover which spelling the author used.
