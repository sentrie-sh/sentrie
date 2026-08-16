---
id: parser.call
type: Function / Endpoint
language: Go
file_path: parser/call.go
tags: call-expression, memoization, arguments, infix-handler
---

# Node: parser.parseCallExpression (Call and Memoization Suffix)

## 1. Architectural Role & Intent
Parses `callee(arg, …)` as an infix operator on the already-parsed callee, and - distinctively - the trailing **memoization suffix** `f(x)!` / `f(x)!30`. Making memoization a syntactic property of the call site (rather than a property of the function) is a deliberate design choice: the policy author decides at each use whether a result should be cached and for how long, which is what [[runtime.eval_call]] keys its memo table on.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.call` | `CALLS` | [[parser.expression]] | `parseExpressionList` parses each argument at `LOWEST`. |
| `parser.call` | `CALLS` | [[parser.primary]] | Reuses `parseIntegerLiteral` to read the TTL after `!`. |
| `parser.call` | `CALLS` | [[ast]] | Emits `ast.NewCallExpression(callee, args, memoized, ttl, span)`. |
| `parser.call` | `IMPORTS` | `std.time` | The TTL integer is interpreted as **seconds** and stored as a `time.Duration`. |
| [[parser.lookups]] | `CALLS` | [[parser.call]] | Registered as the infix handler for `(` at `CALL` precedence. |
| [[parser.pipeline]] | `CALLS` | [[parser.call]] | Pipeline stages are typically call expressions with a `#` hole argument. |
| [[runtime.eval_call]] | `DEPENDS_ON` | [[parser.call]] | Dispatches builtins, JS module functions, derives, and lambdas; honours the memo flag and TTL. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseCallExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `(`, parses the argument list, consumes `)`, and builds a non-memoized call spanning from the callee's start to the closing paren. If the next token is `!`, sets `Memoized = true`; if an integer follows the `!`, it is read as a TTL in **seconds**. A bare `!` means memoize with no expiry.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `(` or `)`, a failed argument, or a malformed TTL literal.

- **Signature:** `parseExpressionList(ctx, parser, end: tokens.Kind) -> []ast.Expression`
  - **Behavior:** Generic comma-separated expression list, terminated by the supplied token kind. Returns an empty (non-nil) slice for `f()`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` - distinguishable from empty - when an element fails to parse.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Argument parsing dominates. The memoization flag itself has no parse-time cost but has significant runtime consequences.
- **Dependencies Risk:**
  - **The memo suffix span is computed but discarded.** The code updates the local `rnge` after reading `!`/TTL, but the `CallExpression` was already constructed from the pre-suffix range, so a memoized call's span **stops at the closing paren** and excludes `!30`. Error messages pointing at such a call will under-report its extent.
  - **`!` is context-overloaded.** It is the memoization marker here, logical negation as a prefix ([[parser.unary]]), and part of the retired fact/shape syntax that [[parser.fact]] and [[parser.shape]] reject explicitly. Same token, three meanings.
  - **TTL units are implicit.** The integer is multiplied by `time.Second` with no unit token, so `f(x)!30` is thirty seconds - easy to misread as milliseconds.
  - **Missing commas are silently tolerated.** `parseExpressionList` only consumes a comma when present and otherwise loops, so `f(a b)` parses as two arguments. Same laxity as [[parser.collection]].
  - **Unterminated argument lists** exit on `hasTokens()` and report `expected ), got EOF`.
