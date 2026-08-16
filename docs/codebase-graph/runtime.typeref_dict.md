---
id: runtime.typeref_dict
type: Function / Endpoint
language: Go
file_path: runtime/typeref_dict.go
tags: type-validation, constraints, collections
---

# Node: runtime.validateAgainstDictTypeRef

## 1. Architectural Role & Intent
Validates a value against the `dict` type. Deliberately shallow: it confirms the value is a dictionary and applies dict-level constraints, but **does not inspect keys or values**. A `dict` is the untyped map escape hatch; `shape` is the mechanism for structural guarantees.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_dict` | `DEPENDS_ON` | [[constraints]] | `constraints.DictContraintCheckers`. |
| `runtime.typeref_dict` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_dict` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_dict]] | Dispatched for `*ast.DictTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstDictTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.DictTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires `v.DictValue()`, then runs dict constraints. No recursion into members.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value %v is not a dict`; `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** O(constraints), independent of dictionary size - the cheapest composite validator precisely because it does not recurse.
- **Dependencies Risk:**
  - **`dict` provides no structural guarantee whatsoever.** Any map passes. Field access on a dict-typed value therefore has no compile-time or runtime backing, and [[runtime.eval_access]] returns `Undefined` for absent members - so a typo'd field name silently yields undefined rather than failing anywhere.
  - **`dict` and `document` are nearly identical validators** with separate constraint tables; the only real difference is which checkers apply. The distinction is not enforced structurally.
  - **The error message lacks a position**, unlike its list and document siblings which include `pos`.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
