---
id: runtime.js.tscompile
type: Module / File
language: Go
file_path: runtime/js/tscompile.go
tags: transpilation, esbuild, typescript, module-wrapping
---

# Node: js.tscompile (TypeScript Transpilation)

## 1. Architectural Role & Intent
Converts TypeScript or JavaScript source into ES2019 CommonJS via esbuild, emitted inside a function expression so goja can compile it once into a reusable module factory, with an inline source map that lets goja report positions in the original source. It is the only place the engine's JavaScript dialect and target level are decided.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.tscompile` | `IMPORTS` | `ext.evanw.esbuild` | `api.Transform` with a fixed option set. |
| [[runtime.js.registry]] | `CALLS` | `runtime.js.tscompile` | `TranspileTS` inside `programFor`, before `goja.Compile`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `TranspileTS(module *ModuleSpec, source string) -> (TranspileResult, error)`
  - **Behavior:** Selects the TypeScript loader when the module is a builtin or its extension is `.ts`, `.tsx`, `.mts`, or `.cts`; otherwise the JavaScript loader. Targets **ES2019** and emits **CommonJS**. The factory wrapper `(function(require, module, exports) {` ... `})` is passed to esbuild as banner and footer, so `TranspileResult.Code` is the whole factory expression, ending in an inline `//# sourceMappingURL=data:` comment (source contents excluded) whose mappings already account for the banner line. The map's source is the base name of `ModuleSpec.SourceName()`, which goja resolves against the program name back to that source name.
  - **Side Effects:** None - pure string transformation.
  - **Exceptions:** Every esbuild error, joined with `errors.Join`, each a `*SourceError` (`errors.go`) of the form `<source name>:<line>:<col>: <text>`, positioned in the original source (esbuild's 0-based column is made 1-based).

- **Signature:** `isTS(module *ModuleSpec) -> bool`
  - **Behavior:** Builtins are always treated as TypeScript; on-disk modules are decided by lowercased extension.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; called by the registry's `programFor` on each compile attempt (once per module on success).
- **Performance/Scale Notes:** esbuild is fast but not free, and this runs on the first load of every module. Minification is fully disabled, which keeps stack traces readable at the cost of larger compiled programs. The inline map is base64 in the program source, so each compiled program carries its map for the process lifetime.
- **Dependencies Risk:**
  - **esbuild stops at the first syntax error.** Every message it does return is reported, but a parse failure yields one; scope errors such as duplicate declarations can yield several.
  - **The source map only works while goja finds it.** goja looks for `//# sourceMappingURL=` past trailing lines that are empty or exactly `})`; changing the footer or appending after the map comment silently drops back to transpiled positions.
  - **Mapped file names go through URL resolution in goja**, so a module file name with spaces or `%` is reported percent-encoded, and `#` or `?` in a name would be misread.
  - **goja maps columns as reported by the source map library**, which can be one short of the 1-based column in the source; lines are exact.
  - **ES2019 is the ceiling.** Optional chaining, nullish coalescing, and `Array.prototype.flat` are ES2020 or later and are downlevelled or unavailable depending on the feature; module authors targeting modern syntax may hit surprises. The target is hardcoded with no override.
  - **`isTS` treats every builtin as TypeScript**, so a Go-provided builtin's TypeScript companion is always parsed with the TS loader regardless of its actual content.
