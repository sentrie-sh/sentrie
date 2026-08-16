---
id: index.shape
type: Class
language: Go
file_path: index/shape.go
tags: type-model, composition, hydration, records
---

# Node: index.Shape (Shape Model and Composition)

## 1. Architectural Role & Intent
`Shape` models both forms of a shape declaration — a simple type alias (`AliasOf`) and a complex record (`Model` with fields) — and owns the `with` composition algorithm that copies a base shape's fields into a derived one. Composition is performed lazily via `resolveDependency`, guarded by an atomic `hydrated` flag so each shape is flattened exactly once, and it is driven in topological order by [[index.commit]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.shape` | `DEPENDS_ON` | [[ast]] | Wraps `ast.ShapeStatement`, `ast.ShapeField`, `ast.TypeRef`, and builds FQNs via `ast.CreateFQN`. |
| `index.shape` | `DEPENDS_ON` | [[xerr]] | Emits `ErrIndex`-wrapped composition errors and inspects `xerr.NotFoundError` to distinguish a namespace miss from a real failure. |
| `index.shape` | `CALLS` | [[index.resolve]] | `resolveDependency` falls back to `idx.ResolveShape` across every namespace when a base shape is not local. |
| `index.shape` | `CALLS` | [[index.namespace]] | Calls `VerifyShapeExported` before composing with a shape from another namespace. |
| [[index.commit]] | `CALLS` | [[index.shape]] | Invokes `resolveDependency` in shape-DAG topological order. |
| [[index.validate]] | `READS_FROM` | [[index.shape]] | Builds the shape DAG from `Model.WithFQN` edges. |
| [[index.builtin_kind]] | `READS_FROM` | [[index.shape]] | Walks `Model.Fields` to resolve field-access kinds statically. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Shape` struct — `{ Statement, Namespace, Policy, Name, FQN, Model *ShapeModel, AliasOf ast.TypeRef, FilePath }` plus an unexported `hydrated uint32`
  - **Behavior:** `Model` and `AliasOf` are **mutually exclusive**: a complex shape has `Model` and a nil `AliasOf`, a simple shape the reverse. `Policy` is nil for namespace-scoped shapes.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `ShapeModel` — `{ WithFQN *ast.FQN, Fields map[string]*ShapeModelField }` / `ShapeModelField` — `{ Node, Name, Optional, TypeRef }`
  - **Behavior:** Fields are a map, so **declaration order is lost**. `Optional` here is field optionality, distinct from a nullable type ref.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `createShape(ns *Namespace, p *Policy, stmt *ast.ShapeStatement) -> (*Shape, error)`
  - **Behavior:** Builds the FQN from the policy when policy-scoped, otherwise the namespace; populates `Model` for complex shapes (recording `With` and each field) or `AliasOf` for simple ones. Fields with an empty name are skipped silently.
  - **Side Effects:** Allocation only.
  - **Exceptions:** `duplicate shape field '%s' at %s` — note this one is **not** wrapped in `xerr.ErrIndex`, unlike its sibling errors.

- **Signature:** `(*Shape).resolveDependency(idx *Index, inPolicy *Policy) -> error`
  - **Behavior:** No-ops if already hydrated, if there is no model, or if there is no `with` clause. Otherwise resolves the base shape by **last segment** — first in the enclosing policy, then the enclosing namespace, then by scanning every namespace in the index — verifies it is exported when it comes from elsewhere, and copies its fields in.
  - **Side Effects:** **Mutates `s.Model.Fields`**; sets `hydrated` in a `defer`, so the flag is set even on the error paths.
  - **Exceptions:** Base shape not found; composing with an alias shape; duplicate field between base and derived; non-exported base shape.

- **Signature:** `(*Shape).String()` / `Span()`
  - **Behavior:** FQN rendering and span passthrough for DAG diagnostics.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** **Mutated in place during commit.** A shape read before commit has only its own fields; after commit it carries inherited ones too. Consumers must run after [[index.commit]].
- **Performance/Scale Notes:** The cross-namespace fallback is **O(namespaces) per unresolved base shape**, and each iteration calls `ResolveShape`. Deep composition chains across many namespaces are the worst case.
- **Dependencies Risk:**
  - **`hydrated` is set in a `defer`, including on failure.** A shape whose composition errored is still marked hydrated, so a retry would silently no-op. Since commit aborts on the first error this is currently benign, but it makes partial recovery impossible.
  - **The namespace lookup overrides the policy lookup.** `resolveDependency` checks the policy's shapes first, then unconditionally overwrites `withShape` if the namespace also has that name — so a namespace shape **shadows** a policy shape of the same name, which is the opposite of the usual inner-scope-wins expectation.
  - **Composition resolves by last segment only.** `with a/b/Base` matches on `Base`, so the leading path is effectively advisory during the local lookups and only meaningful in the cross-namespace fallback.
  - **Topological ordering is assumed, not enforced here.** The field copy assumes the base shape is already hydrated. That assumption holds only because [[index.commit]] walks the DAG in order; calling `resolveDependency` directly can copy a half-flattened base.
  - **Aliases cannot be composed with**, and the error points at the *base* shape's span rather than the offending `with` clause.
