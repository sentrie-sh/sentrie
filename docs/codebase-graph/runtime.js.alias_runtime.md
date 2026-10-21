---
id: runtime.js.alias_runtime
type: Class
language: Go
file_path: runtime/js/alias_runtime.go
tags: commonjs, vm-host, circular-dependencies, cancellation, caching
---

# Node: js.AliasRuntime (CommonJS VM Host)

## 1. Architectural Role & Intent
Hosts exactly one `goja.Runtime` per `use … as alias` binding and implements CommonJS `require()` on top of it, including a per-VM exports cache, circular-dependency support via placeholder registration, and context-cancellation plumbing through goja's interrupt mechanism. It is the boundary object: everything a policy's JavaScript can reach is reachable because this type put it in the VM.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.alias_runtime` | `IMPORTS` | `ext.dop251.goja` | Owns the `goja.Runtime`; uses `Interrupt`, `RunProgram`, `AssertFunction`. |
| `runtime.js.alias_runtime` | `CALLS` | [[runtime.js.registry]] | `LoadRequire` and `programFor` to resolve and compile dependencies. |
| `runtime.js.alias_runtime` | `CALLS` | [[runtime.js.stdlib]] | `SetupStdLib` installs globals before any module runs. |
| `runtime.js.alias_runtime` | `MUTATES` | `ext.dop251.goja` | Passes a per-module `require` closure straight to each module factory; no global is set. |
| [[runtime.modules]] | `DEPENDS_ON` | `runtime.js.alias_runtime` | Creates one instance per module call and invokes exported functions through it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewAliasRuntime(reg *Registry, baseDir string) -> *AliasRuntime`
  - **Behavior:** Creates a fresh VM with an empty exports cache. No standard library is installed yet.
  - **Side Effects:** Allocates a `goja.Runtime`.
  - **Exceptions:** None.

- **Signature:** `Require(ctx context.Context, fromDir, spec string) -> (*goja.Object, error)`
  - **Behavior:** The full CommonJS load sequence:
    1. Resolve through the registry.
    2. **Go-backed module** - call the provider, cache the result, return. No program execution.
    3. **Cache hit** - return the cached exports.
    4. Create `module` and `exports` objects and **place `exports` in the cache before executing**, so a circular `require` resolves to the partially-populated object instead of recursing forever.
    5. Compile, install the interrupt, bind a child `require` closed over this module's directory, run the program to obtain the factory, invoke it with `(require, module, exports)`.
    6. Re-read `module.exports` - the factory may have reassigned it - and update the cache with the final object.
  - **Side Effects:** Executes arbitrary module code; mutates the VM and the cache; registers a context `AfterFunc` for the interrupt.
  - **Exceptions:** Resolution and compilation errors; `module did not evaluate to a function`; any error thrown by the module body. Every failure path removes the placeholder from the cache.

- **Signature:** `InterruptOnDone(ctx context.Context, vm *goja.Runtime) -> (stop func())`
  - **Behavior:** Registers a `context.AfterFunc` that calls `vm.Interrupt(ctx.Err())` when `ctx` ends. `stop` deregisters it and, if it already started, waits for it to finish, so nothing touches the VM after `stop` returns. An interrupt that fired is **left set**: the context is over, and a nested `require` must not clear the interrupt its parent frame relies on. A nil context yields a no-op. `Require` uses it for each module load; [[runtime.modules]] uses it around each call and clears a fired interrupt on the long-lived probe VM after its load.
  - **Side Effects:** VM interrupt state.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** One VM plus one module cache per instance. Since #105 an instance lives for one module call (or, for a module's probe, for the executor's lifetime without ever serving a call), so module-level state never outlives the call that created it.
- **Performance/Scale Notes:** The exports cache means a module body executes once per VM. Because each `use` alias gets its own VM, a module required by two aliases is executed twice, in two separate object graphs. goja is single-threaded, so a VM must not be shared concurrently - [[runtime.modules]] gives each call its own VM.
- **Dependencies Risk:**
  - **Circular dependencies see a stale exports object if the factory reassigns `module.exports`.** The placeholder cached before execution is the original `exports`; the final cache update happens after. A module that does `module.exports = {…}` and participates in a cycle hands its circular consumer the empty placeholder. This is the classic CommonJS hazard, faithfully reproduced.
  - **Go-backed providers bypass the cache check.** The provider branch is evaluated **before** the cache lookup, so every `require` of a Go builtin re-fabricates a fresh exports object and overwrites the cache entry. Two references to the same builtin within one VM therefore get **different objects**, which matters for identity comparison and for any provider that tries to hold state.
  - **A failed load removes the cache entry**, so a subsequent `require` retries - which is the opposite of the registry's permanent compile-error caching, so retry semantics differ by which layer failed.
