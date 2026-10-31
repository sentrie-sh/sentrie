---
id: runtime.typeref_constraints
type: Module / File
language: Go
file_path: runtime/typeref_constraints.go
tags: validation, constraints, type-system, typeref
---

# Node: runtime.typeref_constraints (Constraint Binding and Checking)

## 1. Architectural Role & Intent
The constraint step every typeref validator shares. It splits a type ref's constraints into two phases: `bindConstraints` evaluates the arguments and checks them against the definition before the value is looked at, and `checkConstraints` runs the checkers once the value's kind is known. Binding first means a broken constraint argument is an error for every value, not only for values of the constraint's own kind (#273), so `x is not T` never reads a broken type as a mismatch.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_constraints` | `DEPENDS_ON` | [[constraints]] | `ForTypeRef` selects the table; `CheckArgs` and `Checker` run the definition. |
| `runtime.typeref_constraints` | `CALLS` | [[runtime.eval]] | Evaluates each constraint argument expression. |
| `runtime.typeref_constraints` | `CALLS` | [[runtime.err_typedef]] | `ErrUnknownConstraint`, `ErrConstraintFailed`, `errCheckFailed`. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_constraints]] | Binds the inner type's constraints before a nullable type's null short-circuit. |
| [[runtime.typeref_string]] | `CALLS` | [[runtime.typeref_constraints]] | Binds before the kind check, checks after it; every other typeref validator does the same. |

## 3. Interface Contracts & Public Surface

- **Signature:** `bindConstraints(ctx, ec, exec, p, typeRef) -> ([]boundConstraint, error)`
  - **Behavior:** For each of the type ref's own constraints, in order: evaluates the arguments, looks the name up in `constraints.ForTypeRef(typeRef)`, and runs `CheckArgs`. Returns the bound constraints for `checkConstraints`.
  - **Side Effects:** Argument expressions are evaluated in the caller's context.
  - **Exceptions:** `errCheckFailed` when an argument fails to evaluate; `ErrUnknownConstraint`; `ErrConstraintFailed` wrapping the argument failure, which is **not** a type mismatch (it matches `constraints.ErrInvalidArgument`).

- **Signature:** `checkConstraints(ctx, v, bound, pos) -> error`
  - **Behavior:** Runs each bound checker against `v`, stopping at the first failure.
  - **Side Effects:** None.
  - **Exceptions:** `ErrConstraintFailed`, a type mismatch unless the checker blamed its arguments.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Arguments are evaluated and checked on every validation, and each definition's `Checker` re-runs `CheckArgs` itself, so a constraint's arguments are checked twice per value (per element inside a list element type). For `regexp` both checks and the match reuse one compiled pattern from [[constraints]]' cache, so the pattern is compiled once per process.
- **Dependencies Risk:**
  - **Only the type ref being validated is bound.** A broken argument on a list element type is reached only per element, so an empty list or a non-list value does not report it at runtime. [[index.constraint_args]] covers every position statically, and `sentrie validate`, `exec` and `serve` all validate before evaluating.
  - **An unknown name is reported before the kind check.** The parser already rejects unknown names, so this is reachable only if [[ast]]'s generated tables drift from [[constraints]].
