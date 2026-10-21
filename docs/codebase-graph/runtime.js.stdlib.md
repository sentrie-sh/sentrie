---
id: runtime.js.stdlib
type: Module / File
language: Go
file_path: runtime/js/stdlib.go
tags: sandbox, permissions, environment, globals, security
---

# Node: js.stdlib (VM Global Installation)

## 1. Architectural Role & Intent
Installs the globals available to every module before any JavaScript runs. Currently that is exactly one thing - a filtered `env` object - which makes this node the entire deliberate surface through which host state enters the sandbox. Everything else JavaScript can reach arrives through `require`, not through globals.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.stdlib` | `READS_FROM` | [[infra.os_environment]] | `os.Environ()` supplies the candidate variables. |
| `runtime.js.stdlib` | `CALLS` | [[pack]] | `PackFile.Permissions.CheckEnvAccess(key)` decides each variable individually. |
| `runtime.js.stdlib` | `MUTATES` | [[runtime.js.alias_runtime]] | Sets the `env` global on the VM. |
| `runtime.js.stdlib` | `READS_FROM` | [[capabilities]] | The registry's grant decides whether the deterministic `Date`/`Math.random` shims are installed. |
| [[runtime.modules]] | `CALLS` | `runtime.js.stdlib` | Invoked once per VM, before `Require`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*AliasRuntime).SetupStdLib(ctx context.Context, pack *pack.PackFile) -> error`
  - **Behavior:** The single entry point. Installs `env`, then, when the registry's [[capabilities]] grant sets `deterministic_js`, the deterministic shims (`deterministic.go`). Documented as **required** before `Require` is used or the VM is invoked.
  - **Side Effects:** Sets VM globals.
  - **Exceptions:** Propagated.

- **Signature:** `installDeterministicShims(vm) -> error` / `BeginCall(vm, start time.Time, moduleKey, fn string) -> error` (in `deterministic.go`)
  - **Behavior:** `installDeterministicShims` wraps the global `Date` so `Date.now()`, `new Date()` and `Date()` read a pinned clock (all other constructor forms, `Date.parse`, `Date.UTC` and the prototype are the real ones), and replaces `Math.random` with a PCG stream. The wrapper captures the real `Date` and `Reflect.construct` at install time and looks nothing up on shared prototypes afterwards, so pack code that swaps `bind`/`apply`/`concat` cannot recover the real `Date`; it forwards `new.target`, so subclasses work. The state lives in a Go-side map weakly keyed by `*goja.Runtime` (with a cleanup on collection), so no global is added. `BeginCall`, run by [[runtime.modules]] before every call, always sets `ExecutionStartTimeUnixKey`, and when the shims are installed it pins the clock to the execution start (milliseconds) and reseeds the stream from the execution start and `moduleKey`+`fn`.
  - **Side Effects:** Replaces `Date` and `Math.random` on the VM; mutates shim state per call.
  - **Exceptions:** Propagated from goja.

- **Signature:** `(*AliasRuntime).setupEnvStdLib(_ context.Context, pack *pack.PackFile) -> error`
  - **Behavior:** Walks `os.Environ()`, splits each entry on the first `=`, skips malformed entries, and copies a variable into the `env` object **only** when `pack.Permissions.CheckEnvAccess(key)` returns true. Sets `env` on the VM regardless of how many variables survived.
  - **Side Effects:** Creates and installs the `env` object.
  - **Exceptions:** Propagated from `Set`.

## 4. Operational Context & Gotchas
- **Statefulness:** Runs once per VM. The `env` object is a **snapshot** taken at setup; later changes to the process environment are not visible to already-initialised VMs.
- **Performance/Scale Notes:** O(environment size) per VM, with one permission check per variable. Negligible in absolute terms, but it is paid on every VM creation and therefore on every pool miss.
- **Dependencies Risk:**
  - **A nil pack or nil permissions block exposes nothing.** The guard is `pack != nil && pack.Permissions != nil && CheckEnvAccess(key)`, so the default is an empty `env` object rather than the full environment. That is the correct fail-closed direction, and it is worth preserving deliberately - inverting this condition would leak the entire host environment, including credentials, to pack-supplied JavaScript.
  - **`env` is always installed, even when empty.** Module code cannot distinguish "no permissions configured" from "permitted but unset", since both yield `undefined` on lookup.
  - **Deterministic shims reseed per call, not per draw.** Two calls to the same module function in one execution see the same `Math.random` sequence, and before the first call (module top-level code at load) the clock reads 0. Both are deterministic by design. Only `Date` and `Math.random` are covered: any future host clock or random source must be shimmed here too.
  - **`env` is the only host-state global.** The sandbox's strength rests on that fact, and nothing structurally prevents a future addition here from widening it - there is no allowlist or review gate on what `SetupStdLib` may install.
  - **Permission checks happen at VM setup, not at access time**, so a permission model that ever becomes dynamic would not be reflected in an already-built VM.
  - **The environment snapshot lives as long as the pooled VM does**, so a secret rotated in the process environment remains visible to JavaScript until that VM is evicted.
