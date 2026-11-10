---
id: index.shape
type: Class
language: Go
file_path: index/shape.go
tags: type-model, composition, hydration, records
---

# Node: index.Shape (Shape Model and Composition)

## 1. Architectural Role & Intent
`Shape` models both forms of a shape declaration - a simple type alias (`AliasOf`) and a complex record (`Model` with fields) - and owns the `with` composition algorithm that copies a base shape's fields into a derived one. Composition is performed lazily via `resolveDependency`, which takes its base from `resolveWithBase`, the one `with` resolution the shape cycle check in [[index.validate]] also uses, guarded by an atomic `hydrated` flag so each shape is flattened exactly once, and it is driven parents-first (any `with` depth) by [[index.commit]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.shape` | `DEPENDS_ON` | [[ast]] | Wraps `ast.ShapeStatement`, `ast.ShapeField`, `ast.TypeRef`, and builds FQNs via `ast.CreateFQN`. |
| `index.shape` | `DEPENDS_ON` | [[xerr]] | Emits `ErrIndex`-wrapped composition errors and inspects `xerr.NotFoundError` to distinguish a namespace miss from a real failure. |
| `index.shape` | `CALLS` | [[index.resolve]] | `resolveWithBase` calls `VerifyShapeExported` on a qualified base's namespace and builds the suggested qualified names with `ShapeFQN`. |
| `index.shape` | `READS_FROM` | [[index.namespace]] | `resolveWithBase` reads `Shapes` and `ShapeExports` of the shape's own namespace, the namespace a qualified base names, and every other namespace when an unqualified base is not local. |
| [[index.commit]] | `CALLS` | [[index.shape]] | Invokes `resolveDependency` parents first, the reverse of the shape-DAG sort order. |
| [[index.validate]] | `CALLS` | [[index.shape]] | `detectShapeCycle` adds each shape's `with` edge from `resolveWithBase`, so the cycle check and hydration agree on every parent. |
| [[index.builtin_kind]] | `CALLS` | [[index.shape]] | `lookupShapeFieldTypeReadOnly` takes a `with` parent from `resolveWithBase`, so the static kind check sees the fields hydration will copy in. |
| [[index.builtin_kind]] | `READS_FROM` | [[index.shape]] | Walks `Model.Fields` to resolve field-access kinds statically. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Shape` struct - `{ Statement, Namespace, Policy, Name, FQN, Model *ShapeModel, AliasOf ast.TypeRef, FilePath }` plus unexported `shapeScope *Policy` and `hydrated uint32`
  - **Behavior:** `Model` and `AliasOf` are **mutually exclusive**: a complex shape has `Model` and a nil `AliasOf`, a simple shape the reverse. `Policy` is nil for namespace-scoped shapes.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `(*Shape).ShapeScope() -> *Policy`
  - **Behavior:** The scope the shape's alias and field types resolve in: its policy, or for a namespace-level shape a `Policy` holding only its namespace (no policy shapes, no rules). Mirrors `Derive.ShapeScope` ([[index.derive]]). Set once by `createShape` and never written afterwards, so concurrent evaluations read it without a lock; a hand-built shape falls back to its policy, then a fresh namespace-only scope, then nil when it has neither.
  - **Side Effects:** None for an indexed shape; allocates per call only for a hand-built one.
  - **Exceptions:** None.

- **Signature:** `ShapeModel` - `{ WithFQN *ast.FQN, Fields map[string]*ShapeModelField }` / `ShapeModelField` - `{ Node, Name, Optional, TypeRef, DeclaredIn *Shape }`
  - **Behavior:** Fields are a map, so **declaration order is lost**. `Optional` here is field optionality, distinct from a nullable type ref. `DeclaredIn` is the shape whose declaration holds the field; `resolveDependency` copies field pointers, so a field composed in through `with` keeps its base shape. `(*ShapeModelField).TypeScope(fallback)` is `DeclaredIn.ShapeScope()`, or `fallback` when that is nil. [[runtime.typeref_shape]] and [[index.builtin_kind]] resolve the field's type there ([#281](https://github.com/sentrie-sh/sentrie/issues/281)).
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `createShape(ns *Namespace, p *Policy, stmt *ast.ShapeStatement) -> (*Shape, error)`
  - **Behavior:** Builds the FQN from the policy when policy-scoped, otherwise the namespace; builds `shapeScope`; populates `Model` for complex shapes (recording `With` and each field, with `DeclaredIn` set to the new shape) or `AliasOf` for simple ones. Fields with an empty name are skipped silently.
  - **Side Effects:** Allocation only.
  - **Exceptions:** `duplicate shape field '%s' at %s` - note this one is **not** wrapped in `xerr.ErrIndex`, unlike its sibling errors.

- **Signature:** `(*Shape).resolveDependency(idx *Index) -> error`
  - **Behavior:** No-ops if already hydrated, if there is no model, or if there is no `with` clause. Otherwise takes the base from `resolveWithBase` and copies its fields in.
  - **Side Effects:** **Mutates `s.Model.Fields`**; sets `hydrated` in a `defer`, so the flag is set even on the error paths.
  - **Exceptions:** Any `resolveWithBase` error; composing with an alias shape; duplicate field between the derived shape and **any ancestor** in its `with` chain, not only the direct parent (the base's field map already holds every ancestor's fields when it is copied).

- **Signature:** `(*Index).resolveWithBase(s *Shape) -> (*Shape, error)`
  - **Behavior:** The one resolution of a `with` base (#307, #308). `(nil, nil)` when `s` has no model or no `with` clause. A **qualified** base (two or more segments) resolves only in the namespace its leading segments name, by the last segment, and that namespace must export it, except when it is `s`'s own namespace, which is exempt from the export check. An **unqualified** base resolves in `s.Policy`'s shapes, then `s.Namespace`'s, and the first match wins. An unqualified base that only another namespace exports is an error naming the qualified form(s) to write instead.
  - **Side Effects:** None.
  - **Exceptions:** `shape '<ref>' not found at <span>` (missing namespace, missing shape, or an unqualified name no scope or export holds); `shape '<ref>' not exported at <span>`; `shape '<ref>' at <span> is exported by another namespace and must be qualified: use <ns>/<name>[ or ...]`. All wrap `xerr.ErrIndex`.

- **Signature:** `(*Shape).String()` / `Span()`
  - **Behavior:** FQN rendering and span passthrough for DAG diagnostics.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** **Mutated in place during commit.** A shape read before commit has only its own fields; after commit it carries inherited ones too. Consumers must run after [[index.commit]].
- **Performance/Scale Notes:** Resolution is map lookups, except an unqualified base that is not local, which scans every namespace once (only on the way to an error).
- **Dependencies Risk:**
  - **`hydrated` is set in a `defer`, including on failure.** A shape whose composition errored is still marked hydrated, so a retry would silently no-op. Since commit aborts on the first error this is currently benign, but it makes partial recovery impossible.
  - **Hydration and the cycle check share `resolveWithBase`.** The static kind check in [[index.builtin_kind]] uses it too. Any second resolver for a `with` base can disagree with them, which is how #307 and #308 arose.
  - **Topological ordering is assumed, not enforced here.** The field copy assumes the base shape is already hydrated. That assumption holds only because [[index.commit]] walks the DAG in order; calling `resolveDependency` directly can copy a half-flattened base.
  - **Aliases cannot be composed with**, and the error points at the *base* shape's span rather than the offending `with` clause.
