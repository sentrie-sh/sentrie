---
id: parser.access
type: Module / File
language: Go
file_path: parser/access.go
tags: field-access, index-access, infix-handler, navigation
---

# Node: parser.access (Field and Index Access)

## 1. Architectural Role & Intent
`parser/access.go` provides the two navigation infix handlers: `.field` (static field access by name) and `[expr]` (dynamic index access by evaluated key). Together they are how policies walk into documents, dicts, and lists. Both sit at `INDEX` precedence — the tightest binding in the ladder — so access chains attach to the nearest operand rather than to a surrounding expression.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.access` | `CALLS` | [[parser.expression]] | The index expression is parsed at `LOWEST` (bracket-delimited, so precedence resets). |
| `parser.access` | `CALLS` | [[ast]] | Emits `ast.NewFieldAccessExpression` and `ast.NewIndexAccessExpression`. |
| `parser.access` | `CALLS` | [[parser.parser]] | Uses `advance` and `advanceExpected`. |
| [[parser.lookups]] | `CALLS` | [[parser.access]] | Registered as infix handlers for `.` and `[`. |
| [[runtime.eval_access]] | `DEPENDS_ON` | [[ast]] | Resolves field and index navigation against boxed values at runtime. |
| [[box.value]] | `READS_FROM` | [[ast]] | Dict, document, and list kinds back the two access forms. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseFieldAccessExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `.` and requires an **identifier** — field names are static, never computed. Emits `FieldAccessExpression(left, name)`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` if the operator is not `.` (**without recording an error** — a silent failure, though unreachable via table dispatch) or if the field name is missing.

- **Signature:** `parseIndexAccessExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `[`, parses an arbitrary expression as the key at `LOWEST`, and requires `]`. Emits `IndexAccessExpression(left, index)`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing bracket or a failed index expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable; chains parse iteratively through the Pratt loop rather than recursively.
- **Dependencies Risk:**
  - **Both spans are wrong in opposite ways.** `FieldAccessExpression`'s range starts at the **dot**, not at the left operand, so an error on `user.email` highlights `.email` only. `IndexAccessExpression`'s range takes its `File` from the **closing** bracket while its `From` comes from the opening one — correct in practice, but it silently assumes both brackets are in the same file.
  - **No null-safe form.** There is no `?.` operator; navigating into an undefined value is a runtime concern resolved by [[box.value]]'s undefined semantics, not a syntactic one.
  - **`[` is triply overloaded** — list literal prefix, index-access infix, and computed map key — disambiguated only by parse position.
  - **Field names are identifiers only.** Keys that are not valid identifiers must be reached with `["…"]`, and a keyword used as a field name will fail because the lexer promotes it to a keyword kind before this production sees it.
