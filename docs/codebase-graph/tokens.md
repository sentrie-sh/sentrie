---
id: tokens
type: System / Package
language: Go
file_path: tokens/
tags: lexing, source-positions, language-frontend, diagnostics
---

# Node: Tokens (Lexical Vocabulary & Source Positions)

## 1. Architectural Role & Intent
`tokens` defines the closed lexical vocabulary of the Sentrie policy language (keyword set, operators, punctuation, literal classes) together with the source-position model (`Pos`, `Range`) used to anchor every downstream diagnostic. It is the root leaf of the dependency graph: it imports nothing from the project, so every other language-frontend package can depend on it without creating cycles. Its `Range` type is the universal "where did this come from" carrier threaded from the lexer all the way into runtime error messages.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `tokens` | `IMPORTS` | `std.fmt`, `std.slices` | No internal project dependencies; this is a graph sink. |
| `tokens` | `IMPORTS` | `ext.golang.x-exp` | Uses `golang.org/x/exp/maps` to enumerate the keyword lookup table. |
| [[lexer]] | `LAYERED_ON` | [[tokens]] | Lexer emits `tokens.Instance` values and constructs `tokens.Range` from byte offsets. |
| [[parser]] | `LAYERED_ON` | [[tokens]] | Parser dispatches on `tokens.Kind` for prefix/infix lookup and precedence. |
| [[ast]] | `LAYERED_ON` | [[tokens]] | Every AST node embeds a `tokens.Range` for span reporting. |
| [[trinary]] | `LAYERED_ON` | [[tokens]] | `trinary.FromToken` converts `true`/`false`/`unknown` keyword tokens into tri-state values. |
| [[xerr]] | `LAYERED_ON` | [[tokens]] | Span-anchored error constructors accept `tokens.Range` to render `file:line:col`. |
| [[index.package]] | `LAYERED_ON` | [[tokens]] | Static validation errors are anchored to declaration spans. |
| [[runtime]] | `LAYERED_ON` | [[tokens]] | Runtime errors and builtin arity failures carry originating spans. |

## 3. Interface Contracts & Public Surface

- **Signature:** `New(kind: Kind, value: string, r: Range) -> Instance`
  - **Behavior:** Constructs a lexical token carrying its kind, raw text, and source span.
  - **Side Effects:** None (pure value construction).
  - **Exceptions:** None.

- **Signature:** `EofInstance(file: string, pos: Pos) -> Instance`
  - **Behavior:** Produces the terminal `EOF` sentinel token that drives parser loop termination.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Err(r: Range, message: string) -> Instance`
  - **Behavior:** Produces an `Error`-kind token so lexical failures travel in-band through the token stream instead of aborting.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `IsKeyword(str: string) -> (Kind, bool)`
  - **Behavior:** Resolves an identifier candidate against the reserved-word table; the boolean discriminates keyword from user identifier.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Instance.IsOfKind(kinds: ...Kind) -> bool`
  - **Behavior:** Variadic membership test used pervasively by the parser for lookahead assertions.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewRange(file: string, from: Pos, to: Pos) -> Range` / `NewRangeFromPos(file: string, pos: Pos) -> Range` / `BadRange(file: string) -> Range`
  - **Behavior:** Span constructors; `BadRange` yields the sentinel used for synthesized/derived nodes that have no real source text.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Range.String() -> string`
  - **Behavior:** Renders `file:line:col-line:col`, collapsing to `file:line:col-col` for single-line spans. This is the canonical diagnostic prefix format.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Pos.IsBadPos() -> bool`
  - **Behavior:** Detects the `BadPos` sentinel (all fields `-1`), distinguishing "no known location" from "start of file".
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Fully stateless. All types are immutable value types passed by copy; `Instance`, `Pos`, and `Range` have no pointer identity semantics.
- **Performance/Scale Notes:** `Kind` is a `string` type rather than an integer enum, so comparisons are string comparisons rather than integer compares — this is a hot path in `IsOfKind` during parsing. The keyword table is a package-level map; lookups are O(1) but not lock-free-optimized because it is read-only after init.
- **Dependencies Risk:** Zero inbound failure risk (no I/O, no external state). However, this node is a **change-amplification hotspot**: adding a `Kind` constant forces coordinated updates in [[lexer]] (recognition), [[parser]] (prefix/infix registration and precedence), and often [[ast]] (a new node type). `Pos.Column` counts display characters while `Pos.Offset` counts bytes — conflating the two produces incorrect spans for non-ASCII source.
