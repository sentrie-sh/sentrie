---
id: parser.expression
type: Function / Endpoint
language: Go
file_path: parser/expression.go
tags: pratt-parser, expression-parsing, comment-attachment, core-loop
---

# Node: parser.parseExpression (Pratt Core Loop)

## 1. Architectural Role & Intent
`parser/expression.go` contains `parseExpression`, the Pratt algorithm at the heart of the front-end: parse a prefix, then repeatedly extend it with infix handlers while the next operator binds tighter than the caller's precedence. Every expression in every Sentrie construct — rule bodies, fact defaults, `let` initialisers, constraint arguments, lambda bodies — flows through this one function. Its secondary responsibility is **comment attachment**: leading comments are wrapped around the finished expression and trailing comments are wrapped after each sub-expression, so formatting information survives into the AST.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.expression` | `READS_FROM` | [[parser.lookups]] | Dispatches through `prefixHandlers` and `infixHandlers`. |
| `parser.expression` | `READS_FROM` | [[parser.precedence]] | Loop condition compares `precedences[current.Kind]` against the caller's binding power. |
| `parser.expression` | `CALLS` | [[parser.parser]] | Uses `advance`, `head`, `canExpectAnyOf`, and `noPrefixParseFnError`. |
| `parser.expression` | `CALLS` | [[ast]] | Wraps results in `ast.NewPrecedingCommentExpression` / `ast.NewTrailingCommentExpression`. |
| `parser.expression` | `IMPORTS` | `std.log/slog`, `std.slices` | Emits debug spans around every expression parse; reverses the collected comment list. |
| [[parser.parse]] | `CALLS` | [[parser.expression]] | Indirectly, through every statement production that parses an expression. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Parser).parseExpression(ctx: context.Context, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Four steps.
    1. Drains any leading `LineComment`/`TrailingComment` tokens into a local buffer.
    2. Looks up a prefix handler for `current`; absence is a hard failure.
    3. Runs the Pratt loop: while `precedences[current.Kind] > precedence`, dispatch the infix handler, passing the operator's own precedence as the new binding power. A token with a precedence entry but **no** registered infix handler breaks the loop cleanly rather than erroring.
    4. Wraps the result in `PrecedingCommentExpression` layers (buffer reversed first, so the innermost wrapper is the comment nearest the expression).

    Every sub-result — prefix and each infix extension — passes through `wrapWithTrailingComment`, so a trailing comment can attach at any depth of the chain.
  - **Side Effects:** Consumes tokens; may record parse errors; emits two `slog.DebugContext` entries per invocation (entry and deferred exit).
  - **Exceptions:** Returns `nil` after `noPrefixParseFnError` (`no prefix parse function found for '<value>' at <range>`) when `current` cannot start an expression. Infix handlers may also return `nil`; the loop does not check, so a nil can propagate as `leftExp` into a subsequent handler.

- **Signature:** `wrapWithTrailingComment(expr: ast.Expression, parser: *Parser) -> ast.Expression`
  - **Behavior:** If the head token is a `TrailingComment`, consumes it and wraps `expr`; otherwise returns `expr` unchanged. Nil-safe: a nil expression passes straight through so a failed parse is not masked by a wrapper.
  - **Side Effects:** May consume a token.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Reentrant with respect to the `Parser` — it recurses through handlers that call it again with a higher precedence — but shares the single mutable token window, so it is neither concurrent nor backtrackable.
- **Performance/Scale Notes:** Two `slog.DebugContext` calls per expression node, including the deferred exit log, make debug-level logging extremely chatty and measurably slow on large policies; the arguments (`p.current`, precedence) are evaluated even when the level is disabled because they are passed as variadic values. Otherwise the loop is O(tokens) with a map probe per iteration.
- **Dependencies Risk:**
  - **Nil propagation.** The loop does not verify that `prefix(...)` or an infix handler returned a non-nil expression before using it as `leftExp`. A failed sub-parse therefore yields a partially-nil tree, and the resulting error message often names a construct several tokens past the real problem. Always check `p.err` before touching the returned expression.
  - **Comment wrappers change the node type.** An expression that "is" a call may actually be a `TrailingCommentExpression` wrapping one. Consumers in [[index]] and [[runtime]] must unwrap comment nodes before type-matching, and anything comparing AST shape for equality must normalise them away.
  - **Comment ordering is deliberate.** The leading-comment buffer is reversed before wrapping so that source order is preserved through the nesting; changing the reversal silently inverts comment attribution.
  - **Precedence without a handler is a silent no-op.** A token carrying a precedence entry but no infix registration simply terminates the expression, producing a "unexpected token" error from the *caller* rather than a useful message here.
