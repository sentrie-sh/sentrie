---
id: parser.comment
type: Function / Endpoint
language: Go
file_path: parser/comment.go
tags: comments, trivia, statement-handler
---

# Node: parser.parseCommentStatement (Comment Statement)

## 1. Architectural Role & Intent
Turns a `LineComment` or `TrailingComment` token into an `ast.CommentStatement`, with `Trailing` set for the latter. Sentrie keeps comments **in the tree** rather than discarding them as trivia, so that documentation tooling and formatters can recover them. This production handles the statement-position case; comments in expression position are wrapped instead by [[parser.expression]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.comment` | `CALLS` | [[ast]] | Emits `ast.NewCommentStatement(value, span)`. |
| `parser.comment` | `CALLS` | [[parser.parser]] | Uses `head()` and `advance()`. |
| [[parser.statement]] | `CALLS` | [[parser.comment]] | Registered for both `LineComment` and `TrailingComment` in the top-level table. |
| [[parser.policy]] | `CALLS` | [[parser.comment]] | Registered for the same kinds in the policy-scope table; policy and block bodies reach it through `parsePolicyStatement`. |
| [[parser.shape]] | `CALLS` | [[parser.comment]] | `parseComplexShape` collects shape-body comments into `Cmplx.Comments`. |
| [[parser.expression]] | `DEPENDS_ON` | [[parser.comment]] | Handles the expression-position case with `PrecedingCommentExpression`/`TrailingCommentExpression` wrappers instead. |
| `parser.comment` | `DEPENDS_ON` | [[lexer]] | The lexer decides leading vs trailing by whether non-whitespace preceded the `--` on the line; this production trusts that classification rather than re-deriving it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseCommentStatement(ctx, parser) -> ast.Statement`
  - **Behavior:** Consumes the comment token and wraps its already-trimmed text (the `--` marker and surrounding whitespace were stripped by [[lexer]]) in a `CommentStatement`.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None - it cannot fail, and does not verify the token kind.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **The span covers the comment token.** It runs from the `--` to `comment.Range.To`, just past the comment's last character (#236; it used to be zero width at the start). A comment runs to the end of its line, so in a CRLF file the span includes the `\r`, though the trimmed value does not.
  - **Every comment the parser accepts is kept, through three mechanisms.** Top-level, policy-body and block-body comments become statements ([[parser.parse]], [[parser.policy]], [[parser.block]]); shape-body comments go to `Cmplx.Comments` ([[parser.shape]]); expression comments become wrapper nodes ([[parser.expression]]). [[format]] relies on all three to reprint a file without losing a comment.
  - **Leading vs trailing classification happens in the lexer**, not here. Both kinds map to `CommentStatement`; a `TrailingComment` token sets its `Trailing` flag (`ast.NewTrailingCommentStatement`) so a printer can put the comment back at the end of the line it followed.
