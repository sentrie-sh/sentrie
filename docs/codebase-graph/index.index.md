---
id: index.index
type: Class
language: Go
file_path: index/index.go
tags: symbol-table, aggregate-root, concurrency, lifecycle
---

# Node: index.Index (Semantic Model Aggregate Root)

## 1. Architectural Role & Intent
`index/index.go` declares the `Index` struct - the aggregate root holding every namespace, program, and derive in a policy pack - and `AddProgram`, the ingestion routine that turns one parsed file into semantic model entries. It owns the index lifecycle state: a read/write lock, and two `sync.Once` guards that make validation and commit exactly-once operations whose results are cached for the index's lifetime.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.index` | `DEPENDS_ON` | [[ast]] | Type-switches over top-level statements: shape, policy, derive, shape export, derive export. |
| `index.index` | `DEPENDS_ON` | [[dag]] | Holds `dag.G[*Rule]` and `dag.G[*Shape]`, populated by [[index.validate]]. |
| `index.index` | `DEPENDS_ON` | [[pack]] | Stores the manifest via `SetPack`. |
| `index.index` | `CALLS` | [[index.program]] | `createProgram` builds the per-file record. |
| `index.index` | `CALLS` | [[index.namespace]] | `ensureNamespace` creates namespaces and wires parent/child links. |
| `index.index` | `CALLS` | [[index.policy]] | `createPolicy` for each policy statement. |
| `index.index` | `CALLS` | [[index.shape]] | `createShape` for namespace-level shapes. |
| `index.index` | `CALLS` | [[index.derive]] | `ns.addDerive` and `registerDeriveFQN` for namespace derives. |
| [[loader]] | `CALLS` | [[index.index]] | Supplies programs one at a time. |
| [[runtime]] | `READS_FROM` | [[index.index]] | Reads `Namespaces` and `DerivesByFQN` during evaluation. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Index` struct - `{ Pack *pack.PackFile, Namespaces map[string]*Namespace, Programs map[string]*Program, DerivesByFQN map[string]*Derive }` plus unexported `theLock`, `ruleDag`, `shapeDag`, and the validation/commit once-guards
  - **Behavior:** `Namespaces` is keyed by FQN string, `Programs` by source file path, `DerivesByFQN` by fully-qualified derive path - the flat lookup that makes slash-qualified derive calls resolvable without walking the namespace tree.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `CreateIndex() -> *Index`
  - **Behavior:** Allocates all maps and both `sync.Once` guards. The `validated`/`committed` `uint32` fields are declared as atomic-style flags but are written **without** atomic operations elsewhere in the package.
  - **Side Effects:** Allocation only.
  - **Exceptions:** None.

- **Signature:** `(*Index).SetPack(ctx, p *pack.PackFile) -> error`
  - **Behavior:** Stores the manifest under the write lock. Returns `error` for interface symmetry only; it can never fail, and the context is unused.
  - **Side Effects:** Mutates `Pack`.
  - **Exceptions:** None.

- **Signature:** `(*Index).AddProgram(ctx, astProgram *ast.Program) -> error`
  - **Behavior:** Under the write lock: builds a `Program` record, ensures the namespace exists, then iterates statements **from index 1** dispatching on type - comments skipped, shapes and policies created and registered, derives added with a **cloned snapshot** of the namespace's currently-visible derives, and shape/derive exports recorded. Finally registers the program under its file reference.
  - **Side Effects:** Mutates namespaces, policies, shapes, derives, and `Programs`.
  - **Exceptions:** `ctx.Err()`; propagated conflicts from `addShape`/`addPolicy`/`addDerive`; `cannot export unknown derive %q at %s`; `unsupported top-level statement %T at %s`.

- **Signature:** `(*Index).ensureNamespace(_, namespace *ast.NamespaceStatement) -> (*Namespace, error)`
  - **Behavior:** Returns the existing namespace or creates one, then scans **every** already-known namespace to wire parent/child links in both directions.
  - **Side Effects:** Mutates the namespace tree.
  - **Exceptions:** Conflicts raised by `addChild` when a child's base name collides with a policy, shape, or derive in the parent.

## 4. Operational Context & Gotchas
- **Statefulness:** Mutable aggregate with a strict lifecycle: populate → validate → commit → read. Not reusable after a failed validation.
- **Performance/Scale Notes:** `ensureNamespace` is **O(n) over all namespaces per new namespace**, so building the tree is quadratic in namespace count. Fine for tens of namespaces; noticeable for thousands. `cloneDeriveMap` copies the visible-derive map on every derive registration.
- **Dependencies Risk:**
  - **The `i := 1` loop start is a latent bug.** `AddProgram` assumes the namespace occupies statement zero, but [[parser.parse]] legally emits `CommentStatement` entries *before* it. For a file that opens with a comment, the statement at index 1 (the namespace itself, or the first real declaration) is skipped without any error.
  - **Locking does not cover reads.** `theLock` is taken by `SetPack` and `AddProgram` only. `Validate`, `Commit`, and all resolution methods read unguarded, so an index must be fully populated before it is shared.
  - **The `validated`/`committed` flags are not used atomically.** They are plain `uint32` fields written inside the `Once` bodies; the real exactly-once guarantee comes from `sync.Once`, not from these. Do not read them as a concurrency signal.
  - **Derive visibility snapshots are taken here.** `cloneDeriveMap(ns.Derives)` captures what is visible *at that moment*, which is the mechanism behind the file-ordering constraint documented in [[index.derive]].
