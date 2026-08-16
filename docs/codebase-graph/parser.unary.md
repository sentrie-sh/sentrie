---
id: parser.unary
type: Function / Endpoint
language: Go
file_path: parser/unary.go
tags: unary-operators, negation, prefix-handler
---

# Node: parser.parseUnaryExpression (Prefix Operator)

## 1. Architectural Role & Intent
The prefix handler for `!`, `-`, `+`, and the `not` keyword. It parses its operand at `UNARY` precedence, which makes prefix operators bind tighter than every arithmetic and comparison operator but looser than calls and member access - so `not a.b(c)` negates the whole call result while `not a + b` negates only `a`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.unary` | `CALLS` | [[parser.expression]] | Parses the operand at `UNARY`. |
| `parser.unary` | `CALLS` | [[ast]] | Emits `ast.NewUnaryExpression(operatorString, operand, span)`. |
| [[parser.lookups]] | `CALLS` | [[parser.unary]] | Registered as the prefix handler for `TokenBang`, `TokenMinus`, `TokenPlus`, and `KeywordNot`. |
| [[parser.not]] | `CALLS` | [[parser.unary]] | Produces the same node type when wrapping a negated membership test. |
| [[parser.is]] | `CALLS` | [[parser.unary]] | Produces the same node type for `is not defined` / `is not empty`. |
| [[runtime.eval_unary]] | `DEPENDS_ON` | [[parser.unary]] | Applies numeric negation or [[trinary]] `Not()` depending on the operator and operand kind. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseUnaryExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Consumes the operator token **unconditionally** (no kind check - the dispatch table guarantees it), parses the operand at `UNARY`, and emits a `UnaryExpression` spanning operator to operand end. This is the one production in the expression set whose span is correct end-to-end.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` when the operand fails to parse; the underlying error was already recorded by the operand's handler.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable; stacked prefixes (`not not x`) recurse once per operator.
- **Dependencies Risk:**
  - **`!` and `not` are distinct tokens producing distinct operator strings**, so [[runtime.eval_unary]] must handle both spellings. They are semantically identical but not normalised at parse time.
  - **`not` is dual-registered.** As a prefix it lands here; as an infix it goes to [[parser.not]] for the `x not in […]` form. Which one fires depends purely on whether a left operand exists.
  - **Unary `+` is accepted and preserved.** It reaches the AST as a real node rather than being folded away, so evaluators and any AST-equality comparison must account for a semantically inert wrapper.
  - **No constant folding.** `-1` is a `UnaryExpression` wrapping an `IntegerLiteral`, not a negative literal - which is why negative numbers are rejected as constraint arguments in [[parser.literal]].
