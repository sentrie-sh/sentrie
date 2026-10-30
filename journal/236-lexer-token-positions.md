## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #236
- Lexer and token positions are inconsistent: a token's end is read after the next character, columns are 0-based on line one and 1-based after, and multi-line `Range.String` prints the end line one early.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:49 - Fix lexer positions and Range.String
- Settled on 1-based lines and 1-based columns on every line, with an exclusive end (just past the last rune, on that rune's line). `Pos` already documented 1-based lines and columns, diagnostics are conventionally 1-based, and the exclusive end matches what the lexer already reported mid-line and what #216's `tokenEnd` produced.
- `lexer.readRune` now steps past the consumed rune before reading the next one and records `end` (the position just past it). `currentPosition` is the current rune's own position; the new `endPosition` returns `end`. A newline now sits at the end of its own line instead of column 0 of the next.
- Every token's end is taken after its last rune is consumed. `-`, `==`, `=>` and `|>` used to take the end before the final read, one column short of every other token; they now match.
- A heredoc no longer consumes the newline after its terminator tag, so its end is just past the tag; `skipWhitespace` eats the newline as before.
- `Range.String` prints lines as stored (no `+1`), so the single- and multi-line forms agree. A zero `To` still prints the start only.
- Test expectations that encoded the old base were updated: 0-based lines in index and xerr tests, first-line 0-based columns and the EOF one-short end in `parser/use_test.go`. The use span tests that pin the end just past the last token are unchanged in substance.
- Checked every other consumer of `Line`/`Column`: `lexer/err.go`, `parser/parse.go` and `parser/primary.go` print the raw fields and were one line low before; they are now right with no change. Synthetic ranges in index tests with `Column: 0` assert nothing about printing, so they were left.
- Checked every statement kind in a sandbox program: namespace, shape, export shape, title, description, version, tag, fact, let, rule and export decision all end just past their last token at end of line.
- Left alone: `parser/block.go` ends a block expression at the start of its `}` rather than past it. It is not a statement span and is outside this issue.
- _Model: claude-opus-5-5_

### 15:49 - Remove the tokenEnd workaround from use
- `tokenEnd` (#216) rebuilt a token's end as its start column plus its source width in runes, keeping the lexer's end for a value containing a newline. With the lexer fixed, the lexer's end is the start column plus the source rune count, on the start line, for every single-line token. `tokenEnd`'s rune count is the byte width less the value's multi-byte extras; string escapes only drop ASCII backslashes, so the source and value carry the same multi-byte extras and the two are equal. For a heredoc it already returned the lexer's end. So it was provably a no-op and is removed, with its `unicode/utf8` import; spans take `Range.To` directly.
- The quoted-`@` path error now passes the string token's `Range` as is.
- `parser/use_test.go` span tests pass unchanged after removal, which confirms the no-op.
- Updated the `parser.use` graph node: dropped the `unicode/utf8` edge and replaced the rebuilt-end gotcha.
- _Model: claude-opus-5-5_

### 15:50 - Update the lexer and tokens graph nodes
- `lexer`: new position-model gotcha (1-based lines and columns, exclusive end on the last rune's line) and the heredoc terminator newline note.
- `tokens`: `Range.String` contract now states the base, the exclusive end and the start-only cases; the risk note says `Column: 0` is not a lexer position.
- No `parser` node described the old position semantics or listed #236 as a known defect; `parser.use` was updated with the `tokenEnd` removal. `make docs-graph` passes.
- _Model: claude-opus-5-5_

### 15:55 - Statement span and diagnostic position tests
- Added `parser/span_test.go` to the parser suite. `TestStatementSpanEndsAtLastToken` covers namespace, shape (block and alias), export shape, export derive, title, description, version, tag, fact (plain, optional, default), let (scalar, typed, list, dict), rule, export decision (plain and with attach) and use, each on line 2 after two spaces, followed by a newline, CRLF, a space, a semicolon, a same-line token, a trailing comment and EOF; it asserts the full From and To and that the offsets slice out exactly the statement text.
- `TestStatementSpanCountsColumnsInRunesAndOffsetsInBytes` pins column-vs-byte with 2-, 3- and 4-byte runes. `TestProgramStatementSpans` parses one program with LF and again with CRLF and checks every statement's line and column, the sliced offsets and the printed multi-line range (shape, list `let`, policy). `TestStatementSpansAroundHeredoc` covers the heredoc terminator newline and a heredoc at EOF. `TestParseErrorsPrintOneBasedPositions` pins printed error positions on line 1 and later lines, for parse and lexer errors, and the multi-line form.
- Shown to bite: reverting `lexer/lexer.go` to main fails all six tests, reverting `tokens/position.go` fails the multi-line, program, heredoc and diagnostic tests, and restoring only the heredoc newline consumption fails the heredoc test.
- Ruled out: asserting the end of a statement whose last expression is a block (`rule ... { yield true }`, `let x = { yield 1 }`, `derive d = ... { yield n }`, `attach a as { yield 1 }`). They end at the start of the closing `}` (parser/block.go, left alone), so the multi-line rule test pins only the end line. List and dict literals end past `]`/`}` correctly.
- Ruled out: trailing-comment suffix for statements that end in an expression; the expression wraps the comment and the statement spans it, which is intended.
- Observed, not changed: a heredoc whose terminator line has trailing whitespace or a CR (CRLF files) never terminates, because the terminator must equal the tag exactly. This predates the branch.
- _Model: claude-sonnet-5-5_

