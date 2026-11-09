---
id: runtime.pack_data
type: Function / Endpoint
language: Go
file_path: runtime/pack_data.go
tags: pack-data, static-data, validation, immutability
---

# Node: runtime.packDataValue / validatePackData (Pack Data at Runtime)

## 1. Architectural Role & Intent
Serves the values of `data <name>` declarations and validates typed declarations once. The loader decodes and boxes each `[data]` entry at pack load; this file is where an identifier read finds that shared value, and where `data roles: RoleTable` is checked against the value when the executor is built, so a pack whose data does not match its declared types fails before any request.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.pack_data` | `READS_FROM` | [[pack]] | `PackFile.DataValues`, through `Executor.Index().Pack`. |
| `runtime.pack_data` | `READS_FROM` | [[index.policy]] | Iterates `Policy.Data`, sorted. |
| `runtime.pack_data` | `CALLS` | [[runtime.typeref]] | `validateValueAgainstTypeRef` against each declared type. |
| `runtime.pack_data` | `CALLS` | [[runtime.exec_ctx]] | A fresh `ExecutionContext` per validated declaration. |
| [[runtime.executor]] | `CALLS` | [[runtime.pack_data]] | `NewExecutor` calls `validatePackData` and fails construction on a mismatch. |
| [[runtime.eval_ident]] | `CALLS` | [[runtime.pack_data]] | `packDataValue` resolves a data alias outside a derive. |

## 3. Interface Contracts & Public Surface

- **Signature:** `packDataValue(exec: Executor, d: *ast.DataStatement) -> (box.Value, error)`
  - **Behavior:** Returns the loaded value for `d.Name`, the same `box.Value` for every evaluation.
  - **Side Effects:** None.
  - **Exceptions:** `pack data "<name>" is not loaded at <span>` when the index has no pack or no such loaded entry.

- **Signature:** `(*executorImpl).validatePackData(ctx: context.Context) -> error`
  - **Behavior:** For every typed declaration, sorted, validates the loaded value against the type. Constraint arguments are literals, so nothing from the policy is bound.
  - **Side Effects:** Evaluates literal constraint arguments.
  - **Exceptions:** `pack data "<name>" in policy <fqn> does not match its declared type: ...`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; the values live in the [[pack]] manifest for the executor's lifetime.
- **Performance/Scale Notes:** Validation is once per typed declaration per executor, not per evaluation. Reads are a map lookup. A large value recorded as an identifier's trace result is as large in a trace as an equally large fact.
- **Dependencies Risk:**
  - **Shared values must never be mutated.** Every concurrent evaluation reads the same boxed value. No policy operation mutates a value, and values cross into TypeScript only as copies (`box.TryToBoundaryAny`, then `normalizeBoundaryForJS` in [[runtime.modules]]), so a module cannot change them; a future builtin that mutates a list or dict in place would break this.
  - **`NewExecutor` uses `context.Background()`** for this check because it takes no context; the work is CPU-only over data already in memory and bounded by the loader's file size limit.
