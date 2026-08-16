---
id: runtime.js.stdlib
type: Module / File
language: Go
file_path: runtime/js/stdlib.go
tags: sandbox, permissions, environment, globals, security
---

# Node: js.stdlib (VM Global Installation)

## 1. Architectural Role & Intent
Installs the globals available to every module before any JavaScript runs. Currently that is exactly one thing — a filtered `env` object — which makes this node the entire deliberate surface through which host state enters the sandbox. Everything else JavaScript can reach arrives through `require`, not through globals.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.stdlib` | `READS_FROM` | `ext.os_environment` | `os.Environ()` supplies the candidate variables. |
| `runtime.js.stdlib` | `CALLS` | [[pack]] | `PackFile.Permissions.CheckEnvAccess(key)` decides each variable individually. |
| `runtime.js.stdlib` | `MUTATES` | [[runtime.js.alias_runtime]] | Sets the `env` global on the VM. |
| [[runtime.modules]] | `CALLS` | `runtime.js.stdlib` | Invoked once per VM, before `Require`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*AliasRuntime).SetupStdLib(ctx context.Context, pack *pack.PackFile) -> error`
  - **Behavior:** The single entry point; currently delegates only to the env installer. Documented as **required** before `Require` is used or the VM is invoked.
  - **Side Effects:** Sets VM globals.
  - **Exceptions:** Propagated.

- **Signature:** `(*AliasRuntime).setupEnvStdLib(_ context.Context, pack *pack.PackFile) -> error`
  - **Behavior:** Walks `os.Environ()`, splits each entry on the first `=`, skips malformed entries, and copies a variable into the `env` object **only** when `pack.Permissions.CheckEnvAccess(key)` returns true. Sets `env` on the VM regardless of how many variables survived.
  - **Side Effects:** Creates and installs the `env` object.
  - **Exceptions:** Propagated from `Set`.

## 4. Operational Context & Gotchas
- **Statefulness:** Runs once per VM. The `env` object is a **snapshot** taken at setup; later changes to the process environment are not visible to already-initialised VMs.
- **Performance/Scale Notes:** O(environment size) per VM, with one permission check per variable. Negligible in absolute terms, but it is paid on every VM creation and therefore on every pool miss.
- **Dependencies Risk:**
  - **A nil pack or nil permissions block exposes nothing.** The guard is `pack != nil && pack.Permissions != nil && CheckEnvAccess(key)`, so the default is an empty `env` object rather than the full environment. That is the correct fail-closed direction, and it is worth preserving deliberately — inverting this condition would leak the entire host environment, including credentials, to pack-supplied JavaScript.
  - **`env` is always installed, even when empty.** Module code cannot distinguish "no permissions configured" from "permitted but unset", since both yield `undefined` on lookup.
  - **This is the only host-state global.** The sandbox's strength rests on that fact, and nothing structurally prevents a future addition here from widening it — there is no allowlist or review gate on what `SetupStdLib` may install.
  - **Permission checks happen at VM setup, not at access time**, so a permission model that ever becomes dynamic would not be reflected in an already-built VM.
  - **The environment snapshot lives as long as the pooled VM does**, so a secret rotated in the process environment remains visible to JavaScript until that VM is evicted.
