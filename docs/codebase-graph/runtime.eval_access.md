---
id: runtime.eval_access
type: Function / Endpoint
language: Go
file_path: runtime/eval_access.go
tags: field-access, indexing, null-safety, boundary-values
---

# Node: runtime.evalAccess (Field and Index Access)

## 1. Architectural Role & Intent
Implements `x.field` and `x[i]` over both native boxed containers and raw boundary values returned from JavaScript modules. Its defining behaviour is that **a missing member is `Undefined`, not an error** - access chains degrade gracefully rather than failing, which is what makes `x.a.b.c` usable against loosely-shaped external data without guards at every hop.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_access` | `DEPENDS_ON` | [[box]] | `DictValue`, `ListValue`, `ObjectRef`, `NumberValue`, `StringValue`, `FromBoundaryAny`. |
| `runtime.eval_access` | `CALLS` | [[runtime.eval]] | Evaluates the receiver and, for index access, the index expression. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_access]] | `ast.FieldAccessExpression` and `ast.IndexAccessExpression` dispatch here. |
| [[runtime.modules]] | `READS_FROM` | [[runtime.eval_access]] | Values returned from JS arrive as `ObjectRef` payloads that these functions traverse. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalFieldAccess(ctx, ec, exec, p, t) -> (box.Value, *trace.Node, error)` / `evalIndexAccess(...)`
  - **Behavior:** Evaluate the receiver (and index), then delegate to the pure accessors. Index access evaluates the receiver **before** the index expression.
  - **Side Effects:** Whatever the sub-expressions do.
  - **Exceptions:** Propagated from sub-evaluation, plus accessor errors.

- **Signature:** `accessField(_, obj box.Value, field string) -> (box.Value, error)`
  - **Behavior:** An undefined receiver yields `Undefined` (chain-safe). A boxed dict yields the member or `Undefined`. A raw `map[string]any` from the boundary is converted with `FromBoundaryAny`. Anything else is an error.
  - **Side Effects:** None.
  - **Exceptions:** `cannot access field '%s' on %T`.

- **Signature:** `accessIndex(_, col box.Value, idx box.Value) -> (box.Value, error)`
  - **Behavior:** An undefined collection yields `Undefined`. Lists resolve the index through `listIndex`: an integer, or an integral float such as `2.0`, selects the element, and one outside the list (negative, too large, too large for an int, or an infinity) yields `Undefined`; an undefined index yields `Undefined`. Boxed dicts index by string key, and a missing key yields `Undefined`. Raw `[]any` and `map[string]any` from the boundary behave the same way after conversion.
  - **Side Effects:** None.
  - **Exceptions:** `list index must be an integer, got <kind>` for a string, bool, null or container index on a list, and `list index must be an integer, got <n>` for a number with a fractional part or NaN ([#149](https://github.com/sentrie-sh/sentrie/issues/149)); `index access not supported on %T`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless, pure accessors.
- **Performance/Scale Notes:** Boundary values are converted on **every access**, so repeatedly indexing into a large JS-returned structure re-wraps values each time rather than converting once.
- **Dependencies Risk:**
  - **A list index is strict, a dict key is not.** Until [#149](https://github.com/sentrie-sh/sentrie/issues/149) a string or null list index silently read element zero and `list[1.9]` truncated to `list[1]`; both are now errors. A non-string key on a dict is still `Undefined`, not an error, so the two containers disagree on a wrong-kinded index.
  - **A null receiver is not the same as an undefined receiver.** Only `IsUndefined` is checked, so `null.field` falls through to the error branch while `undefined.field` yields `Undefined` - an asymmetry that makes chain safety depend on which absence representation a value happens to carry.
  - **The error messages format `obj`, not the value's kind.** `%T` on a `box.Value` always prints `box.Value`, so `cannot access field 'x' on box.Value` tells a policy author nothing about what the receiver actually was. `obj.Kind()` would be the useful value.
  - **There is no null-safe access operator** in the language, so this permissive behaviour *is* the null-safety mechanism - which means tightening it would break policies that rely on the degradation.
  - **A missing field and a field explicitly set to undefined are indistinguishable**, so `is defined` cannot separate "absent" from "present but undefined".
