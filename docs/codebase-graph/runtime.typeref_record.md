---
id: runtime.typeref_record
type: Function / Endpoint
language: Go
file_path: runtime/typeref_record.go
tags: type-validation, tuples, positional-typing
---

# Node: runtime.validateAgainstRecordTypeRef

## 1. Architectural Role & Intent
Validates a value against a `record` type - Sentrie's fixed-arity positional tuple. Unlike a shape, which matches by field name, a record matches strictly by **position and length**, which makes it the right type for key/value pairs and other small ordered groupings.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_record` | `CALLS` | [[runtime.typeref]] | Recurses per position against the corresponding field type. |
| `runtime.typeref_record` | `DEPENDS_ON` | [[constraints]] | `constraints.RecordContraintCheckers`. |
| `runtime.typeref_record` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_record]] | Dispatched for `*ast.RecordTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstRecordTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.RecordTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires a list, requires `len(entries) == len(typeRef.Fields)` **exactly**, then validates each entry against the type at the same index, then runs record constraints.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value %v is not a record`; `fields length mismatch: %v`; `%v is not a valid record field: %w`; constraint errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** O(arity) - records are small by construction, so cost is negligible.
- **Dependencies Risk:**
  - **A record is represented as a list at runtime.** There is no distinct value kind, so a record and a list are indistinguishable once boxed; the type ref is the only thing that separates them. A list of the right length and element types validates as a record.
  - **Length must match exactly** - no optional trailing positions, no variadic tail. Adding a field to a record is a breaking change for every producer.
  - **The error messages carry no position information.** Both the length mismatch and the per-field wrapper print the whole value with `%v` and omit which index failed; the source carries `TODO` markers on both. For a record of similarly-typed fields the diagnostic is close to useless.
  - **`pos` is accepted and passed down but never used in this function's own messages**, so record errors are less locatable than list or shape errors.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
