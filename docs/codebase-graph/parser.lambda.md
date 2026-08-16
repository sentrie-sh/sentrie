---
id: parser.lambda
type: Function / Endpoint
language: Go
file_path: parser/lambda.go
tags: lookahead, backtracking, lambda, lexer-manipulation
---

# Node: parser.tryReadLambdaSignature (Speculative Lambda Probe)

## 1. Architectural Role & Intent
A self-contained lookahead routine that reads `( a, b ) =>` **directly from the lexer** - bypassing the parser entirely - and pushes every consumed token back if the pattern does not match. It exists because `(` is ambiguous between a grouped expression and a lambda parameter list, and the parser's two-token window is not enough to tell them apart when the parameter list is longer than one identifier. This is the only unbounded backtracking in the front-end.

This probe handles the untyped form only. Parameters that are typed or optional are deliberately rejected here and picked up by [[parser.typed_lambda]], so the two nodes partition the lambda grammar between them.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.lambda` | `CALLS` | [[lexer]] | Drives `NextToken` and `PushBack` directly; takes a `*lexer.Lexer`, not a `*Parser`. |
| `parser.lambda` | `DEPENDS_ON` | [[tokens]] | Matches on `Ident`, `PunctComma`, `PunctRightParentheses`, and `TokenFatArrow`. |
| [[parser.block]] | `CALLS` | [[parser.lambda]] | The sole caller: `parseGroupedExpression` probes before committing to either interpretation. |

## 3. Interface Contracts & Public Surface

- **Signature:** `tryReadLambdaSignature(lex: *lexer.Lexer) -> (params []string, ok bool)`
  - **Behavior:** Recognises exactly two shapes: `) =>` (empty parameter list, returning an empty non-nil slice) and `ident (, ident)* ) =>`. Every token read is buffered; on any mismatch the buffer is replayed into the lexer's push-back stack in reverse order, restoring the stream exactly. On success the tokens through the fat arrow are **consumed** and the caller must refill its window.
  - **Side Effects:** Advances the lexer on success; leaves it byte-identical on failure via push-back.
  - **Exceptions:** None - it never errors, only reports `ok == false`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless with respect to the parser, but it mutates lexer position and push-back state. It must be called only when the caller can tolerate either outcome.
- **Performance/Scale Notes:** Reads and potentially replays a number of tokens proportional to the parameter list (or to however far it gets before mismatching). The push-back stack is LIFO, so the reverse-order replay in `undo()` is what makes restoration correct - reversing that loop would silently scramble the token stream.
- **Dependencies Risk:**
  - **Deliberately narrow.** It recognises only untyped, non-optional parameters. `(a: string) => …` and `(a?) => …` fail the probe and fall through to [[parser.typed_lambda]] - by design, not by oversight.
  - **The undo path is correctness-critical and untested in isolation.** It is a package-private function with no direct test; its behaviour is only exercised indirectly through grouped-expression parsing.
  - **It bypasses the parser's error accumulation.** Taking a `*lexer.Lexer` rather than a `*Parser` means it cannot report diagnostics - every failure is silent by construction, which is correct for a probe but means a malformed lambda produces an error from a later production instead.
  - **`ok == true` with an empty slice is meaningful** (a zero-parameter lambda) and must not be conflated with `params == nil`.
