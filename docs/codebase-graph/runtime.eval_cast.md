---
id: runtime.eval_cast
type: Function / Endpoint
language: Go
file_path: runtime/eval_cast.go
tags: type-conversion, coercion, validation, defect
---

# Node: runtime.evalCast (Type Cast Evaluation)

## 1. Architectural Role & Intent
Implements `expr as <type>` — the language's only explicit conversion mechanism. It converts to string, number, or trinary; asserts (without converting) for list, dict, and shape targets; and is *intended* to validate the converted value against the full target type, including constraints, before returning. That validation is currently ineffective — see the defect note below.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_cast` | `DEPENDS_ON` | [[ast]] | Switches on the target `ast.TypeRef` implementation. |
| `runtime.eval_cast` | `DEPENDS_ON` | [[box]] | `String`, `Number`, `Trinary`, `TrinaryFrom`, `NumberValue`, `StringValue`, `BoolValue`, `Kind`. |
| `runtime.eval_cast` | `CALLS` | [[runtime.eval]] | Evaluates the operand expression. |
| `runtime.eval_cast` | `CALLS` | [[runtime.typeref]] | `validateValueAgainstTypeRef` — invoked, but its result is discarded. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_cast]] | `ast.CastExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalCast(ctx, ec, e, p, cast *ast.CastExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Evaluates the operand, then dispatches on the target type ref.
  - **Side Effects:** Operand evaluation.
  - **Exceptions:** `cannot cast %s to number`; `cannot cast %s to list`; `cannot cast %s to dict`; `strconv.ParseFloat` errors.

- **Signature:** Conversion rules by target
  - **Behavior:**
    - **String** — always succeeds via `val.String()`; every value has a string rendering.
    - **Number** — accepts a number as-is, parses a string with `ParseFloat`, and maps a boolean to `1`/`0`. Everything else errors.
    - **Trinary** — always succeeds via `box.TrinaryFrom`, applying language truthiness.
    - **List / Dict** — pure **assertions**: the value must already be that kind, and is returned unchanged.
    - **Shape** — returned unchanged with no structural check at this point.
    - **Default** — returned unchanged.
  - **Side Effects:** None.
  - **Exceptions:** As above.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Negligible except for the intended post-cast validation, which would walk shapes and constraints.
- **Dependencies Risk:**
  - **The deferred validation and panic recovery are dead code.** The function's return values are **unnamed**, so the deferred closure's assignments to `result` and `err` cannot affect what is returned. Two consequences: (1) a cast to a constrained or shaped type returns the **unvalidated** value with a `nil` error, so casts are not a type-enforcement mechanism today; (2) a panic inside the switch is recovered but the function then returns zero values — an invalid `box.Value`, a **nil `*trace.Node`**, and a **nil error** — so the failure is invisible and a caller dereferencing the node panics in turn. Filed as [#107](https://github.com/sentrie-sh/sentrie/issues/107); the fix is to name the return values.
  - **String casts never fail.** Because `val.String()` is total, `someDict as string` produces a rendered dictionary rather than an error — combined with the dead validation, there is no guard against nonsensical conversions.
  - **List, dict, and shape casts do not convert.** They are assertions, so `as` cannot be used to coerce a JS-returned structure into a shape; it only checks the outer kind, and for shapes not even that.
  - **Number casts accept booleans**, which combined with trinary truthiness means round-tripping a trinary through number loses the `Unknown` state.
  - **Cast precedence is high** — see [[parser.cast]] — so `a + b as string` casts only `b`. That interacts with the stringifying `+` in [[runtime.eval_infix]] to produce surprising concatenations.
