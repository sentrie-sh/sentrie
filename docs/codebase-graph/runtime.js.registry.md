---
id: runtime.js.registry
type: Class
language: Go
file_path: runtime/js/registry.go
tags: module-resolution, compilation, caching, path-traversal, security
---

# Node: js.Registry (Module Resolution and Compilation)

## 1. Architectural Role & Intent
Owns module identity and compilation for the whole engine. It resolves `use` references and `require` specifiers into canonical keys, decides whether a module is a Go builtin, a TypeScript builtin, or a file on disk, and compiles each exactly once into a reusable `goja.Program`. It is shared across every VM, so compilation cost is paid once per module for the process lifetime.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.registry` | `IMPORTS` | `ext.dop251.goja` | `goja.Compile` produces the cached `Program`. |
| `runtime.js.registry` | `CALLS` | [[runtime.js.tscompile]] | `TranspileTS` then `WrapAsIIFE` before compilation. |
| `runtime.js.registry` | `READS_FROM` | [[infra.filesystem]] | `os.ReadFile` for on-disk modules; `os.Stat` for extension probing. |
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
  - **Behavior:** The `require()` counterpart. Handles `@sentrie/…` builtins, `@local/…` pack-relative paths, and `./` or `/` relative paths. All on-disk resolution goes through `relativeToLocal`, which rejects `..` traversal and symlink escapes outside `PackRoot`. Bare specifiers are rejected - there is no `node_modules` resolution.
  - **Side Effects:** As above.
  - **Exceptions:** `unsupported require spec: %q`; `relative path is outside the packroot: %s`; `module %s not found`.

- **Signature:** `programFor(m *ModuleSpec) -> (*goja.Program, error)`
  - **Behavior:** Behind `sync.Once`: reads the source, transpiles, parses the source map, wraps as an IIFE factory, and compiles. A Go-backed module returns `(nil, nil)` because there is no program to run.
  - **Side Effects:** Populates `TranspiledCode`, `TranspiledMap`, and `Program` on the spec.
  - **Exceptions:** `builtin not found: %s`; file read errors; esbuild errors; source-map unmarshal errors; goja compile errors.

- **Signature:** `(*ModuleSpec).KeyOrPath() -> string`
  - **Behavior:** The canonical identifier - key if set, otherwise path. Used as the cache key throughout the system.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Long-lived and shared. The module map is guarded by an `RWMutex` with correct double-checked locking; each spec's compilation is guarded by its own `sync.Once`.
- **Path containment:** `@local/` specifiers, `use … from local …`, and `./`/`/` relative requires all route through `relativeToLocal` → `confineToPackRoot`. Lexical `..` checks run on cleaned paths; existing path components are symlink-resolved so a link under the pack root cannot target files outside it.
- **Performance/Scale Notes:** Compilation happens once per module per process. `getOrCreateModule` takes the read lock first and only escalates on a miss, so the steady-state path is read-only. Extension probing does up to two `os.Stat` calls, but only on first resolution.
- **Dependencies Risk:**
  - **`RegisterGoBuiltin` and `RegisterTSBuiltin` write their maps without holding a lock**, unlike every other map in the type. This is safe only because registration happens during executor construction before any concurrent use - an invariant nothing enforces.
  - **A compilation failure is cached permanently.** `sync.Once` fires regardless of outcome and `m.err` is retained, so a module that failed to compile - including for a transient reason such as a file read error - can never be retried for the process lifetime.
  - **Extension probing prefers `.ts` over `.js`** silently, so a directory containing both resolves to the TypeScript file with no diagnostic.
  - **`getOrCreateModule` returns nil rather than an error** when nothing resolves, forcing every caller to translate that into `module %s not found` and losing the reason.
  - **Bare specifiers are unsupported**, so no npm dependency can be used; the comment marks vendor resolution as future work.
