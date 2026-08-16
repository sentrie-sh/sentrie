---
id: runtime.typeref_string
type: Function / Endpoint
language: Go
file_path: runtime/typeref_string.go
tags: type-validation, constraints, strings
---

# Node: runtime.validateAgainstStringTypeRef

## 1. Architectural Role & Intent
Validates a value against a `string` type, then applies each attached constraint (`@length`, `@matches`, and the rest of the string constraint table). It is the template every other scalar validator follows: kind check first, then a constraint loop that evaluates each constraint's argument expressions in the caller's scope before invoking the checker.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_string` | `DEPENDS_ON` | [[constraints]] | `constraints.StringContraintCheckers` maps a constraint name to its checker. |
| `runtime.typeref_string` | `CALLS` | [[runtime.eval]] | Evaluates each constraint argument expression. |
| `runtime.typeref_string` | `CALLS` | [[runtime.err_typedef]] | `ErrUnknownConstraint` and `ErrConstraintFailed`. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_string]] | Dispatched for `*ast.StringTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstStringTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.StringTypeRef, valueRange tokens.Range) -> error`
  - **Behavior:** Requires `v.StringValue()` to succeed, then runs every constraint in declaration order, stopping at the first failure.
  - **Side Effects:** Constraint argument expressions are evaluated in the caller's context and can invoke builtins or derives.
  - **Exceptions:** `value %v is not a string`; `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Constraint arguments are re-evaluated on **every** validation. Inside a list-of-strings check this means once per element, so a constraint whose argument calls a builtin multiplies that call by the collection size.
- **Dependencies Risk:**
  - **No implicit coercion.** A number that looks like a string is rejected outright; the type system does not widen at validation time even though [[runtime.eval_cast]] would convert.
  - **An unknown constraint is a *runtime* error.** The name is not verified when the pack is indexed, so a typo in a constraint name passes `sentrie validate` and fails at decision time. See [[runtime.err_typedef]].
  - **`exec.(*executorImpl)`** is asserted without a check, so a non-concrete `Executor` panics here rather than at the dispatcher.
  - **The error message renders the value with `%v`**, so a long or sensitive string is embedded verbatim in the diagnostic - worth remembering when validation errors are surfaced to API callers.
  - **First failure wins.** Constraints do not accumulate, so an author fixing several violations discovers them one round-trip at a time.
