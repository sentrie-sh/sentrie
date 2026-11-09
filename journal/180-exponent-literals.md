# 180-exponent-literals

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #180
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Number literals with an exponent (`1e5`, `2.5e-3`) fail to parse; extend lexer number scanning to accept `[eE][+-]?digits` and produce a float literal.
- No design choice left open by the issue.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:34 - Lex and parse exponent number literals
- `lexer.readNumber` now takes an optional `[eE][+-]?digits` exponent after the integer or fraction part; any exponent makes the token a `Float`, and the value keeps the source spelling for `strconv.ParseFloat`.
- Decision: an `e`/`E` straight after digits always starts an exponent, so `1e`, `1e+`, `1ex` are error tokens (`the exponent needs at least one digit`) rather than `Int` plus identifier. Matches the issue's "invalid forms" list and the brief's ban on silent splitting.
- Decision: `1.e5` is rejected (`a digit must follow the decimal point`) rather than accepted as Go and JavaScript do. The existing fraction rule already needs a digit after `.`, and accepting it would make `1.e5` differ from `1.foo` (field access) only by the letter. Only `.` followed by `e`/`E` is rejected; `1.foo` lexes as before.
- Problem: the malformed-literal error surfaced as `no prefix parse function found for 'invalid number literal ...'` in expressions and as `constraint arguments must be literals, got Error` in constraint args. Fixed both: `noPrefixParseFnError` and `parseConstraintLiteral` now report a lexer `Error` token's message as is.
- `1e400` already fails in `parseFloatLiteral` (`invalid float literal`), so no range check was added in the lexer.
- Updated `FLOAT` in both grammar files and the graph nodes for lexer, parser.parser, parser.expression, parser.primary and parser.literal. Rewrote the legacy headers of `parser/literal.go` and `parser/expression_test.go` on touch.
- `go test ./...`, `make docs-graph` pass. `golangci-lint` panics locally (built with go1.26, toolchain go1.27); `go vet` passes.
- Noticed, not fixed (out of scope): several parser messages format `token.Range.From` (a `tokens.Pos`, which has no `String` method) with `%s`, printing `{%!s(int=1) ...}`, for example the "constraint arguments must be literals" message.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:39 - Test and CI pass for exponent literals
- Added end-to-end and gap tests: `runtime/exponent_literal_test.go` (rules over `1e5`, `2.5e-1`, unary minus, precedence, list and map literals, field `x.e5`), `lang_test/0050-exponent-literals.sentrie` (lets, collections, constraint args), `cmd/validate_test.go` (accepts exponent policy; malformed `1e+` reports the lexer message at `3:35-38`), `parser/exponent_literal_test.go` (unary minus, precedence, collections, `x.e5` field access, unterminated string in expression and constraint positions), and `TestLexerExponentMarkerOutsideNumbers` in `lexer/lexer_test.go`.
- Bite proof: reverting `lexer/lexer.go`, `parser/parser.go` and `parser/literal.go` to main each fails new tests.
- Gate: `make test`, lint (CI-pinned v2.4.0), gofmt, `make docs-graph`, race on lexer, parser, cmd and lang_test all pass.
- Finding left for the implementer: `noPrefixParseFnError` and `parseConstraintLiteral` now print an unknown-character error token's value alone, so `rule r = $` fails with just `$` and no "unexpected character" wording or position text. Not pinned by a test.
- Note: negative constraint args (`@min(-2.5e-1)`) still fail; that is #146.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-10-31 14:44 - Address round 1 review findings
- Warning 1 (unknown character reported bare as `parsing error at ...: $`): fixed in the lexer, whose unknown-character error token now carries `unexpected character "$"`. Quoted with `%q` on `string(char)`; `%q` on the rune printed `'$'`. Added `TestLexerUnknownCharacterNamesTheCharacter` and parser `TestParseErrorUnexpectedCharacter` (`rule r = $` and `@min($)`). The lexer graph node lists the message.
- Warning 2: PR description Testing notes now list the test agent's additions, and Changes by area > Lexer mentions the unknown-character message.
- Suggestion taken: `parseFloatLiteral` reports `invalid float literal` with `errorAtf(token.Range, ...)`, so `1e400` points at its own span (`3:12-17`) without the `tokens.Pos` formatting artifact; `TestParseProgramRejectsOutOfRangeExponent` asserts the span. The wider `tokens.Pos` `%s` bug stays out of scope (see the earlier note).
- Suggestion declined: folding `parser/exponent_literal_test.go` into `expression_test.go`. The separate file groups the exponent-specific cases the test agent wrote, and moving them is churn with no behaviour change.
- Commit queue: the unknown-character change in `lexer/lexer.go` folds into the queued, not yet run lexer commit, because the queue commits whole files and that path is already on it. The `parseFloatLiteral` span fix has its own commit, queued before the parser commit whose test asserts it.
- `go test ./lexer/... ./parser/... ./cmd/... ./lang_test/... ./runtime/... ./index/...` and `make docs-graph` pass.
- _Model: claude-opus-5-5, role: implementer subagent_
