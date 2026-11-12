---
id: format
type: System / Package
language: Go
file_path: format/
tags: formatter, printer, pretty-printer, comments, front-end, tooling
---

# Node: format (Canonical Source Printer)

## 1. Architectural Role & Intent
`format` prints a parsed program back to `.sentrie` source in one canonical style, the engine for a `sentrie fmt` command (#172). It prints from the AST, so it depends on the parser keeping every comment ([[parser.comment]]) and the source form of pipelines (`CallExpression.Piped`). The style rules are listed in the package doc comment in `format/format.go`. Because the AST drops some spellings (redundant parentheses, semicolons, a fact alias equal to its name), output is canonical rather than byte-preserving; `Source` guards against printer bugs by re-parsing its own output and comparing it with the input program.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `format` | `CALLS` | [[parser]] | `NewParser(...).ParseProgram` for the input and again for the output; precedence decisions use the exported `parser.Precedence` constants. |
| `format` | `LAYERED_ON` | [[ast]] | Type-switches over every statement, expression and type-ref node; `Equal` walks nodes by reflection. |
| `format` | `DEPENDS_ON` | [[tokens]] | `Equal` skips `tokens.Range` fields; blank-line detection reads span lines and byte offsets. |
| `format` | `IMPORTS` | `std.reflect` | Generic AST comparison and the comment/heredoc search that forces line breaks. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Source(ctx, filename string, src []byte) -> ([]byte, error)`
  - **Behavior:** Parses `src`, prints it with `Program`, re-parses the output and requires `Equal` to hold. An empty or whitespace-only file formats to empty output.
  - **Side Effects:** None.
  - **Exceptions:** The parse error for invalid input; `formatted output does not parse` or `formatted output changes the program` for a printer bug; `cannot print ...` for a node the printer does not know.

- **Signature:** `Program(prg *ast.Program, src []byte) -> ([]byte, error)`
  - **Behavior:** Prints `prg`. `src` is the text it came from and is read for blank lines between statements, heredoc tags and the exact text of comments; positions decide whether a list, call, pipeline, binary expression or clause list stays on one line.
  - **Side Effects:** None.
  - **Exceptions:** `cannot print statement|expression|type|operator ...` for anything unknown - it never drops a node silently.
  - **Types:** a union prints its members joined by ` | `, each with its own `?` and constraints (the union carries neither); the top type prints as `any`.

- **Signature:** `Equal(a, b *ast.Program) -> error`
  - **Behavior:** Field-by-field reflective comparison ignoring every `tokens.Range` and the file name; comments count, including their order among shape fields; a nil slice equals an empty one. Returns an error naming the path of the first difference.
  - **Side Effects:** None.
  - **Exceptions:** None beyond the difference report.

## 4. Operational Context & Gotchas
- **Statefulness:** A `printer` is built per call and is not shared.
- **Performance/Scale Notes:** `Source` parses twice and prints once; `safeAfterComment` renders a list item a second time when the item before it ends in a trailing comment.
- **Dependencies Risk:**
  - **Precedence is mirrored, not shared.** `opPrec` maps operator text to `parser.Precedence`, and the operand rules follow parser quirks: the right side of `/` binds at index precedence (`ns/fn(x)` calls `ns/fn`), the ternary groups to the left, Elvis reads its right side at comparison precedence, and `(a ?` starts a typed lambda, so a group opening with a ternary on a bare name prints `((a) ? ...)`. A change to [[parser.precedence]] or [[parser.infix]] must be mirrored here; the self-check in `Source` turns a miss into an error rather than a changed policy.
  - **Comment placement follows the parser's attachment.** A comment after a comma (or an opening bracket) attaches to the next element as a preceding comment; `seq` prints it back after the comma on the same line and the element on the next. In a broken list an item ending in a trailing comment or a heredoc drops its comma, and keeps it at the start of the next line when the next item starts with `-`, `+`, `!`, `(`, `[` or `not`. Both rely on [[parser.collection]] and [[parser.call]] treating commas as optional.
  - **Comment text comes from the source.** The lexer trims comment content, so `commentText` reads the bytes at the comment's span to keep inner indentation; `Equal` compares the trimmed content and cannot see that whitespace.
  - **Shape member order is compared separately.** Fields are a map and comments a slice, so `Equal` merges both by offset (`shapeOrder`) to catch a comment that moved past a field or the `...` open-shape marker (#145), which the printer writes in its source position after the fields.
  - **Layout follows the source.** Line breaks are copied from positions, so formatting is idempotent but not layout-independent.
