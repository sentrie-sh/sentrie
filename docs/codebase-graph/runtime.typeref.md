---
id: runtime.typeref
type: Function / Endpoint
language: Go
file_path: runtime/typeref.go
tags: type-validation, dispatch, nullability, runtime-checking
---

# Node: runtime.validateValueAgainstTypeRef (Type Validation Dispatcher)

## 1. Architectural Role & Intent
The single entry point for all runtime type checking. It strips nullability, then dispatches on the concrete `ast.TypeRef` implementation to the matching per-kind validator. Every place a declared type must be enforced against an actual value — facts, lets, derive parameters and returns, casts — funnels through this one function.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref` | `DEPENDS_ON` | [[ast]] | `IsNullableTypeRef`, `UnwrapNullableTypeRef`, and the eight concrete type-ref kinds. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_string]] | String targets. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_number]] | Number targets. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_trinary]] | Trinary and bool targets. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_list]] | List targets; recurses per element. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_dict]] | Dict targets. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_shape]] | Shape targets; recurses per field. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_document]] | Document targets. |
| `runtime.typeref` | `CALLS` | [[runtime.typeref_record]] | Record (positional tuple) targets. |
| [[runtime.executor]] | `CALLS` | [[runtime.typeref]] | Validates injected fact values against their declared types. |
| [[runtime.eval_ident]] | `CALLS` | [[runtime.typeref]] | Validates a let's value against its annotation on first read. |
| [[runtime.derive_invoke]] | `CALLS` | [[runtime.typeref]] | Validates derive parameters and return values. |
| [[runtime.eval_cast]] | `CALLS` | [[runtime.typeref]] | Intended post-cast validation — currently discarded by that caller. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateValueAgainstTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef ast.TypeRef, valueRange tokens.Range) -> error`
  - **Behavior:** If the type ref is nullable and the value is null, passes immediately; otherwise unwraps the nullable and dispatches on the concrete kind. **An unrecognised type ref returns `nil`** — an unconditional pass.
  - **Side Effects:** None directly, but the per-kind validators evaluate constraint argument expressions, which can call builtins and derives.
  - **Exceptions:** Whatever the delegated validator returns.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Recursion depth follows the type's nesting. A list of shapes revalidates every field of every element on every check, and constraint argument expressions are re-evaluated at each level rather than hoisted — so validating a large collection against a constrained element type is proportional to elements times constraints.
- **Dependencies Risk:**
  - **Unknown type refs pass silently.** The switch has no `default`, so a type-ref kind added to [[ast]] without a corresponding validator here is accepted for every value. The failure mode is "no checking at all", which is the least visible one possible.
  - **Null is only permitted through an explicit nullable wrapper.** A bare `string` rejects null with `value <nil> is not a string` rather than a nullability-specific message, which reads as a type error rather than a missing-value error.
  - **`Executor` is taken as the interface but every validator immediately asserts `exec.(*executorImpl)`** to evaluate constraint arguments. That assertion is unchecked, so any other implementation of the exported `Executor` interface — a test double, a wrapper, a future decorator — **panics**. The interface in the signature is therefore misleading; the code requires the concrete type.
  - **The `valueRange` is threaded through but largely unused.** `ErrConstraintFailed` ignores its position argument, and composite validators pass the *container's* range down to element checks, so a failure inside a list points at the whole list.
