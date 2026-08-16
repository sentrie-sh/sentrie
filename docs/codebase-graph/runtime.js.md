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
| `runtime.js` | `LAYERED_ON` | [[constants]] | `APPNAME` builds the `@sentrie/*` builtin namespace; `ExecutionStartTimeUnixKey` pins time. |
| [[runtime.modules]] | `DEPENDS_ON` | [[runtime.js]] | The runtime pools `AliasRuntime` instances and calls into module exports. |
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

- **Signature:** `TranspileTS(module *ModuleSpec, source string) -> (TranspileResult, error)` / `WrapAsIIFE(js string) -> string`
  - **Behavior:** esbuild transpilation and the module factory wrapper. See [[runtime.js.tscompile]].
  - **Side Effects:** None.
  - **Exceptions:** `esbuild: %v` on the first transform error.

- **Signature:** `Builtin<Name>Go` — the Go module providers
  - **Behavior:** Sixteen `ModuleProvider` values (`uuid`, `crypto`, `time`, `encoding`, `collection`, `jwt`, `regex`, `net`, `hash`, `url`, `string`, `json`, `semver`, `math`, plus base64 and the TypeScript `js` module) that fabricate a `module.exports` object directly in the VM. See [[runtime.js.builtins]].
  - **Side Effects:** Sets properties on a new VM object.
  - **Exceptions:** `ErrPermissionDenied` where a builtin is permission-gated.

## 4. Operational Context & Gotchas
- **Statefulness:** The `Registry` is shared and long-lived, caching compiled programs. Each `AliasRuntime` owns exactly one VM with its own module-exports cache. VMs are pooled by [[runtime.modules]].
- **Performance/Scale Notes:** Transpilation and compilation happen **once per module** behind a `sync.Once`; execution reuses the compiled `goja.Program`. VM creation is the expensive part, which is why VMs are pooled rather than created per call. goja is an interpreter — it does not JIT — so hot loops in JavaScript are markedly slower than the equivalent Go builtin.
- **Dependencies Risk:**
  - **`require("@local/../…")` escapes the pack root.** The `@local/` branch of `resolveRequire` skips the containment check that the relative-path branch applies, enabling arbitrary file reads. Filed as [#110](https://github.com/sentrie-sh/sentrie/issues/110); see [[runtime.js.registry]].
  - **The sandbox is defined by omission, not by policy.** goja provides no filesystem, network, or process access by default, and this package adds none — the `net` builtin is pure CIDR arithmetic with no sockets. The boundary therefore holds only as long as no future builtin introduces real I/O. There is no allowlist enforcing that.
  - **Environment access is the one deliberate hole**, gated per-key by pack permissions. A nil permissions block exposes nothing, which is the correct default.
  - **`time.now()` is pinned to the execution start timestamp** injected as a VM global, so JavaScript sees the same clock as the policy evaluator — with a silent fallback to the real wall clock if the global is missing.
  - **Builtins return error objects rather than throwing.** See [[runtime.js.builtins]]; this is the most likely source of quiet incorrect behaviour in module code.
  - **VMs are pooled without reset**, so module-level state persists across evaluations. Filed as [#105](https://github.com/sentrie-sh/sentrie/issues/105); see [[runtime.modules]].
