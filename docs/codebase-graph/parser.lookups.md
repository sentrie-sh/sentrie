---
id: parser.lookups
type: Module / File
language: Go
file_path: parser/lookups.go
tags: dispatch-table, registration, pratt-parser, language-surface
---

# Node: parser.lookups (Handler Registration Tables)

## 1. Architectural Role & Intent
`parser/lookups.go` is the **single declarative map of Sentrie's entire syntax surface**: it registers every prefix handler, infix handler, top-level statement handler, and policy-scoped statement handler against the token kind that triggers it, and declares the three parser function types plus the primitive/aggregate type-keyword sets. Any question of the form "is this token legal here, and which function parses it?" is answered by reading this one file - which is why it is the highest-value navigation node in the parser package.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.lookups` | `CALLS` | [[parser.parser]] | Invoked by `NewParser`; calls `registerPrefix`/`registerInfix` to populate the maps. |
| `parser.lookups` | `DEPENDS_ON` | [[tokens]] | Every table key is a `tokens.Kind`; `PRIMITIVE_TYPES`/`AGGREGATE_TYPES` are kind sets. |
| `parser.lookups` | `DEPENDS_ON` | [[ast]] | The handler function types are defined in terms of `ast.Expression` / `ast.Statement`. |
| `parser.lookups` | `DEPENDS_ON` | [[parser.precedence]] | `infixParser` takes a `Precedence` argument, binding the tables to the precedence ladder. |
| [[parser.expression]] | `READS_FROM` | [[parser.lookups]] | The Pratt loop looks up `prefixHandlers` and `infixHandlers` per token. |
| [[parser.statement]] | `READS_FROM` | [[parser.lookups]] | `parseStatement` looks up `statementHandlers`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Parser).registerParseFns()`
  - **Behavior:** Allocates and fills all four maps. Called exactly once, from `NewParser`, before the token window is primed.
  - **Side Effects:** Replaces the handler maps on the receiver.
  - **Exceptions:** None. Duplicate registrations silently overwrite - the last registration for a kind wins.

- **Signature:** `prefixParser` - `func(ctx, *Parser) ast.Expression`
  - **Behavior:** Parses an expression starting at `current`. Registered for literals (`true`/`false`/`unknown` → trinary, `null`, string, int, float), `Ident`, the pipeline hole `#`, the unary operators `!`/`-`/`+`/`not`, the `transform` keyword, `(` (grouped **or** lambda), `[` (list literal), and `{` (dispatched by lookahead in `parseFromLeftCurly` - map literal vs block).
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on failure after calling `p.errorf`.

- **Signature:** `infixParser` - `func(ctx, *Parser, left ast.Expression, precedence Precedence) ast.Expression`
  - **Behavior:** Extends an already-parsed left operand. Registered for the boolean keywords (`and`, `or`, `xor`), the membership/match keywords (`in`, `matches`, `contains`), `is` (defined/empty), `as` (cast), the arithmetic and comparison operators, `?` (ternary), `[` (index access), `.` (field access), `(` (call), and `|>` (pipeline).
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on failure.

- **Signature:** `statementParser` - `func(ctx, *Parser) ast.Statement`
  - **Behavior:** Parses a declaration. **Two separate tables use this type:**
    - `statementHandlers` (top level): `namespace`, comments, `policy`, `shape`, `derive`, `export`.
    - `policyStatementHandlers` (inside `policy { }`): comments, `title`, `description`, `version`, `tag`, `rule`, `fact`, `export`, `let`, `use`, `shape`, `derive`.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on failure.

- **Signature:** `PRIMITIVE_TYPES []tokens.Kind` - `string`, `number`, `boolean`, `trinary`, `document`
  - **Behavior:** Token kinds that begin a scalar type reference. Note `boolean` appears here even though the value model's canonical logic type is [[trinary]].
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `AGGREGATE_TYPES []tokens.Kind` - `list`, `dict`, `record`
  - **Behavior:** Token kinds that begin a parameterised type reference.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** The maps are per-`Parser` instance state, rebuilt on every construction; the two exported slices are package-level.
- **Performance/Scale Notes:** Roughly 60 map inserts and four allocations per parser construction - irrelevant for a per-file parser, but it is pure repeated work that could be hoisted to package level if parser churn ever mattered.
- **Dependencies Risk:**
  - **`not` is registered in both tables.** As a prefix it is unary negation (`not true`); as an infix it forms the negated-membership form (`x not in [...]`). A change to either handler affects both syntactic roles.
  - **`export` is registered in both statement tables with *different* handlers** - `parseExportStatement` at top level (shape/derive exports) versus `parseRuleExportStatement` inside a policy (`export decision of …`). Reading only one table gives a wrong answer about what `export` does.
  - **`shape` and `derive` are legal at both scopes** and share handlers, so scope-specific rules must live inside those productions rather than in the tables.
  - **`{` and `(` are ambiguous heads** resolved by lookahead inside their handlers (block vs map literal; grouped expression vs lambda). The tables alone do not tell you which node type results.
  - **Silent overwrite.** Registering the same kind twice in one table is not detected, so an accidental duplicate quietly disables the earlier handler.
  - **Exported mutable slices.** `PRIMITIVE_TYPES` and `AGGREGATE_TYPES` are exported `var`s, so any package could mutate them; treat them as constants.
