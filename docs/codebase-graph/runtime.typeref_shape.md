---
id: runtime.typeref_shape
type: Function / Endpoint
language: Go
file_path: runtime/typeref_shape.go
tags: type-validation, shapes, structural-typing, visibility, resolution
---

# Node: runtime.validateAgainstShapeTypeRef

## 1. Architectural Role & Intent
The most substantial validator: it resolves a shape reference through a three-tier scope chain (policy, then namespace, then cross-namespace with export verification), then validates the value structurally field by field. It handles both shape forms - an alias shape delegates straight back to its underlying type, while a complex shape drives a field walk.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_shape` | `READS_FROM` | [[index]] | `Policy.Shapes`, `Namespace.Shapes`, `Index.ResolveNamespace`, `Index.ResolveShape`. |
| `runtime.typeref_shape` | `READS_FROM` | [[index.shape]] | `Shape.ShapeScope` and `ShapeModelField.TypeScope` give the scope the alias and field types recurse in. |
| `runtime.typeref_shape` | `CALLS` | [[index.namespace]] | `VerifyShapeExported` gates cross-namespace references. |
| `runtime.typeref_shape` | `CALLS` | [[runtime.typeref]] | Recurses per field, and for the whole value when the shape is an alias. |
| `runtime.typeref_shape` | `CALLS` | [[runtime.typeref_constraints]] | `bindConstraints` before the kind check, `checkConstraints` after it. |
| `runtime.typeref_shape` | `DEPENDS_ON` | [[xerr]] | `ErrNamespaceNotFound`, `ErrShapeNotFound`. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_shape]] | Dispatched for `*ast.ShapeTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstShapeTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ShapeTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Binds its own constraints first ([[runtime.typeref_constraints]]: arguments evaluated and checked against the definition, #273), then four phases.
    1. **Resolution** (`resolveShapeTypeRef`, shared with the closest-member choice in [[runtime.typeref_union]]) - policy shapes first (the source comment notes this deliberately overrides the namespace), then namespace shapes, then, only when the reference has **more than two segments**, a cross-namespace lookup guarded by `VerifyShapeExported`.
    2. **Alias short-circuit** - if `shape.AliasOf` is set, revalidate the whole value against that type and return. The alias resolves in the shape's own scope (`Shape.ShapeScope`, see below), not in `p`.
    3. **Field walk** - the value must be a dict; each declared field is looked up, missing optional fields are skipped, missing required fields error, a present-but-undefined field errors, and each present field recurses in its declaring shape's scope (`ShapeModelField.TypeScope`), so a field composed in through `with` resolves where its base shape was declared.
    4. **Closed check** - unless `ShapeModel.Open` (the shape's own body ends in `...`), every key of the value must be a declared field, `with` bases' fields included; the first undeclared key in sorted order fails (#145). A key holding undefined counts as absent and is skipped, so `{ "id": "a", "extra": payload.missing } is Item` holds for a closed `Item { id: string }` (PROVISIONAL extension from the #148 decision "undefined is not a value"). It runs after the field walk, so a missing or wrongly typed field is reported in preference to an extra one.
    5. **Constraints** - shape-level checkers.
  - **Scope:** Only the reference itself resolves in `p`. Once the shape is found, its alias and field types (and their constraint arguments) resolve in the shape's scope: its policy for a policy-scoped shape, otherwise a `Policy` holding only its namespace, built once by [[index.shape]] `createShape`. An exported shape whose field names a sibling shape therefore works from any namespace, and a caller's same-named shape is never used ([#281](https://github.com/sentrie-sh/sentrie/issues/281)). The constraints written on the reference (`typeRef.GetConstraints()`) still evaluate in `p`. A hand-built shape with neither a policy nor a namespace falls back to `p`.
  - **Side Effects:** Constraint argument evaluation; index lookups.
  - **Exceptions:** `xerr.ErrNamespaceNotFound`; `xerr.ErrShapeNotFound`; `value %v is not a shape at %s`; `field %s is required at %s`; `field %s cannot be undefined at %s`; `field '%s' is not valid: %w`; `field %s is not declared in shape %s at %s - the shape is closed; ...`; constraint errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless, but reads the committed index - so it depends on [[index.commit]] having hydrated shape composition first.
- **Performance/Scale Notes:** Cost is proportional to declared field count times nested depth. Resolution does up to three map lookups per validation with **no caching**, so validating a list of shapes re-resolves the same shape once per element.
- **Dependencies Risk:**
  - **Shapes are closed by default (#145).** A value carrying a key the shape does not declare fails, nested shapes too, since each field recurses through this validator. A pack passing whole upstream payloads into a narrower shape must end the shape body with `...`. Openness is per declaration: an open `with` base does not open the shape that composes it, and an open child of a closed base accepts extra keys.
  - **A two-segment reference can never resolve globally.** The cross-namespace branch requires `len(Parts) > 2`, so a reference like `ex/Foo` fails with `ErrShapeNotFound` even when that shape exists and is exported. Same segment-count constraint as the derive lookup in [[runtime.eval_call]].
  - **Lookup keys are the reference as written.** `p.Shapes` and `p.Namespace.Shapes` are keyed by bare name, but the lookup uses `typeRef.Ref.String()` - so a fully-qualified reference never matches the local maps and must go through the global path, while a bare name never reaches the global path at all.
  - **Missing optional field and present-undefined field behave differently.** An absent optional field is skipped; the same field present with an undefined value is a hard error. Since [[runtime.eval_access]] returns `Undefined` for anything absent, a value assembled by field access can fail where an omission would have passed.
  - **A missing name rarely reaches here.** [[index.shape_refs]] rejects, at validation, a type naming a shape that is not visible from its site, so this lookup fails only for an unvalidated index or for a shape field's type: until #281 (PR #291) lands, fields are validated in the evaluating policy's scope `p`, while the index resolves them in the shape's own scope.
  - **Policy shapes shadow namespace shapes here**, which is the opposite of the composition path described in [[index.shape]] - the two disagree, and that disagreement is filed as [#106](https://github.com/sentrie-sh/sentrie/issues/106).
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
