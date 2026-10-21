---
id: runtime.js.registry
type: Class
language: Go
file_path: runtime/js/registry.go
tags: module-resolution, compilation, caching, path-traversal, security
---

# Node: js.Registry (Module Resolution and Compilation)

## 1. Architectural Role & Intent
Owns module identity and compilation for the whole engine. It resolves `use` references and `require` specifiers into canonical keys, decides whether a module is a Go builtin, a TypeScript builtin, or a file on disk, and compiles each successfully exactly once into a reusable `goja.Program`. It is shared across every VM, so compilation cost is paid once per module for the process lifetime.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.registry` | `IMPORTS` | `ext.dop251.goja` | `goja.Compile` produces the cached `Program`. |
| `runtime.js.registry` | `CALLS` | [[runtime.js.tscompile]] | `TranspileTS` before compilation. |
| `runtime.js.registry` | `READS_FROM` | [[infra.filesystem]] | `os.ReadFile` for on-disk modules; `os.Stat` for extension probing. |
| `runtime.js.registry` | `READS_FROM` | [[capabilities]] | `checkGrant` refuses ungranted builtin modules and, when local modules are not granted, every pack-local file. |
| `runtime.js.registry` | `DEPENDS_ON` | [[constants]] | `APPNAME` forms the `@sentrie/` builtin key prefix. |
| [[runtime.executor]] | `MUTATES` | `runtime.js.registry` | Registers all Go and TypeScript builtins at construction. |
| [[runtime.js.alias_runtime]] | `CALLS` | `runtime.js.registry` | `LoadRequire` and `programFor` on every `require`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `RegisterGoBuiltin(name string, provider ModuleProvider)` / `RegisterTSBuiltin(name, tsSource string)`
  - **Behavior:** Registers under the key `@<APPNAME>/<name>`. A Go provider takes precedence over a TypeScript source of the same name.
  - **Side Effects:** Mutates the builtin maps.
  - **Exceptions:** None - a duplicate registration silently overwrites.

- **Signature:** `PrepareUse(localFrom string, libFrom []string, fileDir string) -> (*ModuleSpec, error)`
  - **Behavior:** Resolves a `use` statement's target, creates the module spec, and warm-compiles it best-effort. `libFrom[0] == APPNAME` selects a builtin; `"local"` resolves under the pack root via `relativeToLocal` (containment-checked, symlink-resolved); anything else is rejected pending vendor support.
  - **Side Effects:** Populates the module map; may read and compile a file.
  - **Exceptions:** `unsupported library from: %v`; `module %s not found`; compilation errors.

- **Signature:** `LoadRequire(fromDir, spec string) -> (*ModuleSpec, error)`
  - **Behavior:** The `require()` counterpart. Successful resolutions are memoised per `(fromDir, spec)`, because every module call loads into a fresh VM and would otherwise repeat the containment and symlink checks on disk. Handles `@sentrie/…` builtins, `@local/…` pack-relative paths, and `./` or `/` relative paths. All on-disk resolution goes through `relativeToLocal`, which rejects `..` traversal and symlink escapes outside `PackRoot`. Bare specifiers are rejected - there is no `node_modules` resolution.
  - **Side Effects:** As above.
  - **Exceptions:** `unsupported require spec: %q`; `relative path is outside the packroot: %s`; `module %s not found`.

- **Signature:** `SetCapabilities(grant *capabilities.Capabilities)` (in `capabilities.go`)
  - **Behavior:** Installs the operator's grant. After resolution and before a spec is created, `PrepareUse` and `LoadRequire` call `checkGrant`: a builtin key the grant does not list, or any non-builtin (pack-local) key when local modules are not granted, fails. `LoadRequire` serves every `require()`, so TypeScript is held to the same grant as `use`. A nil grant allows everything.
  - **Side Effects:** Sets the grant. Like the `Register*` calls, it is unsynchronised and must happen during executor construction.
  - **Exceptions:** `module %q: not granted by the capabilities file` / `local module %q: ...`, wrapping `xerr.ErrNotGranted`.

- **Signature:** `CheckUseRequires(localFrom, libFrom, fileDir) -> error` (in `capabilities.go`)
  - **Behavior:** With a grant installed: `PrepareUse` (which checks the module itself), then for pack-local code a scan of `TranspiledCode` for static `require("...")` (esbuild emits every import that way), resolving each through `LoadRequire` (so `checkGrant` applies) and recursing into pack-local requires. A regex scan: it may over-match a `require("...")` inside a string literal, which fails closed, and it cannot see dynamic `require(expr)`, which the runtime gate still refuses.
  - **Side Effects:** Compiles the modules it visits.
  - **Exceptions:** `<module>: require(%q): <grant or resolution error>`.

- **Signature:** `programFor(m *ModuleSpec) -> (*goja.Program, error)`
  - **Behavior:** Returns the cached program if one compiled; otherwise, under the spec's `compileMu`, reads the source, transpiles to the factory expression and compiles it under the spec's `SourceName()`. Only a success is kept for the process lifetime. A failure is remembered with the source text that produced it and returned again while that text is unchanged; an edited source is compiled again, and a read error is never remembered. A Go-backed module returns `(nil, nil)` because there is no program to run.
  - **Side Effects:** On success populates `TranspiledCode` and `Program`, then publishes the program through an `atomic.Pointer`, so later calls take no lock. These fields are read only after `programFor` succeeds.
  - **Exceptions:** `builtin not found: %s`; file read errors (unwrapped); esbuild errors and goja compile errors, each a `*SourceError` positioned in the original source. A goja **parse** error carries no file, so its `SourceError` names the module and says its position is in the transpiled output.

- **Signature:** `(*ModuleSpec).SourceName() -> string`
  - **Behavior:** The name errors and stack frames report: the pack-relative key with the resolved file's base name (so `@local/lib/helpers` becomes `@local/lib/helpers.ts`), or the key itself for a builtin.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Registry).CompiledPrograms() -> int64`
  - **Behavior:** How many module programs this registry has compiled; one per module for the registry's lifetime.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `BuiltinModuleKey(name) -> string` / `IsBuiltinModuleKey(key) -> bool` / `IsNondeterministicExport(moduleKey, export) -> bool` / `NondeterministicExports(moduleKey) -> []string` / `NondeterministicModules() -> []string` (in `nondeterminism.go`)
  - **Behavior:** Nondeterminism metadata for builtin module exports, the JavaScript counterpart of `builtins.Decl.Nondeterministic`. A static table keyed by `@sentrie/<name>`: `js` (`random`; its `now` was removed in #207), `time` (`now`), `uuid` (`v4`, `v6`, `v7`). [[cmd]] passes `IsNondeterministicExport` to the index's `NondeterminismWarnings`, so the index never imports this package. A runtime test asserts every listed module is registered by [[runtime.executor]] and every listed export exists.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*ModuleSpec).KeyOrPath() -> string`
  - **Behavior:** The canonical identifier - key if set, otherwise path. Used as the cache key throughout the system.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Long-lived and shared. The module map is guarded by an `RWMutex` with correct double-checked locking; each spec's compile attempts are serialized by its own `compileMu`, with a lock-free `atomic.Pointer` fast path once compiled. Concurrent first callers wait on the mutex and share one program.
- **Path containment:** `@local/` specifiers, `use … from local …`, and `./`/`/` relative requires all route through `relativeToLocal` → `confineToPackRoot`. Lexical `..` checks run on cleaned paths; existing path components are symlink-resolved so a link under the pack root cannot target files outside it.
- **Performance/Scale Notes:** Compilation happens once per module per process. `getOrCreateModule` takes the read lock first and only escalates on a miss, so the steady-state path is read-only. Extension probing does up to two `os.Stat` calls, but only on first resolution.
- **Dependencies Risk:**
  - **`RegisterGoBuiltin` and `RegisterTSBuiltin` write their maps without holding a lock**, unlike every other map in the type. This is safe only because registration happens during executor construction before any concurrent use - an invariant nothing enforces.
  - **A success is permanent, a failure is not.** Once a module compiles, later edits to its file are not picked up until restart (there is no hot reload). A failing on-disk module is re-read on every `PrepareUse` and `require`, which happens per policy evaluation, so a broken module costs one file read per call; it is re-transpiled only when its content changed.
  - **Compile errors appear only when a policy first binds the module**; `sentrie validate` does not compile modules ([#155](https://github.com/sentrie-sh/sentrie/issues/155)).
  - **Extension probing prefers `.ts` over `.js`** silently, so a directory containing both resolves to the TypeScript file with no diagnostic.
  - **`getOrCreateModule` returns nil rather than an error** when nothing resolves, forcing every caller to translate that into `module %s not found` and losing the reason.
  - **Bare specifiers are unsupported**, so no npm dependency can be used; the comment marks vendor resolution as future work.
