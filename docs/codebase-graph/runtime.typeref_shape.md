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
| `runtime.typeref_shape` | `CALLS` | [[index.namespace]] | `VerifyShapeExported` gates cross-namespace references. |
| `runtime.typeref_shape` | `CALLS` | [[runtime.typeref]] | Recurses per field, and for the whole value when the shape is an alias. |
| `runtime.typeref_shape` | `DEPENDS_ON` | [[constraints]] | `constraints.ShapeContraintCheckers`. |
| `runtime.typeref_shape` | `DEPENDS_ON` | [[xerr]] | `ErrNamespaceNotFound`, `ErrShapeNotFound`. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_shape]] | Dispatched for `*ast.ShapeTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstShapeTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ShapeTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Four phases.
    1. **Resolution** - policy shapes first (the source comment notes this deliberately overrides the namespace), then namespace shapes, then, only when the reference has **more than two segments**, a cross-namespace lookup guarded by `VerifyShapeExported`.
    2. **Alias short-circuit** - if `shape.AliasOf` is set, revalidate the whole value against that type and return.
    3. **Field walk** - the value must be a dict; each declared field is looked up, missing optional fields are skipped, missing required fields error, a present-but-undefined field errors, and each present field recurses.
    4. **Constraints** - shape-level checkers.
  - **Side Effects:** Constraint argument evaluation; index lookups.
  - **Exceptions:** `xerr.ErrNamespaceNotFound`; `xerr.ErrShapeNotFound`; `value %v is not a shape at %s`; `field %s is required at %s`; `field %s cannot be undefined at %s`; `field '%s' is not valid: %w`; constraint errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless, but reads the committed index - so it depends on [[index.commit]] having hydrated shape composition first.
- **Performance/Scale Notes:** Cost is proportional to declared field count times nested depth. Resolution does up to three map lookups per validation with **no caching**, so validating a list of shapes re-resolves the same shape once per element.
- **Dependencies Risk:**
  - **Shape validation is open-world: undeclared fields are not rejected.** The walk iterates `shape.Model.Fields`, never the value's keys, so a value carrying extra members validates cleanly. A shape is a **minimum** contract, not an exact one - which is deliberate for external payloads but means a typo'd field name in an input is never reported.
  - **A two-segment reference can never resolve globally.** The cross-namespace branch requires `len(Parts) > 2`, so a reference like `ex/Foo` fails with `ErrShapeNotFound` even when that shape exists and is exported. Same segment-count constraint as the derive lookup in [[runtime.eval_call]].
  - **Lookup keys are the reference as written.** `p.Shapes` and `p.Namespace.Shapes` are keyed by bare name, but the lookup uses `typeRef.Ref.String()` - so a fully-qualified reference never matches the local maps and must go through the global path, while a bare name never reaches the global path at all.
  - **Missing optional field and present-undefined field behave differently.** An absent optional field is skipped; the same field present with an undefined value is a hard error. Since [[runtime.eval_access]] returns `Undefined` for anything absent, a value assembled by field access can fail where an omission would have passed.
  - **Policy shapes shadow namespace shapes here**, which is the opposite of the composition path described in [[index.shape]] - the two disagree, and that disagreement is filed as [#106](https://github.com/sentrie-sh/sentrie/issues/106).
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
