---
id: parser.use
type: Function / Endpoint
language: Go
file_path: parser/use.go
tags: declaration, module-import, javascript-interop, aliasing
---

# Node: parser.parseUseStatement (Module Import Declaration)

## 1. Architectural Role & Intent
Parses `use { f, g } from ("./mod.js" | @pkg/mod) [as alias]` — the bridge between policy code and external JavaScript modules executed by [[runtime.js]]. It exists to distinguish the two source forms (a quoted relative path versus an `@`-prefixed library path) and to compute a sensible default alias so that most imports need no explicit `as`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.use` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect`, `canExpect`, `canExpectAnyOf`, `errorf`. |
| `parser.use` | `CALLS` | [[ast]] | Emits `ast.NewUseStatement(modules, relativeFrom, libFrom, alias, span)`. |
| `parser.use` | `IMPORTS` | `std.strings` | Splits a relative path on `/` to derive the default alias. |
| [[parser.policy]] | `CALLS` | [[parser.use]] | Registered for `tokens.KeywordUse` in the policy-scope table — `use` is policy-scoped, not namespace-scoped. |
| [[runtime.js]] | `DEPENDS_ON` | [[parser.use]] | Resolves and loads the named module, exposing the listed exports under the alias. |
| [[pack]] | `READS_FROM` | [[parser.use]] | Permission grants bound what a resolved module may access. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseUseStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `use`, then a **required non-empty** brace list of identifiers separated by commas, then `from`, then either a `String` (stored as `RelativeFrom`) or `@ident (/ ident)*` (stored as `LibFrom` segments). The default alias is the last path segment of whichever form was used; an explicit `as <ident>` overrides it.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing `use`/`{`/`}`/`from`, a non-identifier in the import list, a missing comma between entries, a source that is neither string nor `@`, or a missing identifier after `as`. The source-form failure reports `expected string or '@' for module import`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Empty import lists are rejected by construction.** The first identifier is consumed unconditionally, so `use {} from "x"` fails with a confusing `expected Ident, got }` rather than a clear "at least one import required".
  - **Unterminated brace list can run away.** The loop condition tests only for `}` — not EOF — so a missing closing brace relies on `expect(PunctComma)` failing to break out. The resulting diagnostic points at whatever token followed, not at the `use`.
  - **Two mutually exclusive source fields.** Consumers must check `RelativeFrom != ""` before falling back to `LibFrom`; exactly one is populated, and nothing in the AST enforces that invariant.
  - **Alias derivation assumes a path shape.** For a relative source the alias is the substring after the last `/`, including any file extension — `from "./util.js"` defaults the alias to `util.js`, not `util`. Explicit `as` is advisable for file-path imports.
  - **`/` inside `@pkg/mod` reuses the division token**, the same overload noted in [[parser.fqn]].
  - **Nothing is resolved here.** Whether the module exists, exports the named symbols, or is permitted by the pack's [[pack]] permissions is determined at runtime.
