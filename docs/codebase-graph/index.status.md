---
id: index.status
type: Function / Endpoint
language: Go
file_path: index/status.go
tags: health, observability, read-only, statistics
---

# Node: index.Status (Read-only Health Report)

## 1. Architectural Role & Intent
A side-effect-free view of an index for health checks. `Validate` and `IsValid` run the pipeline on first call, so they cannot answer "has this index been validated?" without changing the answer. `Status` reports the lifecycle state, the recorded errors and the definition counts instead.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.status` | `READS_FROM` | [[index.index]] | Reads `validated`, `validationError`, `committed`, `commitError`, `Namespaces` and `Programs`, and each namespace's policies, rules and shapes. |
| [[api.health]] | `CALLS` | `index.status` | The deep health check. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).Status() -> Status`
  - **Behavior:** Returns `Status{Validated, ValidationError, Committed, CommitError, Namespaces, Policies, Programs, Rules, Shapes}`. `Validated` and `Committed` mean "that phase has finished", whatever its outcome, so a failed validation reports `Validated: true` with `ValidationError` set and `Committed: false`. When a commit fails inside `Validate`, the same error is in both `ValidationError` and `CommitError`. `Shapes` counts namespace-level and policy-local shapes. State is read under `stateMu`; counts under `theLock.RLock`.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless snapshot; two calls can differ if the index is still being populated or validated.
- **Dependencies Risk:**
  - **Safe against a running `Validate` or `Commit`**, because those now write the flags and errors under `stateMu`. The counts walk maps that `AddProgram` mutates under `theLock`, so they are consistent with population but not with code that mutates the maps directly.
