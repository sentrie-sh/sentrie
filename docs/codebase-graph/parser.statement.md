---
id: parser.statement
type: Function / Endpoint
language: Go
file_path: parser/statement.go
tags: statement-dispatch, scope-enforcement, front-end
---

# Node: parser.parseStatement (Top-Level Statement Dispatcher)

## 1. Architectural Role & Intent
`parser/statement.go` is the dispatcher for **namespace-scope** declarations: it maps the head token to a handler in `statementHandlers` and rejects anything else. Its one piece of real logic is a scope guard that intercepts the four policy-metadata keywords (`title`, `description`, `version`, `tag`) before dispatch and reports a targeted "only allowed inside a policy" error instead of the generic unexpected-token message — a small but deliberate diagnostic-quality investment for the most common misplacement mistake.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.statement` | `READS_FROM` | [[parser.lookups]] | Looks up `statementHandlers` — the **top-level** table only. |
| `parser.statement` | `CALLS` | [[parser.parser]] | Reads `head()` and reports through `errorf`. |
| `parser.statement` | `DEPENDS_ON` | [[tokens]] | Guards on the metadata keyword kinds; dispatch is keyed by kind. |
| `parser.statement` | `CALLS` | [[ast]] | Returns `ast.Statement` values produced by the delegated handlers. |
| [[parser.parse]] | `CALLS` | [[parser.statement]] | The program loop calls this once per statement. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Parser).registerStatementHandler(tokenType: tokens.Kind, fn: statementParser)`
  - **Behavior:** Inserts into the top-level statement table. Called only from `registerParseFns`.
  - **Side Effects:** Mutates `statementHandlers`.
  - **Exceptions:** None; duplicate keys overwrite silently.

- **Signature:** `parseStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Guards the policy-only metadata keywords, then dispatches on head-token kind. Registered top-level heads are `namespace`, `LineComment`, `TrailingComment`, `policy`, `shape`, `derive`, and `export`. The handler owns all token consumption — this function consumes nothing itself.
  - **Side Effects:** None directly; the delegated handler consumes tokens and may record errors.
  - **Exceptions:** Returns `nil` after recording either `'<kind>' is only allowed inside a policy` (metadata keyword at namespace scope) or `unexpected token '<kind>'` (no registered handler).

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless dispatch over the parser's handler map.
- **Performance/Scale Notes:** One switch and one map probe per statement. Negligible.
- **Dependencies Risk:**
  - **This is only half the dispatch story.** Statements *inside* a `policy { }` block go through the separate `policyStatementHandlers` table, consulted by the policy production — not by this function. Reading this file alone gives an incomplete picture of what is parseable; see [[parser.lookups]] for both tables.
  - **`nil` return without a consumed token.** When no handler matches, this returns `nil` **without advancing**, so a caller that loops on `parseStatement` and does not check `p.err` will spin forever on the offending token. [[parser.parse]] avoids this by returning as soon as `p.err` is non-nil — any new caller must do the same.
  - **The metadata guard duplicates knowledge.** The four keywords are hardcoded here and separately registered in the policy table; adding a fifth policy-only keyword requires touching both places or the error degrades to the generic message.
