---
id: runtime.eval_is
type: Function / Endpoint
language: Go
file_path: runtime/eval_is.go
tags: presence, emptiness, undefined, guards
---

# Node: runtime.evalIsDefined / evalIsEmpty (Presence Predicates)

## 1. Architectural Role & Intent
Evaluates `x is defined` and `x is empty`, the documented guards against missing data. They are the only expressions that take an `Undefined` operand and return a definite value instead of propagating it, so they are the point where a policy turns "the data is missing" into an explicit yes or no rather than an accidental `Unknown`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_is` | `CALLS` | [[runtime.eval]] | Evaluates the operand. |
| `runtime.eval_is` | `DEPENDS_ON` | [[box]] | `IsUndefined`, `Kind`, length of string, list, dict and host document; boxes the result with `box.Bool`. |
| `runtime.eval_is` | `CALLS` | [[runtime.trace]] | One `is` node per predicate with the operand as its child. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_is]] | `ast.IsDefinedExpression` and `ast.IsEmptyExpression` dispatch here. |
| [[parser.is]] | `DEPENDS_ON` | [[runtime.eval_is]] | Produces the two nodes; a `not` wraps them in a unary, so negation lands in [[runtime.eval_unary]]. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalIsDefined(ctx, ec, exec, p, t *ast.IsDefinedExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** `true` unless the operand evaluates to `Undefined`. `null` is defined. A missing field, an out-of-range index, an omitted optional derive or lambda parameter and an omitted optional fact with no default all read as `Undefined` (see [[runtime.eval_access]], [[runtime.eval_ident]]), so all of them give `false`.
  - **Side Effects:** Operand evaluation.
  - **Exceptions:** None of its own; an operand evaluation error propagates unchanged, with `Undefined` as the value.

- **Signature:** `evalIsEmpty(ctx, ec, exec, p, t *ast.IsEmptyExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** `true` for `""`, `[]`, `{}`, `null` and `Undefined`, and for a host document wrapping an empty `map[string]any` or `[]any`. Every other value, scalars included, is not empty and gives `false` rather than an error.
  - **Side Effects:** Operand evaluation.
  - **Exceptions:** As for `evalIsDefined`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One operand evaluation and a length check; negligible.
- **Dependencies Risk:**
  - **Only a value can be undefined; an error stays an error.** `x is defined` on an undeclared identifier, a field access on a non-dict, or a failing call still fails. Presence is decided by how the runtime already represents absence, not by swallowing evaluation errors.
  - **The negated forms box differently.** The predicates return `box.Bool`, but `is not defined` / `is not empty` go through Kleene `Not` in [[runtime.eval_unary]], which returns a `box.Trinary`. Both coerce to the same trinary, so decisions agree; code inspecting the raw box sees two kinds.
  - **A field present with an undefined value is indistinguishable from an absent one**, so `is defined` cannot separate them (see [[runtime.eval_access]]).
  - **`null` is defined but empty.** `x is defined and x is not empty` is the guard for "supplied and has content".
