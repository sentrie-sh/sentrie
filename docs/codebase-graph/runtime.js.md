---
id: runtime.js
type: System / Package
language: Go
file_path: runtime/js/
tags: javascript, typescript, goja, sandbox, ffi, standard-library
---

# Node: runtime/js (Embedded JavaScript Subsystem)

## 1. Architectural Role & Intent
Everything required to run TypeScript and JavaScript inside the policy engine: module resolution and compilation, a CommonJS-style runtime host, the Go-backed standard library, and the esbuild transpilation step. It exists so policy authors can drop into a real programming language for logic that is awkward to express declaratively, while keeping that code inside a controlled boundary with no filesystem, network, or process access beyond what this package deliberately exposes.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js` | `IMPORTS` | `ext.dop251.goja` | The JavaScript interpreter; every value crossing the boundary is a `goja.Value`. |
| `runtime.js` | `IMPORTS` | `ext.evanw.esbuild` | Transpiles TypeScript to CommonJS JavaScript before compilation. |
| `runtime.js` | `LAYERED_ON` | [[pack]] | `PackFile.Permissions` gates which environment variables the sandbox may see. |
| `runtime.js` | `LAYERED_ON` | [[constants]] | `APPNAME` builds the `@sentrie/*` builtin namespace. |
| [[runtime.modules]] | `DEPENDS_ON` | [[runtime.js]] | The runtime loads a fresh `AliasRuntime` per module call and calls into module exports. |
| [[runtime.executor]] | `CALLS` | [[runtime.js.registry]] | Registers every Go builtin and the TypeScript builtin at executor construction. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewRegistry(packRoot string) -> *Registry`
  - **Behavior:** Creates the module registry rooted at a pack directory. See [[runtime.js.registry]].
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewAliasRuntime(reg *Registry, baseDir string) -> *AliasRuntime`
  - **Behavior:** Creates one fresh `goja.Runtime` with a CommonJS `require` and a per-VM module cache. See [[runtime.js.alias_runtime]].
  - **Side Effects:** Allocates a VM.
  - **Exceptions:** None.

- **Signature:** `(*AliasRuntime).SetupStdLib(ctx, pack *pack.PackFile) -> error`
  - **Behavior:** Installs the standard library into the VM. **Must** be called before `Require` or any invocation. See [[runtime.js.stdlib]].
  - **Side Effects:** Sets VM globals.
  - **Exceptions:** Propagated from the env installer.

- **Signature:** `TranspileTS(module *ModuleSpec, source string) -> (TranspileResult, error)`
  - **Behavior:** esbuild transpilation to the module factory expression with an inline source map. See [[runtime.js.tscompile]].
  - **Side Effects:** None.
  - **Exceptions:** Every esbuild error as a `*SourceError` positioned in the original source.

- **Signature:** `SourceError{File, Line, Column, Message, Err}`
  - **Behavior:** A module load failure positioned in the original source, formatted `file:line:col: message`; `Unwrap` returns the underlying esbuild or goja error. `File` is `ModuleSpec.SourceName()`, for example `@local/lib/helpers.ts`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Builtin<Name>Go` - the Go module providers
  - **Behavior:** Fourteen `ModuleProvider` values (`uuid`, `crypto`, `time`, `encoding`, `collection`, `jwt`, `regex`, `net`, `hash`, `url`, `string`, `json`, `semver`, `math`) that fabricate a `module.exports` object directly in the VM. The TypeScript `js` module is registered separately from embedded source. See [[runtime.js.builtins]].
  - **Side Effects:** Sets properties on a new VM object.
  - **Exceptions:** `ErrPermissionDenied` where a builtin is permission-gated. Functions throw JavaScript exceptions on failure.

## 4. Operational Context & Gotchas
- **Statefulness:** The `Registry` is shared and long-lived, caching compiled programs. Each `AliasRuntime` owns exactly one VM with its own module-exports cache. [[runtime.modules]] creates one per module call and drops it afterwards.
- **Performance/Scale Notes:** A successful transpile and compile happens **once per module** and the `goja.Program` is reused; a failure is not kept past an edit of the source (see [[runtime.js.registry]]). goja reads each program's inline source map, so runtime exception stacks name the `.ts` file and original line with no frame rewriting. VM creation is the expensive part: since #105 every module call pays it (about 30 us and 55 KB for a small module), in exchange for calls that cannot see each other's module state. Require resolution is memoised per registry so the per-call load does no filesystem work. goja is an interpreter - it does not JIT - so hot loops in JavaScript are markedly slower than the equivalent Go builtin.
- **Dependencies Risk:**
  - **The sandbox is defined by omission, not by policy.** goja provides no filesystem, network, or process access by default, and this package adds none - the `net` builtin is pure CIDR arithmetic with no sockets. The boundary therefore holds only as long as no future builtin introduces real I/O. There is no allowlist enforcing that.
  - **Environment access is the one deliberate hole**, gated per-key by pack permissions. A nil permissions block exposes nothing, which is the correct default.
  - **`time.now()` is pinned to the execution start timestamp** injected as a VM global, so JavaScript sees the same clock (Unix seconds) as the policy evaluator's `now()`. A VM without the global throws; there is no wall-clock fallback (#153).
  - **Builtin failures throw.** Every Go builtin raises a JavaScript exception through `throwErr`, never a returned Error object (#111). See [[runtime.js.builtins]] for the predicate and `null` conventions.
