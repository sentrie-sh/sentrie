---
id: lexer
type: System / Package
language: Go
file_path: lexer/
tags: tokenization, scanner, front-end, source-positions, heredoc
---

# Node: Lexer (Streaming Rune Scanner)

## 1. Architectural Role & Intent
`lexer` is the first stage of the Sentrie compilation front-end: a hand-written, single-pass rune scanner that turns a `.sentrie` byte stream into positioned `tokens.Instance` values. It exists to give [[parser]] a pull-based token feed with exact `file:line:column` provenance, so that every downstream diagnostic - parse errors, index validation failures, and runtime type errors - can point at the original source span. It is deliberately streaming (`bufio.Reader`-backed, no full token slice) and supports a push-back stack so the parser can perform speculative lookahead without a separate buffering layer.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `lexer` | `LAYERED_ON` | [[tokens]] | Emits `tokens.Instance` values; consults `tokens.IsKeyword` to promote identifiers to keyword kinds; builds `tokens.Range` from `tokens.Pos`. |
| `lexer` | `IMPORTS` | `std.bufio`, `std.bytes`, `std.io`, `std.regexp`, `std.slices`, `std.strings`, `std.unicode`, `std.unicode/utf8` | No third-party dependencies; the scanner is self-contained. |
| [[parser]] | `LAYERED_ON` | [[lexer]] | `parser.NewParser` wraps an `io.Reader` in a `Lexer` and drives it via `NextToken`. |
| [[parser]] | `CALLS` | [[lexer]] | Lambda-vs-grouped-expression disambiguation calls `PushBack` to rewind speculatively consumed tokens. |
| [[loader]] | `CALLS` | [[lexer]] | Indirect consumer: each discovered policy file is opened and handed to `parser.NewParser`, which constructs the lexer. |
| `lexer` | `DEPENDS_ON` | [[grammar]] | Conformance only - the token vocabulary implements the terminals declared in the reference grammar; there is no build-time link. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewLexer(reader: io.Reader, filename: string) -> *Lexer`
  - **Behavior:** Constructs a scanner and immediately reads the first rune so that `NextToken` can dispatch without a priming call. `filename` is carried into every emitted `tokens.Range` and is presentation-only - it is never opened.
  - **Side Effects:** Performs the first read against `reader`; compiles the identifier validation regex.
  - **Exceptions:** None. A read failure during priming is absorbed as EOF.

- **Signature:** `(*Lexer).NextToken() -> tokens.Instance`
  - **Behavior:** Returns the next token, popping the push-back stack first. Skips whitespace, then dispatches on the current rune: multi-character operators (`==`, `=>`, `!=`, `<=`, `>=`, `|>`, `...`, `<<<`), single-character operators and punctuation, `--` comments, `"`-delimited strings, `<<<TAG` heredocs, numbers (`Int` for digits only, `Float` when there is a `.digits` fraction, a `[eE][+-]?digits` exponent, or both; #180), and identifiers (promoted to keyword kinds via [[tokens]]). Returns `tokens.EofInstance` at end of input.
  - **Side Effects:** Advances the reader and mutates line/column/offset counters; consumes from the push-back stack.
  - **Exceptions:** **Never returns a Go `error`.** All lexical failures are encoded in-band as `tokens.Error` tokens carrying the message as their value - unterminated strings, malformed heredocs, invalid identifiers, unknown runes (`unexpected character "$"`), a bare `|` without `>`, and malformed number literals (`invalid number literal "1e+": the exponent needs at least one digit`, `invalid number literal "1.e5": a digit must follow the decimal point`). The caller must inspect token kind, not an error return.

- **Signature:** `(*Lexer).PushBack(t: tokens.Instance)`
  - **Behavior:** Returns a token to the stream as a **LIFO** stack; the next `NextToken` yields it. This is the parser's backtracking primitive.
  - **Side Effects:** Mutates the push-back stack.
  - **Exceptions:** None. The stack is unbounded and unvalidated - pushing tokens that were never emitted is accepted and will be replayed verbatim.

- **Signature:** `LexerError` struct - `{ Filename: string, Position: tokens.Pos }`, `Error() -> string`
  - **Behavior:** Positional error carrier wrapped by the constructors below. Renders as `at <file>:<line>:<column>` and is designed to be the `%w` tail of a descriptive message.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `UnterminatedStringError(filename: string, pos: tokens.Pos) -> error` / `InvalidHereDocSyntaxError(filename: string, pos: tokens.Pos) -> error`
  - **Behavior:** Construct the two lexical failure modes. Both are internal to the package's string/heredoc readers and are flattened into `tokens.Error` values before reaching the parser.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Highly stateful and single-use.** A `Lexer` owns reader position, line/column/offset counters, a current-line rune buffer used for comment classification, and the push-back stack. It cannot be reset, rewound beyond the push-back stack, or shared across goroutines; construct one per source file.
- **Performance/Scale Notes:** Streaming and allocation-light for operators and punctuation; identifiers, strings, comments, and heredocs allocate a builder per token. `peekAhead` calls `bufio.Reader.Peek(4)` per lookahead, so multi-character operator dispatch costs a peek per candidate. Policy files are small, so throughput is not a practical constraint.
- **Dependencies Risk:** No external failure domain - the risks are all semantic:
  - **In-band errors.** Because failures are `tokens.Error` tokens rather than returns, a consumer that ignores the error kind will silently parse a corrupt token stream.
  - **Silent truncation.** `readRune` treats *any* reader error (not just `io.EOF`) as end-of-input, so an I/O failure mid-file yields a valid-looking but truncated program instead of a hard error.
  - **Unicode identifiers.** The dispatch predicate is `unicode.IsLetter` but the validation regex is ASCII-only (`^[a-zA-Z_][a-zA-Z0-9_]*$`), so a non-ASCII identifier is scanned and then rejected as `invalid identifier` rather than never being recognised. Diagnose "invalid identifier" complaints here, not in [[parser]].
  - **Comments are tokens.** `LineComment` and `TrailingComment` are emitted into the stream (trailing-vs-leading is decided by whether non-whitespace preceded the `--` on the current line). Consumers must skip or attach them; they are not discarded by the scanner.
  - **Position model (#236).** Lines and columns are 1-based on every line; offsets are 0-based bytes; columns count runes, not bytes. A token's `From` is its first rune; its `To` is exclusive, just past its last rune **on that rune's line** (`endPosition`), never the position of the next rune. A newline belongs to the line it ends, so a token at the end of a line does not end at column 0 of the next line, and a token at EOF is not one column short. Statement parsers take a statement's end straight from its last token's `To`.
  - **Number literal shape (#180).** An `e` or `E` straight after a number's digits always starts an exponent, so `1e`, `1e+` and `1ex` are error tokens rather than an `Int` followed by an identifier, and `1.e5` is an error rather than field access `e5` on `1`. A `.` followed by any other non-digit still ends the number (`1.foo` lexes as `Int`, `.`, `Ident`). The token value keeps the source spelling (`2.5E+10`); [[parser.primary]] converts it with `strconv.ParseFloat`, so an out-of-range exponent (`1e400`) is a parse error.
  - **Heredoc strictness.** `<<<TAG` requires an immediate ASCII identifier tag, permits only whitespace after it on the opening line, and terminates on a line matching the tag **exactly** (no indentation). EOF before the terminator surfaces as an unterminated-string error. The newline after the terminator is left to `skipWhitespace`, so the heredoc token ends just past its tag.
