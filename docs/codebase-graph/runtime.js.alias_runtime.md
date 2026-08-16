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
| `runtime.js.alias_runtime` | `DEPENDS_ON` | `ext.goja` | Owns the `goja.Runtime`; uses `Interrupt`, `ClearInterrupt`, `RunProgram`, `AssertFunction`. |
| `runtime.js.alias_runtime` | `CALLS` | [[runtime.js.registry]] | `LoadRequire` and `programFor` to resolve and compile dependencies. |
| `runtime.js.alias_runtime` | `CALLS` | [[runtime.js.stdlib]] | `SetupStdLib` installs globals before any module runs. |
| `runtime.js.alias_runtime` | `MUTATES` | `ext.goja` | Sets the `__require` global around each factory invocation. |
| [[runtime.modules]] | `DEPENDS_ON` | `runtime.js.alias_runtime` | Pools these instances and invokes exported functions through them. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewAliasRuntime(reg *Registry, baseDir string) -> *AliasRuntime`
  - **Behavior:** Creates a fresh VM with an empty exports cache. No standard library is installed yet.
  - **Side Effects:** Allocates a `goja.Runtime`.
  - **Exceptions:** None.

- **Signature:** `Require(ctx context.Context, fromDir, spec string) -> (*goja.Object, error)`
  - **Behavior:** The full CommonJS load sequence:
    1. Resolve through the registry.
    2. **Go-backed module** — call the provider, cache the result, return. No program execution.
    3. **Cache hit** — return the cached exports.
    4. Create `module` and `exports` objects and **place `exports` in the cache before executing**, so a circular `require` resolves to the partially-populated object instead of recursing forever.
    5. Compile, install the interrupt, bind a child `require` closed over this module's directory, run the program to obtain the factory, invoke it with `(require, module, exports)`.
    6. Re-read `module.exports` — the factory may have reassigned it — and update the cache with the final object.
  - **Side Effects:** Executes arbitrary module code; mutates the VM and the cache; spawns an interrupt-watcher goroutine.
  - **Exceptions:** Resolution and compilation errors; `module did not evaluate to a function`; any error thrown by the module body. Every failure path removes the placeholder from the cache.

- **Signature:** `installInterrupt(ctx context.Context) -> (stop func())`
  - **Behavior:** Spawns a goroutine that calls `VM.Interrupt(ctx.Err())` on cancellation. The returned stop function closes the done channel and clears the interrupt. A nil context yields a no-op.
  - **Side Effects:** Goroutine; VM interrupt state.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** One VM plus one module cache per instance, both long-lived. The VM is **not** reset between uses, so module-level state persists — tracked as a separate issue against [[runtime.modules]].
- **Performance/Scale Notes:** The exports cache means a module body executes once per VM. Because each `use` alias gets its own VM, a module required by two aliases is executed twice, in two separate object graphs. goja is single-threaded, so a VM must not be shared concurrently — the pool in [[runtime.modules]] is what enforces that.
- **Dependencies Risk:**
  - **Nested requires clear the outer interrupt.** `installInterrupt` calls `ClearInterrupt()` on entry and the stop function calls it again on exit. A `require` inside a module body therefore installs its own interrupt and, when it returns, **clears the parent's** — so a context cancellation during the remainder of the outer module body may not stop execution. The deeper the require chain, the more windows exist.
  - **Circular dependencies see a stale exports object if the factory reassigns `module.exports`.** The placeholder cached before execution is the original `exports`; the final cache update happens after. A module that does `module.exports = {…}` and participates in a cycle hands its circular consumer the empty placeholder. This is the classic CommonJS hazard, faithfully reproduced.
  - **Go-backed providers bypass the cache check.** The provider branch is evaluated **before** the cache lookup, so every `require` of a Go builtin re-fabricates a fresh exports object and overwrites the cache entry. Two references to the same builtin within one VM therefore get **different objects**, which matters for identity comparison and for any provider that tries to hold state.
  - **`__require` is a VM global saved and restored around each factory call.** Correct for the single-threaded nesting it is used with, but it means module code can read or overwrite `__require` and alter resolution for the remainder of the load.
  - **A failed load removes the cache entry**, so a subsequent `require` retries — which is the opposite of the registry's permanent compile-error caching, so retry semantics differ by which layer failed.
