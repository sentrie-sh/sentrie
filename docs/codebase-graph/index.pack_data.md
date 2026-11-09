---
id: index.pack_data
type: Function / Endpoint
language: Go
file_path: index/pack_data.go
tags: validation, pack-data, static-data, manifest
---

# Node: index.checkPackData (Pack Data Declaration Check)

## 1. Architectural Role & Intent
The static check that ties each policy's `data <name>` declarations to the pack manifest: every declared name must be an entry of the `[data]` table and must have been loaded by [[loader]]. It runs inside [[index.validate]], so `sentrie validate`, `exec` and `serve` reject a policy that reads data the pack does not ship, before any evaluation.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.pack_data` | `READS_FROM` | [[index.policy]] | Iterates `Policy.Data` for every policy, sorted by namespace, policy and alias. |
| `index.pack_data` | `READS_FROM` | [[pack]] | Reads `PackFile.Data` (manifest entries) and `PackFile.DataValues` (loaded values). |
| [[index.validate]] | `CALLS` | [[index.pack_data]] | Runs after the import `with` check. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkPackData(ctx: context.Context) -> error`
  - **Behavior:** For each declaration: fails when the index has no pack or the pack has no `[data]` table, when the name is not a manifest entry, or when the entry has no loaded value.
  - **Side Effects:** None.
  - **Exceptions:** `data "<name>" at <span>: the pack declares no [data] table`; `... no such entry in the pack's [data] table`; `... the entry was not loaded with the pack`; all wrap `xerr.ErrIndex`. Cancellation yields `validation cancelled`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One map probe per declaration.
- **Dependencies Risk:**
  - **An index without a pack cannot satisfy a declaration.** In-memory indexes built from source alone (such as `lang_test` fixtures) fail on any `data` declaration; tests that exercise data must call `SetPack` with a loaded manifest.
  - **The declared type is not checked here.** Value-against-type validation needs the runtime type checker and happens once per executor in [[runtime.pack_data]].
