---
id: runtime.err_typedef
type: Module / File
language: Go
file_path: runtime/err_typedef.go
tags: errors, diagnostics, type-validation, sentinels
---

# Node: runtime.TypeRefErrors (Type and Constraint Error Sentinels)

## 1. Architectural Role & Intent
The error vocabulary for runtime type validation. It defines one public sentinel, `ErrTypeRef`, and two private sub-sentinels for the two ways a type annotation can fail - an unrecognised constraint name, and a constraint that ran but rejected the value - along with constructors that attach the constraint name and source span, and predicates that let callers distinguish the cases without string matching.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.err_typedef` | `DEPENDS_ON` | [[ast]] | Reads `ast.TypeRefConstraint` for the constraint name and span. |
| `runtime.err_typedef` | `DEPENDS_ON` | [[tokens]] | Accepts a `tokens.Range` position argument. |
| [[runtime.typeref]] | `CALLS` | [[runtime.err_typedef]] | Every per-kind type checker raises these. |
| [[runtime.executor]] | `READS_FROM` | [[runtime.err_typedef]] | Fact-type validation failures surface as these errors. |
| [[runtime.callable]] | `READS_FROM` | [[runtime.err_typedef]] | Parameter and return-type failures wrap these. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ErrTypeRef` (exported) / `errConstraintFailed`, `errUnknownConstraint` (unexported, both wrapping `ErrTypeRef`)
  - **Behavior:** A two-level sentinel hierarchy: `errors.Is(err, ErrTypeRef)` matches any type failure, while the specific predicates discriminate between the two sub-cases.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `ErrUnknownConstraint(c *ast.TypeRefConstraint) -> error`
  - **Behavior:** Reports a constraint name the runtime does not implement, with its source span.
  - **Side Effects:** None.
  - **Exceptions:** Panics on a nil constraint - it dereferences `c.Name`.

- **Signature:** `ErrConstraintFailed(pos tokens.Range, c *ast.TypeRefConstraint, err error) -> error`
  - **Behavior:** Reports a constraint that rejected a value, joining any underlying cause so it remains inspectable.
  - **Side Effects:** None.
  - **Exceptions:** Panics on a nil constraint.

- **Signature:** `IsUnknownConstraint(err) -> bool` / `IsConstraintFailed(err) -> bool`
  - **Behavior:** The intended discrimination API, wrapping `errors.Is` against the unexported sentinels.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Package-level immutable sentinels.
- **Performance/Scale Notes:** Negligible; construction only on the failure path.
- **Dependencies Risk:**
  - **The `pos` parameter of `ErrConstraintFailed` is never used.** The message takes its location from `c.Span()` instead, so callers passing a more precise position - the *value's* span rather than the *constraint's* - silently lose it. Every constraint failure therefore points at the type annotation, not at the offending data.
  - **An unknown constraint is a runtime error, not an index error.** [[ast]] validates constraint arguments at parse time against generated tables, but whether the runtime actually *implements* a constraint is only discovered when a value is validated - so a constraint the parser accepts can still fail at request time.
  - **The sub-sentinels are unexported**, so external packages must use the two predicates; `errors.Is` against a locally declared equivalent will not match.
  - **This is a separate error family from [[xerr]].** Type failures do not carry `xerr.ErrIndex` or similar, so code that classifies errors by `xerr` sentinel alone will not recognise them.
