---
id: runtime.typeref_dict
type: Function / Endpoint
language: Go
file_path: runtime/typeref_dict.go
tags: type-validation, constraints, collections, recursion
---

# Node: runtime.validateAgainstDictTypeRef

## 1. Architectural Role & Intent
Validates a value against a `dict[T]` type. Like [[runtime.typeref_list]] it is recursive: after confirming the value is a dictionary it revalidates **every value** against the value type `T`, then applies dict-level constraints. Keys are not checked; dicts are string-keyed. `document` is the untyped map; `shape` is the mechanism for per-field guarantees.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_dict` | `CALLS` | [[runtime.typeref]] | Recurses once per value against `typeRef.ValueType`, in sorted key order. |
| `runtime.typeref_dict` | `DEPENDS_ON` | [[constraints]] | `constraints.DictContraintCheckers`. |
| `runtime.typeref_dict` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_dict` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_dict]] | Dispatched for `*ast.DictTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstDictTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.DictTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires `v.DictValue()`, validates each value against `ValueType` in sorted key order, then runs dict constraints. Values are checked **before** constraints. A nullable value type (`dict[T?]`) accepts a null value. Since #299; before it any dict conformed to any `dict[T]`.
  - **Side Effects:** Value validation can evaluate nested constraint arguments.
  - **Exceptions:** `value %v is not a dict`; `value for key %q is not valid at %s: %w` (the key is quoted, since fact keys are untrusted) (wrapping the value's own error, so the outermost marker is the value's and a bad value is a mismatch for `is`); `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** O(values) recursions plus a key sort, so as expensive as [[runtime.typeref_list]] for the same size. Each value's constraints re-evaluate their argument expressions independently.
- **Dependencies Risk:**
  - **`dict[T]` guarantees value types, not keys.** Any key set passes, an empty dict trivially passes, and [[runtime.eval_access]] returns `Undefined` for an absent member, so a typo'd key silently yields undefined rather than failing anywhere.
  - **The reported key is the first failing one in sorted order**, not insertion order, so the error is stable across runs; validation stops there.
  - **An imported decision envelope is not a `dict[trinary]`.** Its `value` and attachments need not be truth values; bind it as `document` (or a truth type) instead.
  - **`document` is the dict validator without the value check**, with its own constraint table.
  - **The not-a-dict error lacks a position**, unlike its list and document siblings which include `pos`; the per-key error includes it.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
