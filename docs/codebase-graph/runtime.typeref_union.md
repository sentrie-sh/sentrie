---
id: runtime.typeref_union
type: Function / Endpoint
language: Go
file_path: runtime/typeref_union.go
tags: type-validation, unions, top-type, error-reporting
---

# Node: runtime.validateAgainstUnionTypeRef (Union and `any` Validation)

## 1. Architectural Role & Intent
Validates a value against a union `T1 | T2 | …` and against the top type `any` (#165). A union passes when any member passes. Its other job is the failure message: rather than listing every member's failure, it reports the one member the value came closest to, chosen by a fixed rule so the same input always yields the same diagnostic.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_union` | `CALLS` | [[runtime.typeref]] | Validates the value against each member. |
| `runtime.typeref_union` | `CALLS` | [[runtime.typeref_shape]] | `resolveShapeTypeRef` to read a shape member's outer kind. |
| `runtime.typeref_union` | `DEPENDS_ON` | [[runtime.err_typedef]] | `isTypeMismatch` and the `typeRefFailure` mismatch marker. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_union]] | Dispatched for `*ast.UnionTypeRef` and `*ast.AnyTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstAnyTypeRef(v box.Value, pos tokens.Range) -> error`
  - **Behavior:** Accepts every value except undefined, which is the absence of a value and which no type accepts.
  - **Side Effects:** None.
  - **Exceptions:** `value is undefined at %s - expected any`, a mismatch.
- **Signature:** `validateAgainstUnionTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.UnionTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Validates `v` against every member in source order and returns `nil` on the first that passes. Otherwise: the first member error that is not a mismatch is returned unchanged (an unknown shape, a constraint with invalid arguments). When every member is a mismatch, the closest member is the first, in source order, whose outer kind accepts `v` (`outerKindAccepts`); its error is wrapped as `value %v does not match <union> at %s: closest member <member>: <err>`. When no member's outer kind accepts `v`: `value %v is not <union> at %s - expected one of its members`.
  - **Side Effects:** Constraint-argument evaluation in every member tried.
  - **Exceptions:** As above. Both union-level messages carry the mismatch marker.
- **Signature:** `outerKindAccepts(exec, p, v, t) -> bool`
  - **Behavior:** Whether `v` passes `t`'s outermost kind check only: string, number, bool or trinary (or a decision envelope), list for `list`/`record`, dict for `dict`/`document`/complex shapes; a shape alias is read through; `T?` also accepts null; `any` accepts any defined value; a union accepts what any member accepts. An unresolvable shape accepts nothing.
  - **Side Effects:** Shape lookups.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** A failing union validates the value against every member, so a union of large shapes costs the sum of its members; a passing one stops at the first match.
- **Dependencies Risk:**
  - **A match beats a broken member.** `Missing | string` passes for `"x"` even though `Missing` cannot be resolved; a broken member only surfaces when nothing matched. [[index.is_resolve]] catches unknown shapes in `x is T` at index time; typed facts and lets are checked only here.
  - **"Closest" is coarse.** Among several members of the right outer kind (two complex shapes for a dict) the first in source order wins, however many fields each matched. Discriminated unions are out of scope.
  - **Member renderings omit constraints**, as every type ref's `String()` does, so `string@minlength(3) | number` prints as `string | number`; the wrapped member error names the failed constraint.
  - **Truth binding ignores unions.** `bindTruthValue` in [[runtime.typeref_trinary]] turns a decision envelope into its state only for a bare `trinary`/`boolean` type, not for a union with a trinary member.
