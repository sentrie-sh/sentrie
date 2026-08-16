---
id: runtime.eval_unary
type: Function / Endpoint
language: Go
file_path: runtime/eval_unary.go
tags: operators, negation, kleene-logic, arithmetic
---

# Node: runtime.evalUnary (Prefix Operators)

## 1. Architectural Role & Intent
Implements the four prefix operators: logical negation (`!` and `not`, which are the same operation) and numeric sign (`+`, `-`). Negation goes through Kleene `Not`, so negating an `Unknown` stays `Unknown` rather than flipping to a definite value - the property that keeps three-valued logic sound through negation.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_unary` | `DEPENDS_ON` | [[box]] | `TrinaryFrom`, `NumberValue`, `Trinary`, `Number`. |
| `runtime.eval_unary` | `DEPENDS_ON` | [[trinary]] | `Not()` implements Kleene negation. |
| `runtime.eval_unary` | `CALLS` | [[runtime.eval]] | Evaluates the operand. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_unary]] | `ast.UnaryExpression` nodes dispatch here. |
| [[parser.not]] | `DEPENDS_ON` | [[runtime.eval_unary]] | Negated membership forms desugar to a `not` unary wrapping an infix, so they land here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalUnary(ctx, ec, exec, p, u *ast.UnaryExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Evaluates the operand; an undefined operand short-circuits to `Undefined` before the operator switch, matching [[runtime.eval_infix]].
  - **Side Effects:** Operand evaluation.
  - **Exceptions:** `unary + requires number`; `unary - requires number`; `unsupported unary op: %s`.

- **Signature:** `!` and `not`
  - **Behavior:** Identical: coerce with `box.TrinaryFrom`, apply Kleene `Not`. `not Unknown` is `Unknown`.
  - **Side Effects:** None.
  - **Exceptions:** None - coercion is total.

- **Signature:** `+` and `-`
  - **Behavior:** Require a number; `+` is an identity that still enforces the type, `-` negates.
  - **Side Effects:** None.
  - **Exceptions:** Type errors as above.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Negligible.
- **Dependencies Risk:**
  - **Negation is total, so it never type-errors.** `not "hello"` coerces the string through truthiness and negates it rather than reporting a type problem. Combined with the absence of static operator type checking, a logic error involving a non-trinary operand is silent.
  - **Unary `+` is preserved through the parser** rather than folded away (see [[parser.unary]]), so it reaches here and acts as a runtime number assertion. That makes `+x` a meaningful - if obscure - way to require numeric-ness, since the failure is an error rather than a coercion.
  - **Undefined propagation precedes the switch**, so `-undefined` is `Undefined` rather than a type error, and `not undefined` is `Undefined` rather than `Unknown`. As in the infix evaluator, `Undefined` behaves as a fourth state that absorbs the operator.
  - **The error branch returns `box.Value{}`, not `box.Undefined()`.** The zero value and the undefined value are different things; callers that inspect the value on an error path see an invalid box rather than a defined absence. The same inconsistency appears in [[runtime.eval_cast]] and [[runtime.eval_access]].
  - **Error messages omit the source span**, unlike most diagnostics in the package, so `unary - requires number` gives no location.
