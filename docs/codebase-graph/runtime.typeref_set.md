---
id: runtime.typeref_set
type: Function / Endpoint
language: Go
file_path: runtime/typeref_set.go
tags: type-validation, constraints, collections, sets
---

# Node: runtime.validateAgainstSetTypeRef

## 1. Architectural Role & Intent
Validates a value against `set[T]` ([#162](https://github.com/sentrie-sh/sentrie/issues/162)) and converts a list bound to a `set[T]` declaration into a set. A set is a tagged list (see [[box.set]]), so validation mirrors [[runtime.typeref_list]] with one extra rule: a plain list must hold distinct elements that can be held in a set. A JSON array reaching a `set[T]` fact is therefore validated, never silently deduplicated.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_set` | `CALLS` | [[runtime.typeref]] | Recurses once per element against `typeRef.ElemType`. |
| `runtime.typeref_set` | `CALLS` | [[box.set]] | `NewSetUnique` checks a plain list for duplicates and unsupported elements, and builds the bound set. |
| `runtime.typeref_set` | `DEPENDS_ON` | [[constraints]] | `constraints.ListContraintCheckers`; a set takes the list constraints. |
| `runtime.typeref_set` | `CALLS` | [[runtime.typeref_trinary]] | `bindTypedValue` composes `bindTruthValue` with `bindSetValue`. |
| `runtime.typeref_set` | `CALLS` | [[runtime.err_typedef]] | `errMismatch`, constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_set]] | Dispatched for `*ast.SetTypeRef`. |
| [[runtime.exec_ctx]] | `CALLS` | [[runtime.typeref_set]] | `InjectFact` binds every fact and default through `bindTypedValue`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstSetTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.SetTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires a list; validates each element against `T`; for an untagged list also requires the elements, bound as `T` is, to be distinct and orderable (so `[impT, true]` is a duplicate in a `set[boolean]`); then runs list constraints.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value %v is not a set at %s - expected set` and `value %v is not a valid set at %s: <duplicate or element error>` are mismatches (so `[1, 1] is set[number]` is false); `item is not valid at %s: %w`; constraint errors.

- **Signature:** `bindSetValue(v, typeRef) -> box.Value` / `bindTypedValue(v, typeRef) -> box.Value`
  - **Behavior:** For a (nullable) `set[T]` type and a plain list that can be a set, returns the set, each element first bound as `T` is (`setElementsAs`), so an imported decision in a `set[boolean]` is its state and an unknown state is rejected; otherwise `v` unchanged, for validation to report. `bindTypedValue` is the single bind step used by `let` bindings ([[runtime.eval_ident]]), derive parameters and returns ([[runtime.derive_invoke]]), lambda returns ([[runtime.callable]]) and `InjectFact` ([[runtime.exec_ctx]]).
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** An untagged list is sorted once to check for duplicates on every validation; a fact bound at injection is tagged, so the rule-entry validation of facts skips that sort.
- **Dependencies Risk:**
  - **Only the top-level declared type converts.** A `set[T]` nested in a list, dict, record or shape field is validated (distinct, typed elements) but keeps the list it arrived as, so it serialises in input order and `IsSet` is false. The set builtins accept lists, so set operations on such a field still work.
  - **`set[T]` accepts a distinct plain list**, so `[1, 2] is set[number]` is true even though the value is not tagged.
