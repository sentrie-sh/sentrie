---
id: parser.parser
type: Class
language: Go
file_path: parser/parser.go
tags: parser-state, token-window, error-accumulation, dispatch-tables
---

# Node: parser.Parser (Parser State Machine)

## 1. Architectural Role & Intent
`parser/parser.go` defines the `Parser` struct — the mutable state machine every production operates on — together with its token-window navigation primitives (`advance`, `expect`, `peek`, `canExpect`) and its error accumulator. It exists to centralise the two-token lookahead discipline and the "report and continue" error strategy, so that individual productions can be written as small functions that assume a valid `current`/`next` window and never handle I/O or error plumbing themselves.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.parser` | `DEPENDS_ON` | [[lexer]] | Owns the `*lexer.Lexer`; `advance()` is the sole call site of `NextToken`. |
| `parser.parser` | `DEPENDS_ON` | [[tokens]] | The window holds `tokens.Instance`; all predicates compare `tokens.Kind`; `tokens.Err` builds sentinel error tokens. |
| `parser.parser` | `CALLS` | [[parser.lookups]] | `NewParser` calls `registerParseFns()` to populate all four handler tables before the window is primed. |
| [[parser.parse]] | `CALLS` | [[parser.parser]] | `ParseProgram` drives `hasTokens`, `canExpect`, and `advance`. |
| [[parser.expression]] | `CALLS` | [[parser.parser]] | Pratt loop reads `p.current`, dispatches through the handler maps, and reports `noPrefixParseFnError`. |
| [[parser.statement]] | `CALLS` | [[parser.parser]] | `parseStatement` reads `head()` and dispatches via `statementHandlers`. |
| [[loader]] | `CALLS` | [[parser.parser]] | Constructs one parser per policy file. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Parser` struct — `{ lexer *lexer.Lexer, reference string, current, next tokens.Instance, atEof bool, err error, prefixHandlers, infixHandlers, statementHandlers, policyStatementHandlers map[tokens.Kind]… }`
  - **Behavior:** All state is unexported. `reference` is the filename stamped onto `ast.Program.Reference`. The four maps are the entire dispatch mechanism; there is no switch-based fallback.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `NewParser(input: io.Reader, filename: string) -> *Parser` / `NewParserFromString(input, filename: string) -> *Parser`
  - **Behavior:** Builds the lexer, registers handlers, then calls `advance()` **twice** to fill `current` and `next`. The double-prime is why a production can always peek one token ahead without a nil check.
  - **Side Effects:** Consumes the first two tokens from the reader.
  - **Exceptions:** None.

- **Signature:** `(*Parser).head() -> tokens.Instance` / `peek() -> tokens.Instance`
  - **Behavior:** Read the window without consuming. `peek()` returns a synthetic bare `Instance{Kind: EOF}` — **with a zero-valued `Range`** — once `atEof` is set, so a span taken from a peeked EOF token is meaningless.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Parser).advance() -> tokens.Instance`
  - **Behavior:** Returns the consumed token and shifts the window. Encodes three special cases: advancing past EOF yields an error token rather than panicking; a `tokens.Error` in `current` is converted into an accumulated parse error and returned as-is **without advancing** (this is how lexical failures become parse failures); reaching `EOF` sets `atEof` and stops pulling from the lexer.
  - **Side Effects:** Mutates the window and `atEof`; may append to `p.err`.
  - **Exceptions:** Never panics; never returns a Go error.

- **Signature:** `(*Parser).expect(kind: tokens.Kind) -> bool` / `advanceExpected(kind) -> (tokens.Instance, bool)`
  - **Behavior:** Assert-and-consume. Both report `expected X, got Y at <range>` on mismatch. `expect` discards the token; `advanceExpected` returns it (or a synthetic error token) alongside the success flag.
  - **Side Effects:** Consumes a token on success; appends to `p.err` on failure.
  - **Exceptions:** Never returns an error — the boolean is the only failure signal, and **ignoring it silently continues parsing from an unexpected token**.

- **Signature:** `(*Parser).canExpect(kind) -> bool` / `canExpectAnyOf(kinds ...) -> bool` / `hasTokens() -> bool`
  - **Behavior:** Non-consuming predicates over `current`. `hasTokens()` is simply `!atEof` and is the top-level loop condition.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Parser).errorf(format string, args ...any)`
  - **Behavior:** The single error sink. Prefixes every message with `parsing error at <current range>:` and **joins** onto `p.err` via `errors.Join`, so all diagnostics from a run are retained rather than the first one winning.
  - **Side Effects:** Mutates `p.err`.
  - **Exceptions:** N/A.

- **Signature:** `(*Parser).registerPrefix(kind, fn)` / `registerInfix(kind, fn)` / `noPrefixParseFnError(t)`
  - **Behavior:** Table registration used only during construction, plus the canonical "unparseable token in expression position" diagnostic.
  - **Side Effects:** Mutate the handler maps / `p.err`.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Single-use, mutable, not goroutine-safe.** The window, EOF flag, error, and handler maps all live on the instance; there is no reset path and no snapshot/restore. The only backtracking available is `lexer.PushBack`, which rewinds the *lexer* but not `current`/`next` — the reason speculative parsing is confined to a couple of carefully written sites.
- **Performance/Scale Notes:** Every dispatch is a single map lookup; the four maps are rebuilt per `Parser`, so constructing a parser per file costs four map allocations plus ~60 inserts. Negligible for CLI use, worth noting if a server ever parses many small files in a tight loop.
- **Dependencies Risk:**
  - **Boolean-only failure signalling.** `expect`/`advanceExpected` return `bool`; callers that ignore it keep parsing from an unexpected token and typically produce a cascade of downstream errors whose first entry is the only meaningful one. When reading a joined parse error, trust the earliest message.
  - **`advance()` on an error token does not progress.** It reports and returns the same token, so a production that loops on `advance()` without checking `p.err` can spin. All loops must be bounded by `hasTokens()` *and* an error check.
  - **`errorf` uses the current range, not the failing construct's range.** Reported positions point at wherever the window happened to be, which can be one token past the real culprit.
  - **No error recovery or synchronisation.** There is no panic-mode resync to a statement boundary, so a single syntax error usually invalidates everything after it in the file.
