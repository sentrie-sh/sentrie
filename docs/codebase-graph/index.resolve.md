---
id: index.resolve
type: Function / Endpoint
language: Go
file_path: index/resolve.go
tags: lookup, resolution, visibility, public-api
---

# Node: index.Resolve (Lookup and Export Verification API)

## 1. Architectural Role & Intent
The read-side API of the index: exact-match lookups for namespaces, policies, shapes, and derives; export verification for rules, shapes, and derives; and the canonical FQN string builders that keep every subsystem agreeing on what a fully-qualified name looks like. This is the surface [[runtime]] and [[cmd]] use to turn a name into a model object.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.resolve` | `DEPENDS_ON` | [[xerr]] | Returns `ErrNamespaceNotFound`, `ErrPolicyNotFound`, `ErrShapeNotFound`, `ErrNotExported`, `ErrIndex`. |
| `index.resolve` | `DEPENDS_ON` | [[ast]] | Uses `ast.FQNSeparator` to join FQN segments. |
| `index.resolve` | `READS_FROM` | [[index.index]] | Reads `Namespaces` and `DerivesByFQN` directly. |
| `index.resolve` | `READS_FROM` | [[index.namespace]] | Reads `Policies`, `Shapes`, `ShapeExports`, `DeriveExports`. |
| [[index.segments]] | `CALLS` | [[index.resolve]] | `ResolveSegments` probes `ResolveNamespace` and `ResolvePolicy` repeatedly. |
| [[index.shape]] | `CALLS` | [[index.resolve]] | Composition falls back to `ResolveShape` across namespaces and calls `VerifyShapeExported`. |
| [[index.validate]] | `CALLS` | [[index.resolve]] | Rule-cycle detection resolves imported policies; shape-cycle detection resolves base shapes. |
| [[runtime]] | `CALLS` | [[index.resolve]] | Entry point for evaluating a named policy or rule. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).ResolveNamespace(ns string) -> (*Namespace, error)`
  - **Behavior:** Exact map lookup by FQN string. No parent traversal, no prefix matching.
  - **Side Effects:** None.
  - **Exceptions:** `xerr.ErrNamespaceNotFound(ns)`.

- **Signature:** `(*Index).ResolvePolicy(ns, policy string) -> (*Policy, error)`
  - **Behavior:** Resolves the namespace exactly, then the policy by simple name within it. Explicitly **does not traverse parents**, as its doc comment states.
  - **Side Effects:** None.
  - **Exceptions:** Namespace not found; `xerr.ErrPolicyNotFound` with a `filepath.Join`-built path.

- **Signature:** `(*Index).ResolveShape(ns, shape string) -> (*Shape, error)`
  - **Behavior:** Same shape as `ResolvePolicy`, over `Namespace.Shapes`. Only namespace-level shapes are reachable; policy-local shapes are not.
  - **Side Effects:** None.
  - **Exceptions:** Namespace not found; `xerr.ErrShapeNotFound`.

- **Signature:** `(*Index).ResolveDerive(fqn string) -> (*Derive, error)`
  - **Behavior:** Flat lookup in `DerivesByFQN`, covering both namespace- and policy-scoped derives. **Does not check visibility** - that is the caller's job via `Derive.VisibleFromPolicy`.
  - **Side Effects:** None.
  - **Exceptions:** `derive %q not found`, wrapped in `xerr.ErrIndex`.

- **Signature:** `(Policy).VerifyRuleExported(rule string) -> error` / `(Namespace).VerifyShapeExported(shape string) -> error` / `(Namespace).VerifyDeriveExported(name string) -> error`
  - **Behavior:** Membership checks against `RuleExports`, `ShapeExports`, and `DeriveExports`. Note all three take **value receivers**, copying the struct on every call.
  - **Side Effects:** None.
  - **Exceptions:** `xerr.ErrNotExported` carrying the canonical FQN.

- **Signature:** `RuleFQN(ns, policy, rule) -> string` / `ShapeFQN(ns, shape) -> string` / `DeriveFQN(ns, derive) -> string`
  - **Behavior:** Package-level joiners using `ast.FQNSeparator`. These define the canonical key format for `DerivesByFQN` and for every not-exported diagnostic.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless reads over the index maps.
- **Performance/Scale Notes:** All O(1) map probes. The value receivers on the three `Verify*` methods copy a `Policy` or `Namespace` struct - many maps and slices' worth of header copying - on every call; harmless for correctness since only maps are read, but wasteful on hot paths.
- **Dependencies Risk:**
  - **No lock is taken.** These read the index maps directly while `AddProgram` writes under `theLock`. Resolving concurrently with population is a data race; finish building before sharing.
  - **Resolution is exact-match only.** There is no parent-namespace fallback anywhere, so a policy referenced by an unqualified or partially-qualified name will not be found. [[index.segments]] exists precisely to bridge that gap.
  - **`filepath.Join` builds not-found messages.** On Windows the reported path uses backslashes while real FQNs use the `ast.FQNSeparator`, so error text does not round-trip as a valid name.
  - **`ResolveShape` misses policy-local shapes**, which live on `Policy.Shapes` and have no resolver of their own.
  - **`ResolveDerive` bypasses visibility.** Using it without a `VisibleFromPolicy` check would let a caller reach another policy's private derive.
