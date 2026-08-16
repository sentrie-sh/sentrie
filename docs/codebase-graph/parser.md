---
id: parser
type: System / Package
language: Go
file_path: parser/
tags: pratt-parser, recursive-descent, front-end, syntax-analysis
---

# Node: Parser (Pratt / Recursive-Descent Front-End)

## 1. Architectural Role & Intent
`parser` is the hand-written front-end that turns the token stream from [[lexer]] into an `ast.Program`. It is a **Pratt (top-down operator-precedence) parser** for expressions layered over table-driven recursive descent for statements: four handler maps (prefix, infix, top-level statement, policy-scoped statement) are registered once at construction and dispatched on token kind. It exists to enforce Sentrie's structural rules that a grammar alone cannot express — namespace-first files, policy-only metadata, arity-checked type constraints — and to produce nodes with exact source spans so every later diagnostic can point back at source.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser` | `LAYERED_ON` | [[lexer]] | Owns a `*lexer.Lexer` and pulls tokens via `NextToken`; calls `PushBack` for speculative lookahead in lambda-vs-grouped disambiguation. |
| `parser` | `LAYERED_ON` | [[tokens]] | Dispatches on `tokens.Kind`; all handler maps are keyed by kind. |
| `parser` | `CALLS` | [[ast]] | Every production emits AST nodes through `ast.New*` constructors; this package's only output. |
| `parser` | `CALLS` | [[ast.typeref]] | `parseTypeRef` calls `AddConstraint`, so constraint-name and arity errors surface as parse errors. |
| `parser` | `LAYERED_ON` | [[trinary]] | Trinary keyword literals are resolved to `trinary.Value` at parse time. |
| `parser` | `DEPENDS_ON` | [[grammar]] | Conformance only — implements the reference productions with no generated linkage. |
| [[loader]] | `CALLS` | [[parser]] | `LoadPrograms` constructs one parser per `.sentrie` file and calls `ParseProgram`. |
| [[index.package]] | `LAYERED_ON` | [[parser]] | Consumes this package's AST output rather than calling the parser directly. |

## 3. Interface Contracts & Public Surface

The package exposes a deliberately narrow surface: everything except construction, `ParseProgram`, and two type/constant sets is unexported.

- **Signature:** `NewParser(input: io.Reader, filename: string) -> *Parser`
  - **Behavior:** Constructs the lexer, registers all four handler tables, and primes the two-token lookahead window (`current`, `next`) with a double `advance()`.
  - **Side Effects:** Reads the first tokens from `input`.
  - **Exceptions:** None; lexical failures surface later as parse errors.

- **Signature:** `NewParserFromString(input: string, filename: string) -> *Parser`
  - **Behavior:** Convenience wrapper over a `strings.Reader`, used heavily by tests.
  - **Side Effects:** As above.
  - **Exceptions:** None.

- **Signature:** `(*Parser).ParseProgram(ctx: context.Context) -> (*ast.Program, error)`
  - **Behavior:** The single entrypoint. Enforces the namespace-first rule and drives statement parsing to EOF. See [[parser.parse]] for the full contract.
  - **Side Effects:** Consumes the entire token stream; accumulates errors on the parser.
  - **Exceptions:** Returns the joined error set; `(nil, nil)` for an empty file.

- **Signature:** `Precedence` type and its level constants (`LOWEST` … `PRIMARY`)
  - **Behavior:** The exported precedence ladder consumed by infix handlers. See [[parser.precedence]].
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `PRIMITIVE_TYPES`, `AGGREGATE_TYPES` (exported `[]tokens.Kind` vars)
  - **Behavior:** Token-kind sets used to recognise type-reference heads. Exported package-level **mutable slices** — treat as read-only constants.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `ErrParse` sentinel
  - **Behavior:** Declared in [[parser.err]] but not returned by any production — parse failures are `fmt.Errorf` values joined together. Do not match on it.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Single-use and stateful.** A `Parser` owns the lexer, a two-token window, an `atEof` flag, and an accumulated `err`. It cannot be reset or reused, and is not safe for concurrent use. One instance per source file.
- **Performance/Scale Notes:** Single pass with a two-token window and O(1) map dispatch per token; parse cost is linear in file size. `parseExpression` emits `slog.DebugContext` on entry and exit of **every** expression, so debug-level logging makes parsing dramatically noisier and slower. [[loader]] parses files sequentially, so whole-pack parse time is the sum of file times.
- **Dependencies Risk:** No external failure domain. The recurring hazards across this package:
  - **Errors accumulate but productions keep going.** Handlers call `p.errorf` and frequently return `nil` nodes while parsing continues, so a partially-built tree with nil children exists until the top-level loop checks `p.err`. Never consume an AST from a parser whose error is non-nil.
  - **Two statement namespaces.** `statementHandlers` (top level) and `policyStatementHandlers` (inside `policy { }`) are separate tables; a keyword valid in one is a syntax error in the other. `title`/`description`/`version`/`tag` are explicitly rejected at top level with a targeted message.
  - **Comments are woven into the tree.** They arrive as tokens and are re-emitted as `CommentStatement`, `PrecedingCommentExpression`, or `TrailingCommentExpression` wrappers. Any consumer matching on expression shape must unwrap them first.
  - **Lexer errors are in-band.** `advance()` converts a `tokens.Error` token into a parse error, so lexical and syntactic failures are indistinguishable in the returned error.
