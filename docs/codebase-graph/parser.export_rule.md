---
id: parser.export_rule
type: Function / Endpoint
language: Go
file_path: parser/export_rule.go
tags: declaration, decision-export, attachment, policy-output
---

# Node: parser.parseRuleExportStatement (Decision Export)

## 1. Architectural Role & Intent
Parses `export decision of <rule> (attach <ident> as <expr>)*` - the declaration that promotes one rule's outcome to the policy's public decision and attaches arbitrary named metadata expressions to it. It is the policy's **output contract**: what the CLI prints, what the HTTP API returns, and what another namespace can import. The attachment clauses are how a decision carries evidence (reasons, offending values, remediation hints) alongside its [[trinary]] verdict.

The `export` keyword is context-sensitive: inside a policy block it reaches this handler, while at top level it dispatches to [[parser.export_shape]] instead. Reading either node alone gives half the picture.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.export_rule` | `CALLS` | [[parser.expression]] | Parses each attachment's `as <expr>` at `LOWEST`. |
| `parser.export_rule` | `CALLS` | [[parser.parser]] | Uses `head`, `advance`, `advanceExpected`, `expect`, `errorf`. |
| `parser.export_rule` | `CALLS` | [[ast]] | Emits `ast.NewRuleExportStatement` and `ast.NewAttachmentClause`. |
| [[parser.policy]] | `CALLS` | [[parser.export_rule]] | Registered for `tokens.KeywordExport` in the **policy-scope** table. |
| [[runtime.decision]] | `DEPENDS_ON` | [[parser.export_rule]] | Resolves the exported rule and evaluates attachments into the decision payload. |
| [[parser.import]] | `DEPENDS_ON` | [[parser.export_rule]] | Cross-namespace `import decision` consumes what this declares. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseRuleExportStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `export` (unconditionally - the table guarantees the head), then **rejects `export derive` outright** with a scope-specific message, then requires `decision`, `of`, and a rule identifier. Zero or more `attach` clauses follow, each extending the span.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for `policy-level derives cannot be exported; only namespace-level derives may use export derive`, or for a missing `decision`/`of`/rule identifier, or a failed attachment.

- **Signature:** `parseAttachmentClause(ctx: context.Context, p: *Parser) -> *ast.AttachmentClause`
  - **Behavior:** Parses `attach <ident> as <expr>`. The identifier is the attachment key; the expression is evaluated at decision time and becomes its value.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing identifier, a missing `as`, or a failed expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable; attachment expressions are evaluated at runtime, not here.
- **Dependencies Risk:**
  - **`export` is scope-overloaded.** Inside a policy it means "export decision"; at namespace level it means "export shape/derive" via [[parser.export_shape]]. Two productions, one keyword - the `export derive` rejection here exists precisely to make the wrong-scope case explain itself.
  - **The exported rule is a bare name.** Nothing checks that the rule exists in this policy, is unique, or is exported once; a policy could declare multiple `export decision of` statements and the parser would accept all of them. [[index.validate]] owns those rules.
  - **Attachment keys are unvalidated and unordered by the AST's consumers.** Duplicate `attach x as …` clauses are both retained in the slice - unlike shape fields, nothing deduplicates - so the resolution rule (first wins, last wins, or error) is decided downstream.
  - **Attachment expressions are arbitrary.** They may call functions or reference facts, so an export can fail at evaluation time for reasons unrelated to the rule's own logic.
