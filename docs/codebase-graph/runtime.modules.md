---
id: runtime.modules
type: Class
language: Go
file_path: runtime/modules.go
tags: javascript, ffi, sandboxing, pooling, boundary-codec, security
---

# Node: runtime.ModuleBinding (JavaScript Call Boundary)

## 1. Architectural Role & Intent
The foreign-function boundary between Sentrie and embedded JavaScript. A `jsModule` holds the pool of `JSInstance` VMs for one resolved module path, shared by every `use` of that path; a `ModuleBinding` is one `use` statement's view of it (alias plus requested export names). `Call` acquires an instance, injects the execution start time, marshals arguments into goja values, installs a context-cancellation interrupt, invokes the export, and validates and unmarshals the return. It is where policy evaluation leaves the Go type system and comes back.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.modules` | `IMPORTS` | `ext.dop251.goja` | `goja.Runtime`, `goja.Value`, `AssertFunction`, `Interrupt`/`ClearInterrupt`. |
| `runtime.modules` | `IMPORTS` | `ext.jackc.puddle` | Pools `*JSInstance` per binding. |
| `runtime.modules` | `IMPORTS` | `ext.fatih.structs` | Converts struct returns to `map[string]any`. |
| `runtime.modules` | `DEPENDS_ON` | [[box]] | `IsBoundaryUndefined` drives undefined normalization on the way in. |
| `runtime.modules` | `DEPENDS_ON` | [[constants]] | `ExecutionStartTimeUnixKey` is the global injected into every VM. |
| `runtime.modules` | `READS_FROM` | [[runtime.exec_ctx]] | Reads `CreatedAt()` so JS-side time matches Go-side time. |
| [[runtime.executor]] | `CALLS` | [[runtime.modules]] | Constructs bindings and their pools during `bindUses`. |
| [[runtime.eval_call]] | `CALLS` | [[runtime.modules]] | A field-access callee on a bound alias dispatches to `Call`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `JSInstance` - `{ rt *goja.Runtime, exports *goja.Object, keys map[string]struct{} }` / `jsModule` - `{ key string, pool *puddle.Pool[*JSInstance] }` / `ModuleBinding` - `{ CanonicalKey, Alias string, exports []string, module *jsModule }`
  - **Behavior:** A `JSInstance` carries the module's **whole** exports object plus its own enumerable keys captured at load; it holds no restriction. The restriction lives on the `ModuleBinding`: a `use` that names exports reaches only those names, and a `use` that names none reaches only the module's own enumerable exports. `CanonicalKey` is the resolved module path, the cache key of the shared `jsModule`.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `(*jsModule).checkExports(ctx, names []string) -> error`
  - **Behavior:** Acquires an instance and fails if any requested name is absent from the exports object. Runs once per `use` at bind time.
  - **Side Effects:** Pool acquisition.
  - **Exceptions:** `module %s missing required export %q`; pool acquisition failure.

- **Signature:** `(ModuleBinding).Call(ctx, ec, fn string, args ...any) -> (any, error)`
  - **Behavior:** Acquires a pooled VM, sets the execution-start global, looks the export up through this binding's projection, asserts it is callable, spawns a watchdog goroutine that interrupts the VM on context cancellation, marshals arguments, invokes, then maps `undefined` and `null` returns to boundary sentinels before checking the return's reflect kind against an allowlist (map, slice, array, string, int64, float64, bool, struct) and exporting other values.
  - **Side Effects:** Executes arbitrary JavaScript; mutates the VM's global scope; releases the instance back to the pool.
  - **Exceptions:** `module has no JS binding`; pool acquisition failure; `function %s not found in module %q`; `export '%q' is not callable`; `unexpected return type %s`; any JS exception or interrupt.

- **Signature:** `normalizeBoundaryForJS(v any) -> any`
  - **Behavior:** Recursively rewrites Sentrie's boundary-undefined marker into `goja.Undefined()` through lists and maps, so JS sees a genuine `undefined` rather than a sentinel object.
  - **Side Effects:** None; builds new containers.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Pooled and reused. VMs are **not** reset between calls beyond the interrupt clear and the start-time global, so any global a module mutates persists across policy evaluations that land on the same instance.
- **Performance/Scale Notes:** Pools cap at 10 instances per binding, so concurrent evaluations exceeding that block on `Acquire`. Every call spawns a watchdog goroutine and allocates a channel. Argument marshalling deep-copies lists and maps.
- **Dependencies Risk:**
  - **VM state leaks between evaluations.** Because instances are pooled without a reset, a module that writes to a global carries that value into an unrelated policy's execution. This is a cross-tenant concern anywhere one executor serves multiple callers.
  - **The interrupt watchdog races with release.** `close(done)` runs before `poolInstance.Release()`, but the goroutine's `ClearInterrupt()` may still be in flight when the instance is handed to the next acquirer, which could clear an interrupt that acquirer just installed.
  - **The return-type allowlist rejects by kind, not by contract.** `reflect.Int64` is listed but JS numbers export as `float64`, and any unlisted kind produces `unexpected return type` naming the reflect type rather than the JS value - an opaque message for module authors.
  - **This is the sandbox boundary.** Permissions are applied when the VM is constructed in [[runtime.js]]; nothing in `Call` re-checks them, so a binding created with broad permissions stays broad for its whole cached lifetime.
