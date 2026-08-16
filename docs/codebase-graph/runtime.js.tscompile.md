---
id: runtime.js.tscompile
type: Module / File
language: Go
file_path: runtime/js/tscompile.go
tags: transpilation, esbuild, typescript, module-wrapping
---

# Node: js.tscompile (TypeScript Transpilation)

## 1. Architectural Role & Intent
Converts TypeScript or JavaScript source into ES2019 CommonJS via esbuild, then wraps the result in a function expression so goja can compile it once into a reusable module factory. It is the only place the engine's JavaScript dialect and target level are decided.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.tscompile` | `DEPENDS_ON` | `ext.esbuild` | `api.Transform` with a fixed option set. |
| [[runtime.js.registry]] | `CALLS` | `runtime.js.tscompile` | `TranspileTS` then `WrapAsIIFE` inside `programFor`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `TranspileTS(module *ModuleSpec, source string) -> (TranspileResult, error)`
  - **Behavior:** Selects the TypeScript loader when the module is a builtin or its extension is `.ts`, `.tsx`, `.mts`, or `.cts`; otherwise the JavaScript loader. Targets **ES2019**, emits **CommonJS**, and produces an external source map with source contents excluded.
  - **Side Effects:** None — pure string transformation.
  - **Exceptions:** `esbuild: %v` carrying only the **first** error's text.

- **Signature:** `WrapAsIIFE(js string) -> string`
  - **Behavior:** Wraps the transpiled code as `(function(require, module, exports) {; … ;})`, producing an expression that evaluates to the module factory. The leading and trailing semicolons guard against source that begins or ends in a way that would break the wrapper.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `isTS(module *ModuleSpec) -> bool`
  - **Behavior:** Builtins are always treated as TypeScript; on-disk modules are decided by lowercased extension.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; called once per module behind the registry's `sync.Once`.
- **Performance/Scale Notes:** esbuild is fast but not free, and this runs on the first load of every module. Minification is fully disabled, which keeps stack traces readable at the cost of larger compiled programs.
- **Dependencies Risk:**
  - **Only the first esbuild error is reported.** A file with several type or syntax problems surfaces them one at a time, each requiring a full reload cycle.
  - **The source map is parsed and stored but never used.** [[runtime.js.registry]] unmarshals it into `ModuleSpec.TranspiledMap` and nothing reads it, so runtime JavaScript errors report positions in the **transpiled, wrapped** source rather than the author's original. The plumbing for better diagnostics exists but is not connected.
  - **ES2019 is the ceiling.** Optional chaining, nullish coalescing, and `Array.prototype.flat` are ES2020 or later and are downlevelled or unavailable depending on the feature; module authors targeting modern syntax may hit surprises. The target is hardcoded with no override.
  - **`isTS` treats every builtin as TypeScript**, so a Go-provided builtin's TypeScript companion is always parsed with the TS loader regardless of its actual content.
  - **The IIFE wrapper shifts all line numbers by the wrapper prefix**, compounding the unused-source-map problem.
