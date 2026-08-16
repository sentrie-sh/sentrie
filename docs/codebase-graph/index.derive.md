---
id: index.derive
type: Class
language: Go
file_path: index/derive.go
tags: pure-functions, visibility, scoping, snapshot-semantics
---

# Node: index.Derive (Pure Function Model and Visibility)

## 1. Architectural Role & Intent
`Derive` models a `derive name = (…) => { … }` declaration - Sentrie's pure, reusable function abstraction - at either namespace or policy scope. Its distinguishing feature is the **bind-time visibility snapshot**: each derive records, in `DefineShort` and `DefineFQN`, exactly which other derives were visible when it was registered, so name resolution inside a derive body is deterministic and independent of what gets loaded afterwards.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.derive` | `DEPENDS_ON` | [[ast]] | Wraps `ast.LambdaExpression`, `ast.DeriveStatement`, `ast.ExportDeriveStatement`; builds FQNs with `ast.CreateFQN`. |
| `index.derive` | `DEPENDS_ON` | [[xerr]] | Duplicate declarations become `ErrConflict` with both spans. |
| `index.derive` | `CALLS` | [[index.namespace]] | `addDerive` goes through `checkNameAvailable` for cross-kind uniqueness. |
| [[index.index]] | `CALLS` | [[index.derive]] | `AddProgram` registers namespace derives with a cloned visibility snapshot. |
| [[index.policy_stmt]] | `CALLS` | [[index.derive]] | Policy derives are registered with an **overlay** of namespace snapshot plus policy-so-far. |
| [[index.derive_cycle]] | `READS_FROM` | [[index.derive]] | Resolves callees through `DefineShort` and `DefineFQN`. |
| [[index.derive_purity]] | `READS_FROM` | [[index.derive]] | Uses the same snapshots to decide which identifiers and calls are legal. |
| [[runtime]] | `READS_FROM` | [[index.derive]] | Looks derives up by FQN and enforces visibility at call time. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Derive` struct - `{ Name, FQN, Lambda *ast.LambdaExpression, Namespace, Policy *Policy, Statement, DefineShort map[string]*Derive, DefineFQN map[string]*Derive }`
  - **Behavior:** `Policy` is nil for namespace-scoped derives. `DefineShort` is keyed by simple name, `DefineFQN` by fully-qualified path - the two lookup routes a derive body may use.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `(*Derive).VisibleFromPolicy(caller *Policy) -> error`
  - **Behavior:** Namespace-scoped derives (nil `Policy`) are always visible. A policy-scoped derive is visible only to its own policy, compared by FQN string.
  - **Side Effects:** None.
  - **Exceptions:** `derive %q is policy-scoped and not visible outside its policy`; `derive %q is not visible from policy %q`.

- **Signature:** `(*Derive).VisibleFromDeriveCaller(caller *Derive) -> error`
  - **Behavior:** The derive-to-derive form; delegates to `VisibleFromPolicy(caller.Policy)`. A namespace-scoped derive therefore **cannot** call a policy-scoped one.
  - **Side Effects:** None.
  - **Exceptions:** Same as above.

- **Signature:** `(*Derive).String()` / `Span()`
  - **Behavior:** FQN rendering; span falls back to the lambda when `Statement` is nil.
  - **Side Effects:** None.
  - **Exceptions:** Panics only if both are nil.

- **Signature:** `(*Namespace).addDerive(idx, stmt, visibleBefore) -> (*Derive, error)` / `(*Policy).addDerive(idx, stmt, nsSnapshot, policySoFar) -> (*Derive, error)`
  - **Behavior:** Both check for conflicts, build the derive with a **cloned** snapshot, and register it globally by FQN. The policy variant overlays the policy's own derives on top of the namespace snapshot, so a policy derive shadows a namespace derive of the same name.
  - **Side Effects:** Mutate the scope's `Derives` map and `idx.DerivesByFQN`; the policy variant also records the name in `seenIdentifiers`.
  - **Exceptions:** `ErrConflict("derive declaration", …)` from the local map, the shared name gate, or the global FQN registry.

- **Signature:** `ExportedDerive` - `{ Name, Statement }`
  - **Behavior:** Records a namespace-level derive export. Unlike shape exports, the exported name **is** verified against declarations, in [[index.index]].
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** Immutable after registration. The snapshots are copies, so later mutations to the source maps do not leak in.
- **Performance/Scale Notes:** Every registration performs a **full map clone plus an FQN index rebuild**, making derive registration O(visible derives) each - quadratic across a namespace with many derives. Memory grows with the square of the derive count in a large namespace.
- **Dependencies Risk:**
  - **File load order is semantically significant.** This is the package's sharpest edge, and it is called out in the source: a derive registered before its helper cannot see that helper, permanently. Either keep helpers in the same file or add programs in dependency order. There is no diagnostic for getting this wrong - the call simply fails purity checking with "unknown derive".
  - **Namespace derives cannot call policy derives.** `VisibleFromDeriveCaller` returns an error when the caller has no policy, so hoisting a helper to namespace scope can break its call sites in the opposite direction from what one expects.
  - **Visibility is compared by FQN string**, not pointer identity, so two `*Policy` values with the same FQN are treated as the same policy.
  - **The global FQN registry is flat.** A namespace derive and a policy derive can never collide (their FQNs differ by the policy segment), but two policies with the same FQN would.
