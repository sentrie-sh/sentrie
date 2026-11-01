---
id: index.shape_refs
type: Function / Endpoint
language: Go
file_path: index/shape_refs.go
tags: static-analysis, name-resolution, shapes, diagnostics
---

# Node: index.checkShapeRefs (Static Shape Name Check)

## 1. Architectural Role & Intent
Resolves, at index time, every shape named in a type outside `x is T` (#282): facts, lets, shape aliases and fields, derive and lambda signatures, block lets and casts, and the types nested in them. Before this check a missing shape there passed `sentrie validate` and failed only when the expression was evaluated, and a typed `let` that was never evaluated never reported it at all. Names resolve with [[index.is_resolve]]'s `LookupShape`, which mirrors the runtime's lookup ([[runtime.typeref_shape]]).

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.shape_refs` | `CALLS` | [[index.typeref_walk]] | `forEachTypeRef` supplies every type ref with its scope. |
| `index.shape_refs` | `CALLS` | [[index.is_resolve]] | `LookupShape` resolves each name; `joinSiteErrors` orders the report. |
| `index.shape_refs` | `READS_FROM` | [[index.namespace]] | Namespace shapes and, for a qualified name, the target namespace's exports. |
| `index.shape_refs` | `DEPENDS_ON` | [[xerr]] | Errors wrap `xerr.ErrIndex`; matches `xerr.NotExportedError` from the lookup to pick the message. |
| [[index.validate]] | `CALLS` | [[index.shape_refs]] | Runs right after `resolveIsExpressions`, before any check reads a type. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkShapeRefs(ctx) -> error`
  - **Behavior:** For every `ShapeTypeRef` the walk yields outside an `is` type, looks the name up in the site's scope: the policy's shapes, then the namespace's, then for three or more segments the shape the last segment names in the namespace the rest names, which must export it. Collects every failure, sorted into source order (file, line, column).
  - **Side Effects:** None.
  - **Exceptions:** Joined, one per type ref, at the type ref's span: `shape '<name>' not found at <span>: index error` (no such shape, or no such namespace), or `shape '<name>' not exported at <span>: index error`. Each matches `xerr.ErrIndex`. The text is the one the runtime and shape composition ([[index.shape]]) use for a missing shape. `validation cancelled` on context cancellation.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One walk plus a map lookup or two per shape type ref.
- **Dependencies Risk:**
  - **`is` types are skipped.** [[index.is_resolve]] has already resolved and reported their shape names in its own wording; reporting them here too would double every error.
  - **Scope is the declaring site's.** Inside a derive the scope is the derive's policy, or its namespace for a namespace-level derive; a shape's field and alias types resolve in the shape's scope. Until #281 (PR #291) lands, the runtime resolves a shape field's type in the evaluating policy's scope instead (see [[index.typeref_walk]]), and until #271 (PR #276) lands it resolves a derive's types in the caller's; so a namespace shape whose field, or a namespace derive whose signature, names a policy-only shape is rejected here even where one policy could have evaluated it. Once both land the runtime agrees with this check, and this note should be deleted.
  - **A two-segment qualified name never resolves**, matching the runtime.
