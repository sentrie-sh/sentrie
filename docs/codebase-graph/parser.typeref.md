---
id: parser.typeref
type: Module / File
language: Go
file_path: parser/typeref.go
tags: type-syntax, constraints, nullable, recursive-descent
---

# Node: parser.typeref (Type Reference and Constraint Syntax)

## 1. Architectural Role & Intent
`parser/typeref.go` parses every type annotation in the language: the scalar keywords, the parameterised aggregates `list[T]`/`dict[T]`/`record[T, U]`, named shape references, the `?` nullable suffix, and the `@name(args)` constraint suffixes. It is the syntactic counterpart of [[ast.typeref]] and the point where **parse-time constraint validation** happens — an unknown or wrongly-arity'd constraint is rejected here, before any evaluation, by delegating to the AST's generated permission tables.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.typeref` | `CALLS` | [[ast.typeref]] | Constructs each concrete ref and calls `AddConstraint`, surfacing its rejections as parse errors. |
| `parser.typeref` | `CALLS` | [[parser.literal]] | Constraint arguments are parsed with the literal-only grammar, never as general expressions. |
| `parser.typeref` | `CALLS` | [[parser.fqn]] | An `Ident` head is parsed as a shape reference path. |
| `parser.typeref` | `READS_FROM` | [[parser.lookups]] | Uses `PRIMITIVE_TYPES` and `AGGREGATE_TYPES` to decide whether a type reference can start here. |
| [[parser.fact]] | `CALLS` | [[parser.typeref]] | Fact type annotations. |
| [[parser.let]] | `CALLS` | [[parser.typeref]] | Optional `let` annotations. |
| [[parser.shape]] | `CALLS` | [[parser.typeref]] | Simple-shape types and each complex-shape field type. |
| [[parser.cast]] | `CALLS` | [[parser.typeref]] | The target type of an `as` cast. |
| [[parser.typed_lambda]] | `CALLS` | [[parser.typeref]] | Parameter and return types. |
| [[constraints]] | `DEPENDS_ON` | [[ast.typeref]] | Evaluates the parsed constraint arguments against values at runtime. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseTypeRef(ctx, p) -> ast.TypeRef`
  - **Behavior:** Validates that the head is an `Ident` or one of the primitive/aggregate keywords, then dispatches:
    - `string` / `number` / `document` → the corresponding scalar ref.
    - `boolean` **or** `trinary` → both map to `TrinaryTypeRef`; `boolean` is a pure alias with no distinct node.
    - `Ident` → an FQN parsed into a `ShapeTypeRef`.
    - `list` / `dict` → constructed with a nil child, then a required `[ T ]` fills it in recursively.
    - `record` → constructed empty, then `[ T, U, … ]` fills a positional field list.

    Afterwards an optional `?` wraps the ref in `NullableTypeRef`, and a loop consumes `@constraint(...)` suffixes, each validated through `AddConstraint`.
  - **Side Effects:** Consumes tokens; mutates the constructed ref (children, span, constraints).
  - **Exceptions:** `expected one of %v, got %s` for an invalid head; `cannot add constraint %s: %s at %s` wrapping the arity/unknown-name errors from [[ast.typeref]]; `nil` on a missing bracket or failed child type.

- **Signature:** `parseTypeRefConstraint(ctx, p, _ ast.TypeRef) -> *ast.TypeRefConstraint`
  - **Behavior:** Parses `@name(lit, lit, …)`. Arguments go through [[parser.literal]], so only compile-time constants are accepted. The span covers `@` through the closing paren.
  - **Side Effects:** Consumes tokens; emits two debug logs.
  - **Exceptions:** Returns `nil` on a missing `@`, name, parenthesis, or a non-literal argument.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless functions; the returned ref is mutated during its own construction.
- **Performance/Scale Notes:** Two debug logs per constraint. Recursion depth follows type nesting, which is shallow in practice.
- **Dependencies Risk:**
  - **The head-validation list is rebuilt on every call.** Three slice allocations per type reference — harmless, but it means `PRIMITIVE_TYPES`/`AGGREGATE_TYPES` are copied constantly rather than checked in place.
  - **Nullable ordering is fixed: `T?@c(...)`, not `T@c(...)?`.** The `?` is consumed before the constraint loop, and because `NullableTypeRef` delegates `AddConstraint` to its inner type, constraints written after the `?` still attach to the underlying type. The two spellings are therefore not interchangeable in the source even though they produce the same constraint placement.
  - **`boolean` and `trinary` are indistinguishable after parsing.** A policy written with `boolean` produces a `TrinaryTypeRef`, so diagnostics and any type-name rendering will say "trinary".
  - **Commas are optional in `record[…]`.** The loop consumes a comma only if present, so `record[string number]` parses as two fields.
  - **Unterminated brackets rely on the head check, not EOF.** A missing `]` in a record loops until `advanceExpected` fails, reporting at end-of-input.
  - **Constraint rejection depends on generated tables.** An "unknown constraint" error most often means [[ast]]'s generated file is stale relative to [[constraints]] — see the generation hazard documented there.
