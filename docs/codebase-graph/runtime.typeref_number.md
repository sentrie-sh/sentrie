---
id: runtime.typeref_number
type: Function / Endpoint
language: Go
file_path: runtime/typeref_number.go
tags: type-validation, constraints, numbers
---

# Node: runtime.validateAgainstNumberTypeRef

## 1. Architectural Role & Intent
Validates a value against a `number` type and applies the numeric constraint table (`@min`, `@max`, range checks). Structurally identical to the string validator: kind check, then a constraint loop.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_number` | `DEPENDS_ON` | [[constraints]] | `constraints.NumberContraintCheckers`. |
| `runtime.typeref_number` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_number` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_number]] | Dispatched for `*ast.NumberTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstNumberTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.NumberTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires `v.NumberValue()` to succeed, then runs each constraint in order.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value %v is not a number`; `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Same per-element constraint argument re-evaluation cost as the string validator.
- **Dependencies Risk:**
  - **There is one numeric kind.** [[box]] stores all numbers as `float64`, so this validator cannot distinguish an integer from a float. A type intended to mean "integer" must express that as a constraint, not as a type.
  - **Numeric strings are rejected**, so a fact arriving as `"42"` from JSON fails a `number` declaration rather than coercing.
  - **Float precision applies to the constraint checks too** — a large integer beyond exact float representation validates against a bound it does not actually satisfy.
  - Shares the `exec.(*executorImpl)` assertion and the first-failure-wins behaviour described in [[runtime.typeref_string]].
