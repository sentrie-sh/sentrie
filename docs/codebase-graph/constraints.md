---
id: constraints
type: System / Package
language: Go
file_path: constraints/
tags: validation, type-refinement, predicates, registry
---

# Node: constraints (Type Refinement Predicates)

## 1. Architectural Role & Intent
The registry of predicates that refine a type beyond its kind - `string@email`, `number@range(1,10)`, `list@not_empty`. Each value kind has its own table, and the runtime type validators select a table by the type ref they are checking. This is where a declarative type gains semantic meaning, and it is the mechanism by which policy authors express input contracts without writing rule logic.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `constraints` | `LAYERED_ON` | [[box]] | Every checker receives and inspects `box.Value`. |
| `constraints` | `LAYERED_ON` | [[ast]] | `ForTypeRef` maps a type ref node to its table. |
| [[index.constraint_args]] | `READS_FROM` | `constraints` | `ForTypeRef` and `CheckArgs`, the static argument check run by `sentrie validate`. |
| [[runtime.typeref_constraints]] | `READS_FROM` | `constraints` | `ForTypeRef`, `CheckArgs` before the value's kind, then `Checker`. |
| `constraints` | `IMPORTS` | `ext.google.uuid` | The `uuid` string constraint. |
| [[runtime.typeref_string]] | `READS_FROM` | `constraints` | `StringContraintCheckers`. |
| [[runtime.typeref_number]] | `READS_FROM` | `constraints` | `NumberContraintCheckers`. |
| [[runtime.typeref_trinary]] | `READS_FROM` | `constraints` | `TrinaryConstraintCheckers`. |
| [[runtime.typeref_list]] | `READS_FROM` | `constraints` | `ListContraintCheckers`. |
| [[runtime.typeref_dict]] | `READS_FROM` | `constraints` | `DictContraintCheckers` - **empty**. |
| [[runtime.typeref_document]] | `READS_FROM` | `constraints` | `DocumentContraintCheckers` - **empty**. |
| [[runtime.typeref_record]] | `READS_FROM` | `constraints` | `RecordContraintCheckers` - **empty**. |
| [[runtime.typeref_shape]] | `READS_FROM` | `constraints` | `ShapeContraintCheckers` - **empty**. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ConstraintChecker = func(ctx context.Context, val box.Value, args []box.Value) error`
  - **Behavior:** Returns nil when the value satisfies the constraint, an error describing the violation otherwise. Every table entry is built with `define`, whose `Checker` runs `CheckArgs` before the value check, so a broken argument is reported whatever the value's kind.
  - **Side Effects:** None - all checkers are pure.
  - **Exceptions:** The violation error. A failure caused by the constraint's own arguments (wrong count, wrong kind, an invalid regexp, a zero divisor) matches `ErrInvalidArgument` under `errors.Is`, with its message unchanged; a failure caused by the value does not.

- **Signature:** `ErrInvalidArgument`
  - **Behavior:** Sentinel that separates a broken constraint from a non-conforming value. [[runtime.err_typedef]] reads it so `x is T` ([[runtime.eval_is_type]]) propagates a broken constraint as an error instead of returning `false`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `ConstraintDefinition` - `{Name, NumArgs, Arg, Args, Checker}`
  - **Behavior:** The argument metadata. `NumArgs` is the exact count, or `-1` for one or more. `Arg` checks each argument (number, string, boolean, a regexp that compiles, a non-zero divisor, a number or list of numbers) and is nil exactly when the constraint takes no arguments. `Args`, when set, checks rules across arguments (`range`: minimum not above maximum).
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(ConstraintDefinition).CheckArgs(args []box.Value) -> error`
  - **Behavior:** Checks the count, then each argument, then the cross-argument rule, without a value. It is the single argument check: the index runs it when a pack is validated ([[index.constraint_args]]), the runtime runs it before a value's kind ([[runtime.typeref_constraints]]), and every `Checker` runs it first.
  - **Side Effects:** None.
  - **Exceptions:** An error matching `ErrInvalidArgument`, with the message the checkers have always used (`expected number, got string`, `invalid regexp pattern: ...`, `divisor cannot be zero`, `range minimum 5 is greater than maximum 1`, `<name> constraint requires 1 argument`); a constraint given arguments it does not take says `<name> constraint takes no arguments`.

- **Signature:** `ForTypeRef(typeRef ast.TypeRef) -> map[string]ConstraintDefinition`
  - **Behavior:** The table for a type ref's own constraints, nil for an unknown kind. Callers unwrap a nullable first.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `StringContraintCheckers` - 20 constraints
  - **Behavior:** Length (`length`, `minlength`, `maxlength`), pattern (`regexp`, `starts_with`, `ends_with`, `has_substring`, `not_has_substring`), format (`email`, `url`, `uuid`), character class (`alphanumeric`, `alpha`, `numeric`, `lowercase`, `uppercase`), and set membership (`trimmed`, `not_empty`, `one_of`, `not_one_of`).
  - **Side Effects:** None.
  - **Exceptions:** Kind errors, argument-count errors, violation errors.

- **Signature:** `NumberContraintCheckers` - 19 constraints
  - **Behavior:** Bounds (`min`, `max`, `gt`, `lt`, `range`), equality (`eq`, `neq`), membership (`in`, `not_in`), parity and divisibility (`even`, `odd`, `multiple_of`), sign (`positive`, `negative`, `non_negative`, `non_positive`), and float classification (`finite`, `infinite`, `nan`).
  - **Side Effects:** None.
  - **Exceptions:** As above.

- **Signature:** `TrinaryConstraintCheckers` - 5 constraints
  - **Behavior:** `not_unknown`, `eq`, `neq`, `is_true`, `is_false`. `not_unknown` is the one that matters most: it is how a policy demands a definite answer.
  - **Side Effects:** None.
  - **Exceptions:** As above.

- **Signature:** `ListContraintCheckers` - 1 constraint
  - **Behavior:** `not_empty` only.
  - **Side Effects:** None.
  - **Exceptions:** As above.

## 4. Operational Context & Gotchas
- **Statefulness:** Package-level maps built at initialisation, read-only thereafter.
- **Performance/Scale Notes:** Checkers are cheap, but they run on **every** validation. Inside [[runtime.typeref_list]] that means once per element, and the constraint's argument expressions are re-evaluated per element too. The `regexp` constraint caches each compiled pattern in a package-level map keyed by the pattern, so `CheckArgs` (at validation, before each value, and inside the `Checker`) and the match share one compilation per process; patterns come from source literals, so the cache is bounded by the pack.
- **Dependencies Risk:**
  - **Four of the eight tables are empty.** `dict`, `document`, `record`, and `shape` have no constraints at all. Because [[runtime.err_typedef]] turns an unrecognised constraint into a **runtime** error, and nothing validates constraint names at index time, any constraint written on one of these types passes `sentrie validate` and then fails at decision time with `unknown constraint`. Filed as [#113](https://github.com/sentrie-sh/sentrie/issues/113).
  - **`list` has only `not_empty`.** There is no size, min-length, or unique constraint, so list contracts must be expressed as rule logic instead.
  - **Constraint names are not validated statically.** A typo like `string@emial` is indistinguishable from a constraint that has not been implemented yet - both surface only when a decision is requested. This is the single highest-value fix in the constraint pipeline.
  - **String length constraints count bytes, not runes.** `len(s)` on a Go string is its byte length, so a `length(5)` constraint rejects a five-character string containing any multi-byte character. For a policy engine handling international identifiers this is a correctness issue, not a nit.
  - **New argument checks must use `invalidArgf`.** An argument failure built with plain `fmt.Errorf` reads as a value mismatch, so `x is not T` with that broken constraint would be `true`.
  - **New argument rules belong in `Arg` or `Args`, not in a checker body.** That is what keeps the static check, the runtime and a direct `Checker` call in agreement (#273); a checker body may assume its arguments passed `CheckArgs`.
  - **`NumArgs` is enforced twice from one value.** `CheckArgs` checks it, and the parser checks it earlier through [[ast]]'s generated tables, which `go generate` copies from `NumArgs`; a stale generated file is the only way the two can disagree.
  - **The `in` constraint cannot be written in source.** `in` lexes as a keyword and the parser expects an identifier after `@`, so `number@in(...)` is a parse error; `not_in` is unaffected.
  - **A checker has no policy context.** The `*index.Policy` parameter, unused by every checker, was dropped so [[index]] can import this package for the static argument check.
