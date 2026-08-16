---
id: parser.left_curly
type: Function / Endpoint
language: Go
file_path: parser/left_curly.go
tags: disambiguation, lookahead, prefix-handler, ambiguity
---

# Node: parser.parseFromLeftCurly (Block vs Map Disambiguation)

## 1. Architectural Role & Intent
A three-line dispatcher that resolves Sentrie's one genuine syntactic ambiguity: a `{` may begin either a **map literal** or a **block expression**. It decides by peeking a single token - a string, `[`, or an immediate `}` means map; anything else means block. This is why `{ "a": 1 }` and `{ let x = 1 yield x }` can share an opening token without a grammar-level conflict.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.left_curly` | `CALLS` | [[parser.collection]] | Delegates to `parseMapLiteral` when the lookahead indicates a map. |
| `parser.left_curly` | `CALLS` | [[parser.block]] | Delegates to `parseBlockExpression` otherwise. |
| `parser.left_curly` | `CALLS` | [[parser.parser]] | Uses `peek()` - the sole reason the parser maintains a two-token window. |
| [[parser.lookups]] | `CALLS` | [[parser.left_curly]] | Registered as the prefix handler for `PunctLeftCurly`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseFromLeftCurly(ctx, p) -> ast.Expression`
  - **Behavior:** Peeks one token past the `{`. `String`, `PunctLeftBracket` (a computed map key), or `PunctRightCurly` (an empty map) route to the map literal; every other token routes to the block expression. Consumes nothing itself - the delegate consumes the brace.
  - **Side Effects:** None directly.
  - **Exceptions:** None; failures come from the delegate.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One token peek. Trivial.
- **Dependencies Risk:**
  - **`{}` is always an empty map, never an empty block.** That is the right default (an empty block has no `yield` and would be invalid anyway), but it means the ambiguity is resolved by fiat rather than by intent.
  - **One token of lookahead is the entire decision.** A future map-key form that does not start with a string or `[` - bare identifier keys, say - would be silently parsed as a block and fail with a confusing "expected yield" error. Any change to map-key syntax must be mirrored here.
  - **The delegates are not symmetric about brace consumption.** `parseMapLiteral` consumes `{` with a bare `advance()` (trusting this dispatcher), while `parseBlockExpression` uses `advanceExpected`. Calling `parseMapLiteral` from anywhere else would silently swallow a non-brace token.
