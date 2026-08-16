---
id: parser.primary
type: Module / File
language: Go
file_path: parser/primary.go
tags: literals, atoms, prefix-handlers, leaf-expressions
---

# Node: parser.primary (Atomic Prefix Handlers)

## 1. Architectural Role & Intent
`parser/primary.go` holds the six leaf prefix handlers - null, trinary, identifier, pipeline hole, integer, and string/float literals - that terminate every expression descent. It is where raw token text is converted into typed AST payloads: `strconv` for numerics, [[trinary]]'s keyword mapping for the tri-state literals. Everything else in the expression grammar eventually bottoms out here.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.primary` | `CALLS` | [[trinary]] | `parseTrinaryLiteral` calls `trinary.FromToken` to resolve `true`/`false`/`unknown`. |
| `parser.primary` | `CALLS` | [[ast]] | Emits `NullLiteral`, `TrinaryLiteral`, `Identifier`, `PipelineHoleExpression`, `IntegerLiteral`, `StringLiteral`, `FloatLiteral`. |
| `parser.primary` | `IMPORTS` | `std.strconv` | Parses integer (base 10, 64-bit) and float (64-bit) literals. |
| [[parser.lookups]] | `CALLS` | [[parser.primary]] | All six are registered as prefix handlers. |
| [[parser.literal]] | `CALLS` | [[parser.primary]] | The constraint-argument parser reuses these same handlers for its literal-only subset. |
| [[parser.call]] | `CALLS` | [[parser.primary]] | Reuses `parseIntegerLiteral` to read a memoization TTL suffix. |
| [[box.value]] | `DEPENDS_ON` | [[parser.primary]] | Literal payloads become boxed values at evaluation time. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseNullLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Consumes the token and emits `NullLiteral`. Defensively re-checks the kind even though the dispatch table guarantees it.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** `expected \`null\` literal, got %s at %s` on a kind mismatch, recorded through `errorf` like every other production.

- **Signature:** `parseTrinaryLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Maps the `true`/`false`/`unknown` keyword token through `trinary.FromToken` and emits `TrinaryLiteral`.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None - `FromToken` is total, so an unexpected token silently yields `Unknown`.

- **Signature:** `parseIdentifier(ctx, p) -> ast.Expression`
  - **Behavior:** Wraps the token value in `Identifier`. No resolution, no scope check - every name is unbound until [[index.resolve]] and [[runtime.eval_ident]] run.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None.

- **Signature:** `parsePipelineHoleExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Emits the `#` placeholder node consumed by [[parser.pipeline]].
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None - a `#` outside a pipeline parses successfully here and must be rejected downstream.

- **Signature:** `parseIntegerLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** `strconv.ParseInt(value, 10, 64)` → `IntegerLiteral`.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** `invalid integer literal %q at %s: %w` on overflow or malformed input.

- **Signature:** `parseFloatLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** `strconv.ParseFloat(value, 64)` → `FloatLiteral`.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** `invalid float literal %q at %s: %w`.

- **Signature:** `parseStringLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Wraps the already-unescaped token value (escape processing happened in [[lexer]]) in `StringLiteral`. Heredoc content arrives through this same path.
  - **Side Effects:** Consumes a token.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One `strconv` call per numeric literal; nothing else notable.
- **Dependencies Risk:**
  - **Regression history.** `parseNullLiteral` previously assigned `p.err` directly instead of calling `errorf`, which dropped earlier diagnostics (bypassing `errors.Join`) and omitted the standard position prefix. Fixed; every production in this package must report through `errorf` to preserve the error-accumulation contract described in [[parser.parser]].
  - **Integer/float distinction is lost downstream.** `IntegerLiteral` carries an `int64`, but [[box.value]] stores all numbers as `float64`, so integers beyond 2^53 silently lose precision at evaluation even though they parsed exactly.
  - **`FromToken` never fails.** A malformed trinary token yields `Unknown` rather than an error, so a lexer change that alters keyword kinds would degrade silently into "unknown" verdicts.
  - **Identifiers are unresolved strings.** Nothing here distinguishes a fact, a `let`, a derive, or a module alias - they are all `Identifier` nodes.
