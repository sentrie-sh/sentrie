---
id: parser.policy
type: Function / Endpoint
language: Go
file_path: parser/policy.go
tags: declaration, policy-block, scope, statement-dispatch
---

# Node: parser.parseThePolicyStatement (Policy Block and Inner Dispatch)

## 1. Architectural Role & Intent
Parses `policy <ident> { … }` and — importantly — owns the **second statement dispatcher**, `parsePolicyStatement`, which routes the body against `policyStatementHandlers` rather than the top-level table. This file is therefore where Sentrie's two-level scoping is realised in the front-end: `title`/`rule`/`fact`/`let`/`use` are legal only because this dispatcher exists, and top-level-only forms like `namespace` are unreachable from inside a policy for the same reason.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.policy` | `READS_FROM` | [[parser.lookups]] | `parsePolicyStatement` looks up `policyStatementHandlers`; `registerPolicyStatementHandler` fills it. |
| `parser.policy` | `CALLS` | [[parser.parser]] | Uses `expect`, `advanceExpected`, `hasTokens`, `canExpect`, `canExpectAnyOf`. |
| `parser.policy` | `CALLS` | [[ast]] | Emits `ast.NewPolicyStatement` wrapping the collected body statements. |
| `parser.policy` | `CALLS` | [[parser.rule]] | Via the handler table, for each `rule` in the body. |
| `parser.policy` | `CALLS` | [[parser.fact]] | Via the handler table, for each `fact`. |
| `parser.policy` | `CALLS` | [[parser.let]] | Via the handler table, for each `let`. |
| `parser.policy` | `CALLS` | [[parser.use]] | Via the handler table, for each `use`. |
| `parser.policy` | `CALLS` | [[parser.export_rule]] | Via the handler table, for `export decision of …`. |
| [[parser.statement]] | `CALLS` | [[parser.policy]] | Registered for `tokens.KeywordPolicy` at top level. |
| [[index.policy]] | `DEPENDS_ON` | [[ast]] | Consumes the produced `PolicyStatement`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseThePolicyStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `policy`, the name identifier, and `{`; loops parsing body statements until `}` or EOF; consumes an optional `;` and **discards** any trailing/line comments between statements; then consumes `}` and extends the span to it.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` if the keyword, name, `{`, or `}` is missing, if a body statement records an error, or if a body statement returns `nil`.

- **Signature:** `parsePolicyStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** The policy-scope dispatcher. Registered heads: comments, `title`, `description`, `version`, `tag`, `rule`, `fact`, `export`, `let`, `use`, `shape`, `derive`.
  - **Side Effects:** None directly; the delegate consumes tokens.
  - **Exceptions:** Records `unexpected token '<kind>'` and returns `nil` for an unregistered head — **without advancing**, so any caller must break on `p.err`.

- **Signature:** `(*Parser).registerPolicyStatementHandler(tokenType: tokens.Kind, fn: statementParser)`
  - **Behavior:** Inserts into the policy-scope table; called only from `registerParseFns`.
  - **Side Effects:** Mutates `policyStatementHandlers`.
  - **Exceptions:** None; duplicates overwrite silently.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless functions over parser state; policy bodies do not nest, so there is no recursion here.
- **Performance/Scale Notes:** One map probe per body statement. Negligible.
- **Dependencies Risk:**
  - **Comments inside a policy body are dropped.** The loop advances past `TrailingComment`/`LineComment` without building nodes, whereas [[parser.parse]] preserves them at top level. Comment-preserving tooling (formatters, doc generators) will see policy-internal comments vanish — except those attached to expressions by [[parser.expression]].
  - **Unterminated blocks report the wrong thing.** The loop exits on EOF as well as `}`, so a missing closing brace surfaces as `expected }, got EOF` positioned at end-of-file rather than at the policy header.
  - **`export` means something different here.** The policy table binds `export` to [[parser.export_rule]] (`export decision of …`), not the top-level shape/derive export. Same keyword, different production.
  - **Body statements are stored flat.** `PolicyStatement.Statements` mixes metadata, facts, rules, and declarations in source order with no grouping; consumers must filter by type.
