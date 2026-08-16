---
id: runtime.typeref_trinary
type: Function / Endpoint
language: Go
file_path: runtime/typeref_trinary.go
tags: type-validation, constraints, kleene-logic
---

# Node: runtime.validateAgainstTrinaryTypeRef

## 1. Architectural Role & Intent
Validates a value against the `trinary` type. Unlike the other scalar validators it **accepts two representations** - a native boolean, widened via `trinary.From`, or an already-trinary value - and normalises to a trinary before running constraints, so a policy declaring `trinary` transparently accepts a boolean fact.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_trinary` | `DEPENDS_ON` | [[trinary]] | `trinary.From` widens a boolean into the three-valued domain. |
| `runtime.typeref_trinary` | `DEPENDS_ON` | [[constraints]] | `constraints.TrinaryConstraintCheckers`. |
| `runtime.typeref_trinary` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_trinary` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_trinary]] | Dispatched for `*ast.TrinaryTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstTrinaryTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.TrinaryTypeRef, valueRange tokens.Range) -> error`
  - **Behavior:** Accepts a bool or a trinary, normalises to `trinary.Value`, then runs constraints against the **re-boxed** trinary rather than the original value.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value '%v' is not a bool at %s - expected bool`; `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Negligible beyond constraint argument evaluation.
- **Dependencies Risk:**
  - **The error message says "bool" for a trinary type.** `value '%v' is not a bool ... expected bool` is the only diagnostic, so an author who declared `trinary` is told they needed a bool - misleading given `Unknown` is a legal value the message never mentions.
  - **Constraints see the normalised trinary, not the input.** A boolean input is re-boxed before checking, so a constraint cannot distinguish "was a bool" from "was a trinary that happened to be definite".
  - **Truthiness coercion is *not* applied here.** Unlike [[runtime.eval_infix]], which coerces anything through `TrinaryFrom`, this validator only accepts the two exact kinds. So a value that behaves as true in a condition still fails a `trinary` declaration - validation is stricter than evaluation.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
