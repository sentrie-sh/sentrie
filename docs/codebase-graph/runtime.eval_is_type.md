---
id: runtime.eval_is_type
type: Function / Endpoint
language: Go
file_path: runtime/eval_is_type.go
tags: type-check, shape-conformance, validation, predicates
---

# Node: runtime.evalIsType (Shape and Type Check)

## 1. Architectural Role & Intent
Implements `x is T` (#150): a total predicate that answers whether a value validates against a shape or type. Unlike a cast ([[runtime.eval_cast]]) or a typed let, a value that does not conform is not an error here but `false`, which lets a policy branch on the shape of heterogeneous input without aborting the rule. It reuses the one type checker, [[runtime.typeref]], and tells "does not conform" from "could not check" through the mismatch marker that checker carries; a failure to check is always an error, so `x is not T` can never fail open.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_is_type` | `DEPENDS_ON` | [[ast]] | Reads `ast.IsExpression` (`Left`, `Right`, `Type`). |
| `runtime.eval_is_type` | `CALLS` | [[runtime.eval]] | Evaluates the operands. |
| `runtime.eval_is_type` | `CALLS` | [[runtime.typeref]] | `validateValueAgainstTypeRef`, then `isTypeMismatch` on its error. |
| `runtime.eval_is_type` | `CALLS` | [[index.is_resolve]] | `(*Index).LookupShape`, only for a node an unvalidated index left ambiguous. |
| `runtime.eval_is_type` | `READS_FROM` | [[runtime.exec_ctx]] | `evalDerive` selects the derive's own shape scope. |
| `runtime.eval_is_type` | `DEPENDS_ON` | [[box]] | `box.Bool` results; `EqualValues` on the equality path. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_is_type]] | `ast.IsExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalIsType(ctx, ec, exec, p, is *ast.IsExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** With a type operand: evaluates `Left`; an undefined value gives `false`; otherwise validates it against `Type` and returns `true`, or `false` when the validation error is a type mismatch. With a value operand (`Type` cleared by [[index.is_resolve]]): `evalIsEquality`, the same operation as the infix `is` in [[runtime.eval_infix]] (undefined if either side is undefined, else `box.EqualValues`), without allocating an infix node. With both set (an index that was never validated): looks the name up with `LookupShape` and takes the type reading when it is a shape, the value reading otherwise.
  - **Side Effects:** Operand and constraint-argument evaluation.
  - **Exceptions:** Errors from evaluating `Left`; validation errors that are not mismatches: an unknown or unexported shape (including one named by a nested field), a constraint whose own arguments are invalid (`constraints.ErrInvalidArgument`: wrong kind or count, invalid regexp, zero divisor), a constraint argument that fails to evaluate; `LookupShape` errors on the unvalidated path.

- **Signature:** `isShapeScope(ec, p) -> *index.Policy`
  - **Behavior:** The policy whose shapes the check resolves against: `p` outside a derive; inside a derive body `(*index.Derive).ShapeScope()`: the derive's own policy, or for a namespace-level derive a `Policy` holding only the derive's namespace. This is the scope [[index.is_resolve]] resolved the names in.
  - **Side Effects:** None; the derive's scope comes from `(*index.Derive).ShapeScope`, built once when the derive is indexed.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One full validation of the value per evaluation; a list of shapes checks every field of every element.
- **Dependencies Risk:**
  - **Derive bodies are allowed.** The check resolves no identifier or callee itself; `Left` and the constraint arguments go through [[runtime.eval]], whose identifier and call paths apply the `evalDerive` gates.
  - **Derive shapes resolve in the derive's scope, unlike casts and typed parameters.** An exported derive that checks `v is Foo` works from any namespace and never picks up a caller's shape named `Foo`. Casts, typed parameters and return types inside derives still validate against the caller's policy (`invokeDerive` passes `callerPolicy`); that pre-existing problem is a follow-up.
  - **Nested field shapes resolve in the same scope as the outer shape's site**, not in the namespace that declared the outer shape, as for every typed value; a qualified shape from another namespace whose fields name that namespace's local shapes is the same pre-existing problem.
  - **An omitted optional fact fails before `is` runs.** Evaluating the operand raises "identifier not found" today; once #134 lands an omitted optional fact resolves as undefined and `x is T` is `false`.
  - **`is not T` and `is T` both return a bool.** `not` produces a `box.Trinary`, which boxes a known outcome as `ValueBool` since [#135](https://github.com/sentrie-sh/sentrie/issues/135).
  - **`dict[T]` checks only the outer kind.** [[runtime.typeref_dict]] never validates values against `T`, so `{"k": "s"} is dict[number]` is `true`.
