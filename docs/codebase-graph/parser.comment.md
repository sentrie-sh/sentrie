---
id: parser.comment
type: Function / Endpoint
language: Go
file_path: parser/comment.go
tags: comments, trivia, statement-handler
---

# Node: parser.parseCommentStatement (Comment Statement)

## 1. Architectural Role & Intent
Turns a `LineComment` or `TrailingComment` token into an `ast.CommentStatement`. Sentrie keeps comments **in the tree** rather than discarding them as trivia, so that documentation tooling and formatters can recover them. This production handles the statement-position case; comments in expression position are wrapped instead by [[parser.expression]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.comment` | `CALLS` | [[ast]] | Emits `ast.NewCommentStatement(value, span)`. |
| `parser.comment` | `CALLS` | [[parser.parser]] | Uses `head()` and `advance()`. |
| [[parser.statement]] | `CALLS` | [[parser.comment]] | Registered for both `LineComment` and `TrailingComment` in the top-level table. |
| [[parser.policy]] | `CALLS` | [[parser.comment]] | Registered for the same kinds in the policy-scope table — though the policy body loop discards comments before they reach it. |
| [[parser.expression]] | `DEPENDS_ON` | [[ast]] | Handles the expression-position case with `PrecedingCommentExpression`/`TrailingCommentExpression` wrappers instead. |
| [[lexer]] | `DEPENDS_ON` | [[tokens]] | Classifies leading vs trailing by whether non-whitespace preceded the `--` on the line. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseCommentStatement(ctx, parser) -> ast.Statement`
  - **Behavior:** Consumes the comment token and wraps its already-trimmed text (the `--` marker and surrounding whitespace were stripped by [[lexer]]) in a `CommentStatement`.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None — it cannot fail, and does not verify the token kind.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **The span is degenerate.** It is built from `head.Range.From` to `comment.Range.**From**` — both the same token's start position, since `head()` and the subsequent `advance()` return the same instance. Every `CommentStatement` therefore has a zero-width range that does not cover its own text.
  - **Comment retention is inconsistent across the tree.** Top-level comments become statements ([[parser.parse]]); policy-body comments are silently dropped ([[parser.policy]]); block-body comments are parsed as statements; expression comments become wrapper nodes. Any tool that needs all comments cannot rely on a single mechanism.
  - **Leading vs trailing classification happens in the lexer**, not here, and both kinds map to the same AST node — so the distinction is lost at statement level even though it is preserved in expression position.
