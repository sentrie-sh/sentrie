---
id: parser.block
type: Module / File
language: Go
file_path: parser/block.go
tags: block-expression, yield, lambda, speculative-parsing, backtracking
---

# Node: parser.block (Block Expressions and Grouped/Lambda Disambiguation)

## 1. Architectural Role & Intent
`parser/block.go` owns two constructs that look unrelated but share a design problem. `parseBlockExpression` parses `{ let … yield expr }` — Sentrie's expression-with-locals form, where `yield` is **mandatory** because a block is an expression that must produce a value. `parseGroupedExpression` handles `(` and performs the package's only true **backtracking**: it speculatively reads ahead through the lexer's push-back stack to decide whether `(` opens a parenthesised expression or a lambda parameter list.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.block` | `CALLS` | [[parser.lambda]] | `tryReadLambdaSignature` performs the untyped-lambda lookahead over the raw lexer. |
| `parser.block` | `CALLS` | [[parser.typed_lambda]] | Falls back to the typed-parameter lambda path when the fast probe fails. |
| `parser.block` | `CALLS` | [[parser.policy]] | Block statements are parsed with `parsePolicyStatement` — the policy-scope table. |
| `parser.block` | `CALLS` | [[parser.expression]] | Parses the `yield` expression and grouped expressions at `LOWEST`. |
| `parser.block` | `CALLS` | [[lexer]] | Directly manipulates `PushBack` and `NextToken`, bypassing the parser's own window. |
| `parser.block` | `CALLS` | [[ast]] | Emits `BlockExpression` and `LambdaExpression`. |
| [[parser.lookups]] | `CALLS` | [[parser.block]] | `parseGroupedExpression` is the prefix handler for `(`. |
| [[parser.rule]] | `CALLS` | [[parser.block]] | Rule bodies are normally block expressions. |
| [[parser.derive]] | `DEPENDS_ON` | [[parser.block]] | Derives require a lambda, which requires a block body. |
| [[runtime.eval_block]] | `DEPENDS_ON` | [[ast]] | Evaluates statements in order then returns the yielded value. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseBlockExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Consumes `{`, then parses statements **while the head is `let` or a line comment**, then requires the `yield` keyword, one expression, and `}`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `{`, a failed statement, a missing `yield`, a failed yield expression, or a missing `}`.

- **Signature:** `parseGroupedExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Three-stage disambiguation.
    1. Consumes `(`, pushes both window tokens back into the lexer, and calls `tryReadLambdaSignature` to probe for `( a, b ) =>`. On success it checks for duplicate parameter names, refills the window, and parses the body as a block — yielding an untyped `LambdaExpression`.
    2. Otherwise it refills the window and, if the head looks like a typed parameter list (`)`, or `ident :`, or `ident ?`), delegates to [[parser.typed_lambda]].
    3. Otherwise it parses a plain expression and requires `)`, returning the inner expression **unwrapped** — parentheses leave no trace in the AST.
  - **Side Effects:** Consumes tokens and **mutates the lexer's push-back stack directly**.
  - **Exceptions:** `duplicate lambda parameter %q`; `lambda body must be a block expression { ... yield ... }`; `nil` on a missing `)` or a failed inner expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless functions, but `parseGroupedExpression` reaches around the parser's abstraction and reassigns `p.current`/`p.next` from the lexer directly. It is the only production that does so, and the reason [[lexer]] exposes `PushBack` at all.
- **Performance/Scale Notes:** The lambda probe can read an arbitrary number of tokens before backtracking, so a long parenthesised expression is scanned twice. Bounded by parameter-list length in practice.
- **Dependencies Risk:**
  - **Blocks accept only `let` and line comments as statements.** The loop condition is a two-kind check, so a `fact`, `use`, or trailing comment inside a block does not error as "not allowed here" — it simply ends the statement loop and then fails with `expected yield`. That error message is the single most misleading diagnostic in the parser.
  - **`yield` is mandatory and terminal.** There is no implicit last-expression return, and nothing may follow the yield expression before `}`.
  - **Block spans use `From` for both ends.** The range is built from `lCurly.Range.From` and `rCurly.Range.**From**`, so a block's span stops at the *start* of its closing brace.
  - **Parentheses are erased.** Grouping is not represented in the AST, so a formatter cannot reproduce the author's parenthesisation and any AST-equality check treats `(a + b)` and `a + b` as identical.
  - **The two lambda paths produce different nodes.** The fast path yields `NewLambdaExpression` (untyped params, no return type); the typed path yields `NewLambdaExpressionFull`. Consumers must handle nil `ParamTypes`/`ParamOpts` slices from the fast path — see [[ast]].
