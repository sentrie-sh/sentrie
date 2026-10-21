---
id: runtime.modules
type: Class
language: Go
file_path: runtime/modules.go
tags: javascript, ffi, sandboxing, pooling, boundary-codec, security
---

# Node: runtime.ModuleBinding (JavaScript Call Boundary)

## 1. Architectural Role & Intent
The foreign-function boundary between Sentrie and embedded JavaScript. A `jsModule` is one resolved module path, shared by every `use` of that path, and loads a fresh `JSInstance` VM for every call; a `ModuleBinding` is one `use` statement's view of it (alias plus requested export names). `Call` acquires a fresh instance (its execution start time pinned at load), marshals arguments into goja values, installs a context-cancellation interrupt, invokes the export, and validates and unmarshals the return. It is where policy evaluation leaves the Go type system and comes back.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.modules` | `IMPORTS` | `ext.dop251.goja` | `goja.Runtime`, `goja.Value`, `AssertFunction`. |
| `runtime.modules` | `CALLS` | [[runtime.js]] | `InterruptOnDone` bounds each call by its context. |
| `runtime.modules` | `IMPORTS` | `ext.fatih.structs` | Converts struct returns to `map[string]any`. |
| `runtime.modules` | `DEPENDS_ON` | [[box]] | `IsBoundaryUndefined` drives undefined normalization on the way in. |
| `runtime.modules` | `DEPENDS_ON` | [[constants]] | `ExecutionStartTimeUnixKey` is the global injected into every VM. |
| `runtime.modules` | `READS_FROM` | [[runtime.exec_ctx]] | Reads `CreatedAt()` so JS-side time matches Go-side time. |
| [[runtime.executor]] | `CALLS` | [[runtime.modules]] | Constructs bindings and their pools during `bindUses`. |
| [[runtime.eval_call]] | `CALLS` | [[runtime.modules]] | A field-access callee on a bound alias dispatches to `Call`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `JSInstance` - `{ rt *goja.Runtime, exports *goja.Object, keys map[string]struct{} }` / `jsModule` - `{ key string, load instanceLoader, slots chan struct{}, probeMu sync.Mutex, probe *JSInstance }` / `ModuleBinding` - `{ CanonicalKey, Alias string, exports []string, module *jsModule }`
  - **Behavior:** A `JSInstance` carries the module's **whole** exports object plus its own enumerable keys captured at load; it holds no restriction. The restriction lives on the `ModuleBinding`: a `use` that names exports reaches only those names, and a `use` that names none reaches only the module's own enumerable exports. `CanonicalKey` is the resolved module path, the cache key of the shared `jsModule`.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `newJSModule(ctx, key, load) -> (*jsModule, error)` / `(*jsModule).acquire(ctx, start time.Time) -> (*JSInstance, release func(), error)`
  - **Behavior:** `newJSModule` loads the module once into its **probe** instance, which verifies it loads and answers export checks; the probe never serves a call. `acquire` takes one of 10 slots (waiting honours `ctx`), then loads the module into a **fresh** `goja.Runtime` through the registry's cached `goja.Program`. `release` frees the slot; the VM is dropped.
  - **Side Effects:** Executes the module body once per acquisition.
  - **Exceptions:** `ctx.Err()` while waiting for a slot; module load errors.

- **Signature:** `(*jsModule).checkExports(names []string) -> error`
  - **Behavior:** Fails if any requested name is absent from the probe's exports object. Runs once per `use` at bind time, under `probeMu` because reading an export may run a getter.
  - **Side Effects:** None.
  - **Exceptions:** `module %s missing required export %q`.

- **Signature:** `(ModuleBinding).Call(ctx, ec, fn string, args ...any) -> (any, error)`
  - **Behavior:** Acquires a fresh VM, looks the export up through this binding's projection, asserts it is callable, installs `js.InterruptOnDone` for the call, marshals arguments, invokes, then maps `undefined` and `null` returns to boundary sentinels before checking the return's reflect kind against an allowlist (map, slice, array, string, int64, float64, bool, struct) and exporting other values.
  - **Side Effects:** Executes the module body and the export in a VM nobody else sees; frees its slot on return.
  - **Exceptions:** `module has no JS binding`; slot wait or module load failure; `function %s not found in module %q`; `export '%q' is not callable`; `unexpected return type %s`; any JS exception or interrupt.

- **Signature:** `normalizeBoundaryForJS(v any) -> any`
  - **Behavior:** Recursively rewrites Sentrie's boundary-undefined marker into `goja.Undefined()` through lists and maps, so JS sees a genuine `undefined` rather than a sentinel object.
  - **Side Effects:** None; builds new containers.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Isolated per call ([#105](https://github.com/sentrie-sh/sentrie/issues/105)). Each call loads the module into a VM no other call has touched, so module-level state (counters, memo tables, `let` bindings) lives for exactly that call - including between two calls in one rule body and between concurrent rules under `ExecPolicy`.
- **Performance/Scale Notes:** A call costs a VM construction plus the module body: about 30 us and 55 KB for a small module, against about 1 us for the old pooled reuse. Compilation is still once per module. At most 10 VMs per module are live at once; further calls wait. Argument marshalling deep-copies lists and maps.
- **Dependencies Risk:**
  - **Module authors cannot rely on module state between calls.** Anything that must persist across calls has to come in through arguments or facts. A module whose top level is expensive pays that cost on every call.
  - **The return-type allowlist rejects by kind, not by contract.** `reflect.Int64` is listed but JS numbers export as `float64`, and any unlisted kind produces `unexpected return type` naming the reflect type rather than the JS value - an opaque message for module authors.
  - **This is the sandbox boundary.** Permissions are applied when the VM is constructed in [[runtime.js]]; nothing in `Call` re-checks them, and every fresh VM is built with the same pack permissions.
