---
id: ast.typeref
type: Class
language: Go
file_path: ast/typeref.go
tags: type-system, constraints, parse-time-validation, code-generation
---

# Node: ast.TypeRef (Type Reference Family)

## 1. Architectural Role & Intent
`TypeRef` is the sealed interface for every type annotation writable in Sentrie - `string`, `number`, `trinary`, `list[T]`, `dict[T]`, `record[…]`, `document`, a named `shape`, and the `T?` nullable wrapper - plus the constraint suffixes those types accept. Its distinguishing responsibility is that it is the **only AST family that validates itself at parse time**: each concrete type ref carries a generated table of which named constraints it permits and how many arguments each takes, so `string.between(1, 5)` on a `number` is rejected while the parser is still running rather than at evaluation.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `ast.typeref` | `INHERITS_FROM` | [[ast.node]] | `baseTypeRef` embeds `*baseNode`; `TypeRef` extends `Node` with `typeref()`, `GetConstraints()`, `AddConstraint()`. |
| `ast.typeref` | `DEPENDS_ON` | [[tokens]] | Constructors take a `tokens.Range`; `AddConstraint` extends `Rnge.To` over each accepted constraint. |
| `ast.typeref` | `DEPENDS_ON` | [[xerr]] | Returns `xerr.NotFoundError{}` when a constraint name is not in the type's permitted set. |
| `ast.typeref` | `DEPENDS_ON` | [[constraints]] | **Generated, build-time only.** The `gen*Constraints` maps in `typeref_constraint_args_gen.go` are emitted from the checker tables in [[constraints]] by `ast/gen.go`. |
| [[parser]] | `CALLS` | [[ast.typeref]] | `parser/typeref.go` builds the type ref then calls `AddConstraint` per parsed suffix, surfacing rejections as parse errors. |
| [[runtime]] | `DEPENDS_ON` | [[ast.typeref]] | `runtime/typeref*.go` type-switches on the concrete refs to coerce and validate [[box.value]] instances at cast and assignment boundaries. |
| [[index]] | `DEPENDS_ON` | [[ast.typeref]] | Resolves `ShapeTypeRef.Ref` FQNs against declared shapes during validation. |
| [[constraints]] | `CALLS` | [[ast.typeref]] | At evaluation time the checker for each `TypeRefConstraint` name is looked up and applied to the value. |

## 3. Interface Contracts & Public Surface

- **Signature:** `TypeRef` interface - `Node` + `typeref()` (unexported) + `GetConstraints() -> []*TypeRefConstraint` + `AddConstraint(*TypeRefConstraint) -> error`
  - **Behavior:** Sealed to the `ast` package. Every implementation is a struct embedding `*baseTypeRef`, so the constraint machinery is shared and uniform.
  - **Side Effects:** See `AddConstraint`.
  - **Exceptions:** See `AddConstraint`.

- **Signature:** `(*baseTypeRef).AddConstraint(c: *TypeRefConstraint) -> error`
  - **Behavior:** Validates the constraint against the receiver's permitted table, then appends it. Arity rules: an entry of `-1` means variadic and requires **at least one** argument; any other value requires an **exact** match.
  - **Side Effects:** Appends to the constraint slice and **mutates `Rnge.To`** so the type ref's span grows to cover its suffixes - this is why an error span for a constrained type covers the whole `string.min(3)` phrase.
  - **Exceptions:** `xerr.NotFoundError{}` when the name is absent from the table (surfaces as "unknown constraint"); `constraint %s requires at least 1 argument` for an empty variadic; `invalid number of arguments for constraint %s` for an arity mismatch.

- **Signature:** `(*baseTypeRef).GetConstraints() -> []*TypeRefConstraint`
  - **Behavior:** Returns the accumulated constraints in declaration order.
  - **Side Effects:** Returns the **live backing slice**, not a copy.
  - **Exceptions:** None.

- **Signature:** `TypeRefConstraint` struct - `{ Name string, Args []Expression }` + `NewTypeRefConstraint(name, args, span)`
  - **Behavior:** A parsed constraint suffix. `Args` are arbitrary AST expressions, not literals, so constraint arguments are evaluated by [[runtime]] rather than folded at parse time.
  - **Side Effects:** None.
  - **Exceptions:** None; the constructor never validates - validation happens only on `AddConstraint`.

- **Signature:** Scalar refs - `NewStringTypeRef(span)`, `NewNumberTypeRef(span)`, `NewTrinaryTypeRef(span)`, `NewDocumentTypeRef(span)`
  - **Behavior:** Leaf types. Each wires its own generated constraint table (`genStringConstraints`, `genNumberConstraints`, …). `String()` renders the canonical keyword.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** Composite refs - `NewListTypeRef(elem, span)`, `NewDictTypeRef(valueType, span)`, `NewRecordTypeRef(fields, span)`, `NewShapeTypeRef(ref: *FQN, span)`
  - **Behavior:** `list[T]` and `dict[T]` carry a single child type (dicts are string-keyed, so only the value type is expressible). `record[…]` carries a positional field list. `ShapeTypeRef` is a **by-name reference** holding an `FQN` that stays unresolved until [[index]] links it to a `ShapeStatement`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewNullableTypeRef(inner, span)` + `IsNullableTypeRef(t) -> bool` + `UnwrapNullableTypeRef(t) -> TypeRef`
  - **Behavior:** The `T?` wrapper. It **delegates** `GetConstraints`/`AddConstraint` to its inner type - its own permitted table is empty - so constraints attach to the underlying type and the wrapper only adjusts its own span. `UnwrapNullableTypeRef` is the idempotent accessor consumers should use before type-switching.
  - **Side Effects:** `AddConstraint` mutates the **inner** ref and then the wrapper's span.
  - **Exceptions:** Propagates the inner type's errors unchanged.

## 4. Operational Context & Gotchas
- **Statefulness:** Mutable during parsing only (constraint accumulation and span extension), read-only thereafter. Not safe for concurrent construction; safe to read concurrently once the parse completes.
- **Performance/Scale Notes:** Constraint lookup is a single map probe per suffix, so parse-time cost is negligible. `String()` on composite refs recurses and allocates; it is a diagnostic helper, not a hot path. The `gen*Constraints` maps are package-level and shared by every instance of a given type - they must never be written to.
- **Dependencies Risk:** No external failure domain. The hazards:
  - **Generation drift.** The permitted-constraint tables are generated from [[constraints]]; adding a checker there without re-running `go generate` (which itself **panics unless `GIT_USER_NAME` and `GIT_USER_EMAIL` are set**) makes the parser reject a constraint the runtime fully supports. This is the single most likely cause of a spurious "unknown constraint" error.
  - **Nullable delegation surprises.** `NullableTypeRef.GetConstraints()` returns the *inner* type's constraints, so code that checks whether a wrapper "has constraints" sees through the wrapper. Always call `UnwrapNullableTypeRef` before matching on concrete type, and never assume a `NullableTypeRef` owns its own constraint list.
  - **Unresolved shape references.** `ShapeTypeRef.Ref` is a raw FQN - dangling until [[index]] validates it. Anything consuming a type ref straight from the parser must treat shape references as unverified.
  - **Live slice exposure.** `GetConstraints()` hands back the internal slice; appending to the result corrupts the node.
