---
id: index.policy_stmt
type: Module / File
language: Go
file_path: index/policy_stmt.go
tags: classification, phase-ordering, helpers, tags
---

# Node: index.PolicyStmtKind (Statement Classification and Tag Grouping)

## 1. Architectural Role & Intent
A small support file for [[index.policy]] holding two things: `policyStmtKind`, a coarse classification of policy-body statements into the four ordering phases (metadata, fact, use, body) plus comment and unknown; and `buildTagsByKey`, which folds the ordered `TagPairs` slice into a queryable multi-map. It exists to give the phase model a single named vocabulary even though `createPolicy` still branches on concrete types.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.policy_stmt` | `DEPENDS_ON` | [[ast]] | Type-switches over all policy-body statement types to assign a kind. |
| [[index.policy]] | `CALLS` | [[index.policy_stmt]] | `createPolicy` uses `policyStmtKindOf` to skip comments and `buildTagsByKey` to build `TagsByKey`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `policyStmtKind` constants — `policyStmtComment`, `policyStmtMetadata`, `policyStmtFact`, `policyStmtUse`, `policyStmtBody`, `policyStmtUnknown`
  - **Behavior:** Names the phases that `createPolicy`'s `policyPhase` state machine advances through. `policyStmtBody` covers lets, rules, rule exports, shapes, and derives collectively.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `policyStmtKindOf(stmt ast.Statement) -> policyStmtKind`
  - **Behavior:** Total classification; anything unrecognised returns `policyStmtUnknown` rather than failing.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `isMetadataStmt` / `isFactStmt` / `isUseStmt` / `isBodyStmt`
  - **Behavior:** Thin predicates over `policyStmtKindOf`. Per the file's own comment, these exist **mainly for tests**, not for production dispatch.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `buildTagsByKey(pairs []PolicyTagPair) -> map[string][]string`
  - **Behavior:** Groups tag values by key, preserving append order within each key. Returns **nil** — not an empty map — when there are no pairs.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless pure functions.
- **Performance/Scale Notes:** Negligible; one pass over the tag list.
- **Dependencies Risk:**
  - **The classifier is advisory, not authoritative.** `createPolicy` re-derives phase transitions from a concrete type switch and only uses `policyStmtKindOf` to skip comments. Adding a statement type here without also adding a case in `createPolicy` yields `unsupported statement in policy`, and the two can drift silently — the file's leading comment flags exactly this.
  - **`policyStmtUnknown` is never acted on.** Unrecognised statements fall through to `createPolicy`'s `default` branch, so the kind constant is effectively unused.
  - **`buildTagsByKey` returns nil for the empty case.** Reading `TagsByKey` is safe (nil-map reads work), but writing to it or asserting non-nil is not.
  - **Tag duplication is preserved, not resolved.** A key repeated with different values yields a multi-value slice; nothing in the index picks a winner, so consumers must define that policy themselves.
