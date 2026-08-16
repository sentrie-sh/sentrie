---
id: runtime.typeref_list
type: Function / Endpoint
language: Go
file_path: runtime/typeref_list.go
tags: type-validation, constraints, collections, recursion
---

# Node: runtime.validateAgainstListTypeRef

## 1. Architectural Role & Intent
Validates a value against a `list<T>` type. It is the first of the recursive validators: after confirming the value is a list it revalidates **every element** against the element type before applying list-level constraints such as `@size`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_list` | `CALLS` | [[runtime.typeref]] | Recurses once per element against `typeRef.ElemType`. |
| `runtime.typeref_list` | `DEPENDS_ON` | [[constraints]] | `constraints.ListContraintCheckers`. |
| `runtime.typeref_list` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_list` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_list]] | Dispatched for `*ast.ListTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstListTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ListTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires `v.ListValue()`, validates each element, then runs list constraints. Elements are checked **before** constraints, so a `@size` violation is reported only if every element already type-checks.
  - **Side Effects:** Element validation can evaluate nested constraint arguments.
  - **Exceptions:** `value %v is not an array at %s - expected array`; `item is not valid at %s: %w`; constraint errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** **This is the expensive validator.** Cost is O(elements) recursions, and each element's constraints re-evaluate their argument expressions independently. A list of shapes with constrained fields multiplies out quickly, and there is no memoisation of constraint arguments across elements. A large fact validated against a deeply-typed list is a plausible latency source.
- **Dependencies Risk:**
  - **Element errors do not say which element failed.** The wrapper is `item is not valid at %s` with the *container's* range and no index - the source carries a `TODO` acknowledging this. Debugging a failure in a hundred-element list means bisecting by hand.
  - **Validation stops at the first bad element**, so a caller cannot see how widespread a problem is from one response.
  - **Raw boundary lists are not accepted.** Only `v.ListValue()` is tried, not `ObjectRef` holding a `[]any`, so a list returned from a JavaScript module may fail a `list<T>` declaration depending on how it was boxed - unlike [[runtime.eval_access]], which handles both representations.
  - **An empty list trivially passes the element check**, so an element type is no guarantee the list is non-empty; that requires an explicit size constraint.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
