---
id: parser.infix
type: Function / Endpoint
language: Go
file_path: parser/infix.go
tags: binary-operators, precedence, pratt-parser, fqn-disambiguation
---

# Node: parser.parseInfixExpression (Generic Binary Operator)

## 1. Architectural Role & Intent
The single handler behind every ordinary binary operator - arithmetic, comparison, equality, and the boolean/membership keywords (`and`, `or`, `xor`, `in`, `matches`, `contains`). It parses the right operand at the operator's own precedence and emits a uniform `InfixExpression` carrying the operator as a **string**. It also holds one targeted special case: `/` is parsed at `INDEX` precedence so that slash-separated callee chains like `com/ex/f(x)` group as a call on the chain rather than as division by a call result.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.infix` | `CALLS` | [[parser.expression]] | Recurses for the right operand at the computed precedence. |
| `parser.infix` | `READS_FROM` | [[parser.precedence]] | Receives the operator's binding power from the Pratt loop; overrides it for `/`. |
| `parser.infix` | `CALLS` | [[ast]] | Emits `ast.NewInfixExpression(left, right, operatorString, span)`. |
| [[parser.lookups]] | `CALLS` | [[parser.infix]] | Registered for roughly fifteen token kinds. |
| [[parser.not]] | `CALLS` | [[parser.infix]] | Builds the same node type for negated membership forms. |
| [[parser.is]] | `CALLS` | [[parser.infix]] | Falls back to this node type for the general `a is b` comparison. |
| [[runtime.eval_infix]] | `DEPENDS_ON` | [[parser.infix]] | Dispatches on the operator string and applies [[trinary]] Kleene logic for boolean operators. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseInfixExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes the operator token, then parses the right operand at `precedence` - **left-associative**, since an equal-precedence operator on the right will not be absorbed. For `TokenDiv` the right-hand precedence is raised to `INDEX`, making `/` bind tighter than a following call. The span runs from the **operator** to the end of the right operand.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` when the right operand fails to parse **or** when `p.err` is already set - one of the few productions that checks the accumulated error defensively.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One recursion per operator; chains are linear.
- **Dependencies Risk:**
  - **The operator is an untyped string.** `InfixExpression.Operator` is the raw token text, so [[runtime.eval_infix]] dispatches on string comparison. A lexer change to an operator's spelling silently changes evaluation behaviour with no compile-time error anywhere.
  - **The span omits the left operand.** Ranges start at the operator, so a diagnostic about `a + b` highlights `+ b`. This is consistent across the package but consistently misleading.
  - **The `/` override is load-bearing and subtle.** It exists so slash chains behave as paths in callee position, which means `a / b(c)` - genuine division by a call - parses as `(a/b)(c)`. Division followed by a parenthesised expression is therefore a syntax trap.
  - **No operand type checking.** `"a" - 1` parses cleanly; type errors surface only at evaluation.
