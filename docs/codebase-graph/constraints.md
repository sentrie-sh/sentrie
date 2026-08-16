---
id: constraints
type: System / Package
language: Go
file_path: constraints/
tags: validation, type-refinement, predicates, registry
---

# Node: constraints (Type Refinement Predicates)

## 1. Architectural Role & Intent
The registry of predicates that refine a type beyond its kind — `string@email`, `number@range(1,10)`, `list@not_empty`. Each value kind has its own table, and the runtime type validators select a table by the type ref they are checking. This is where a declarative type gains semantic meaning, and it is the mechanism by which policy authors express input contracts without writing rule logic.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `constraints` | `DEPENDS_ON` | [[box]] | Every checker receives and inspects `box.Value`. |
| `constraints` | `DEPENDS_ON` | [[index.package]] | Checkers take a `*index.Policy` — currently unused by every checker, but part of the contract. |
| `constraints` | `DEPENDS_ON` | `ext.google_uuid` | The `uuid` string constraint. |
| [[runtime.typeref_string]] | `READS_FROM` | `constraints` | `StringContraintCheckers`. |
| [[runtime.typeref_number]] | `READS_FROM` | `constraints` | `NumberContraintCheckers`. |
| [[runtime.typeref_trinary]] | `READS_FROM` | `constraints` | `TrinaryConstraintCheckers`. |
| [[runtime.typeref_list]] | `READS_FROM` | `constraints` | `ListContraintCheckers`. |
| [[runtime.typeref_dict]] | `READS_FROM` | `constraints` | `DictContraintCheckers` — **empty**. |
| [[runtime.typeref_document]] | `READS_FROM` | `constraints` | `DocumentContraintCheckers` — **empty**. |
| [[runtime.typeref_record]] | `READS_FROM` | `constraints` | `RecordContraintCheckers` — **empty**. |
| [[runtime.typeref_shape]] | `READS_FROM` | `constraints` | `ShapeContraintCheckers` — **empty**. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ConstraintChecker = func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error`
  - **Behavior:** Returns nil when the value satisfies the constraint, an error describing the violation otherwise.
  - **Side Effects:** None — all checkers are pure.
  - **Exceptions:** The violation error.

- **Signature:** `ConstraintDefinition` — `{Name, NumArgs, Checker}`
  - **Behavior:** `NumArgs` declares the expected argument count, though each checker also re-checks its own argument count defensively.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `StringContraintCheckers` — 20 constraints
  - **Behavior:** Length (`length`, `minlength`, `maxlength`), pattern (`regexp`, `starts_with`, `ends_with`, `has_substring`, `not_has_substring`), format (`email`, `url`, `uuid`), character class (`alphanumeric`, `alpha`, `numeric`, `lowercase`, `uppercase`), and set membership (`trimmed`, `not_empty`, `one_of`, `not_one_of`).
  - **Side Effects:** None.
  - **Exceptions:** Kind errors, argument-count errors, violation errors.

- **Signature:** `NumberContraintCheckers` — 19 constraints
  - **Behavior:** Bounds (`min`, `max`, `gt`, `lt`, `range`), equality (`eq`, `neq`), membership (`in`, `not_in`), parity and divisibility (`even`, `odd`, `multiple_of`), sign (`positive`, `negative`, `non_negative`, `non_positive`), and float classification (`finite`, `infinite`, `nan`).
  - **Side Effects:** None.
  - **Exceptions:** As above.

- **Signature:** `TrinaryConstraintCheckers` — 5 constraints
  - **Behavior:** `not_unknown`, `eq`, `neq`, `is_true`, `is_false`. `not_unknown` is the one that matters most: it is how a policy demands a definite answer.
  - **Side Effects:** None.
  - **Exceptions:** As above.

- **Signature:** `ListContraintCheckers` — 1 constraint
  - **Behavior:** `not_empty` only.
  - **Side Effects:** None.
  - **Exceptions:** As above.

## 4. Operational Context & Gotchas
- **Statefulness:** Package-level maps built at initialisation, read-only thereafter.
- **Performance/Scale Notes:** Checkers are cheap, but they run on **every** validation. Inside [[runtime.typeref_list]] that means once per element, and the constraint's argument expressions are re-evaluated per element too. The `regexp` constraint compiles its pattern on each invocation rather than caching, so a regex constraint on a list element type recompiles per element.
- **Dependencies Risk:**
  - **Four of the eight tables are empty.** `dict`, `document`, `record`, and `shape` have no constraints at all. Because [[runtime.err_typedef]] turns an unrecognised constraint into a **runtime** error, and nothing validates constraint names at index time, any constraint written on one of these types passes `sentrie validate` and then fails at decision time with `unknown constraint`. Filed as an issue.
  - **`list` has only `not_empty`.** There is no size, min-length, or unique constraint, so list contracts must be expressed as rule logic instead.
  - **Constraint names are not validated statically.** A typo like `string@emial` is indistinguishable from a constraint that has not been implemented yet — both surface only when a decision is requested. This is the single highest-value fix in the constraint pipeline.
  - **String length constraints count bytes, not runes.** `len(s)` on a Go string is its byte length, so a `length(5)` constraint rejects a five-character string containing any multi-byte character. For a policy engine handling international identifiers this is a correctness issue, not a nit.
  - **`NumArgs` is declared but not enforced by the framework.** Each checker re-validates its own argument count, so the field is documentation that can silently disagree with the implementation.
  - **The `*index.Policy` parameter is unused by every checker.** It exists so a constraint could one day resolve policy-scoped context; today it is dead weight in the signature.
