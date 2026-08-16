---
id: parser.is
type: Function / Endpoint
language: Go
file_path: parser/is.go
tags: predicates, definedness, emptiness, infix-handler, undefined-semantics
---

# Node: parser.parseIsExpression (Definedness and Emptiness Predicates)

## 1. Architectural Role & Intent
Parses the `is` family: `x is defined`, `x is empty`, their negated forms with an optional `not`, and the general `a is b` comparison. The definedness predicate is architecturally significant — it is the syntax through which a policy can *ask* whether a fact was supplied rather than having its absence silently propagate as `Unknown` through [[trinary]] logic, which is the difference between an explicit "I cannot decide" and an accidental one.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.is` | `CALLS` | [[parser.expression]] | Parses the right operand for the general comparison form. |
| `parser.is` | `CALLS` | [[ast]] | Emits `IsDefinedExpression`, `IsEmptyExpression`, or `InfixExpression`, optionally wrapped in `UnaryExpression`. |
| `parser.is` | `CALLS` | [[parser.parser]] | Uses `expect(KeywordIs)`, `head`, `advance`, `canExpect`. |
| [[parser.lookups]] | `CALLS` | [[parser.is]] | Registered as the infix handler for `KeywordIs` at `EQUALITY` precedence. |
| [[box.value]] | `READS_FROM` | [[ast]] | `IsDefined` resolves against the boxed undefined sentinel rather than null. |
| [[runtime.eval]] | `DEPENDS_ON` | [[ast]] | Evaluates all three produced forms. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseIsExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `is`, optionally consumes `not`, then dispatches on the follower:
    - `defined` → `IsDefinedExpression(left)`
    - `empty` → `IsEmptyExpression(left)`
    - anything else → parses a right operand and builds `InfixExpression(left, "is", right)`

    A captured `not` wraps whichever node was produced in a `UnaryExpression`, so `is not defined` desugars to `not (is defined)`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` if `is` is missing, or if the general form's right operand fails to parse (the underlying error having been recorded by the operand's own handler).

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Regression history.** The general-comparison branch previously called `right.Span()` with no nil check, so a malformed right operand (`x is` at end of input) **panicked** instead of reporting a parse error. Fixed and covered by `parser/not_test.go`; keep the nil guard when touching this branch.
  - **The predicate spans are truncated.** `IsDefinedExpression` and `IsEmptyExpression` are built with the range of the `is` token captured *before* the follower is consumed, so their spans cover neither the left operand nor the `defined`/`empty` keyword.
  - **`defined` vs `null` are different questions.** `is defined` tests presence (the undefined sentinel in [[box.value]]), not nullness — a fact explicitly set to null **is** defined. Conflating them is the most common semantic mistake with this operator.
  - **`is not X` is desugared, not a distinct node**, so tooling cannot distinguish `x is not defined` from `not (x is defined)`.
