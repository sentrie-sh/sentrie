---
id: parser.data
type: Function / Endpoint
language: Go
file_path: parser/data.go
tags: declaration, pack-data, static-data, contextual-keyword
---

# Node: parser.parseDataStatement (Pack Data Declaration)

## 1. Architectural Role & Intent
Parses `data <ident> [':' <type>] ['as' <ident>]`, the policy-scoped declaration that reads one entry of the pack manifest's `[data]` table (#169). It is the syntax half of static reference data: the declaration makes a policy's dependency on pack data explicit, and the optional type is the shape the loaded value is checked against. `data` is a **contextual keyword**, recognised only at the head of a policy statement, so it is not added to the lexer and every other use of `data` as an identifier keeps working.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.data` | `CALLS` | [[parser.typeref]] | Parses the optional declared type. |
| `parser.data` | `CALLS` | [[parser.parser]] | Uses `advance`, `advanceExpected`, `canExpect`, `errorf`. |
| `parser.data` | `CALLS` | [[ast]] | Emits `ast.NewDataStatement(name, type, alias, span)`; `Type` is nil when omitted. |
| [[parser.policy]] | `CALLS` | [[parser.data]] | `parsePolicyStatement` dispatches here when the head is the identifier `data`, before the token-kind table. |
| [[index.policy]] | `DEPENDS_ON` | [[parser.data]] | Indexes each declaration under its alias in `Policy.Data`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseDataStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes the `data` identifier and the entry name, then an optional `: <type>` and an optional `as <ident>` (the alias defaults to the name). The span is extended after each clause.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing name, a failed type reference, or a missing alias after `as`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Contextual, not reserved.** A bare identifier at the head of a policy statement was always a parse error, so claiming `data` there changes no valid program. Making `data` a lexer keyword instead would break every fact, let, field and parameter named `data`.
  - **Names that are keywords cannot be declared.** The entry name must lex as an identifier; the loader rejects keyword-named `[data]` entries for that reason.
  - **Nothing here checks the manifest.** Whether the name exists in `[data]` is decided by [[index.pack_data]].
