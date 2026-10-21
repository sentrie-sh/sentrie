## Binaek Sarkar (binaek89@gmail.com)

### 22:00 - Pick up #216
- Branch cut from `214-js-leftover-cleanup`. Chain 214, 216, 217: #217 changes the same `parseUse` code and the same fixtures, so the span fix goes first and #217 stacks on it.
- _Model: claude-opus-5-5_

### 22:04 - Root cause and fix
- Reproduced with a scratch parser test: the span's `To` is the last token's lexer end, and the lexer's end position is `currentPosition()` after consuming the token, i.e. the position of the following rune. When that rune is `\n`, `readRune` has already bumped the line and reset the column, so the end reads as column 0 of the next line. At EOF the column is not bumped, so the end is one column short. Offsets are right in all three cases.
- So the bug is in the lexer, not in which token `parseUse` picks. Fixing `currentPosition` would move the end of every token followed by a newline, which changes spans everywhere; the task is to fix `use` only, so the fix is parser-local.
- Added `tokenEnd` in `parser/use.go`: rebuilds the end from the token's start plus its source width in runes (byte width from offsets, less the multi-byte extra bytes of the value, which string escapes carry over unchanged). It agrees with the lexer exactly when the next rune is on the same line, so a `use` followed by a space gets the same span as before. A value containing a newline (heredoc) keeps the lexer's end.
- Also: the single-segment form `from @local` never moved `To` past the `@`; it now ends at the segment.
- Convention chosen: the end is exclusive (just past the last rune), matching what the lexer already reports for a token followed by a space.
- Tests: every `from` form, with and without `as`, multi-byte and escaped string paths, each followed by a newline, a space and EOF; plus a whole-program test. The new tests fail 21 assertions against the old code.
- Survey of other statements (scratch test over a program with one of each): namespace, shape, export shape, title, description, version, tag, fact (with and without default), let and `export decision` (with and without `attach`) all end at column 0 of the next line when they end a line. Rule and derive spans end on their own line. Not fixed here; reported to the parent.
- Other lexer quirks found on the way (reported, not fixed): columns are 0-based on the first line and 1-based on every later line (`readRune` sets `column = 1` after a newline but starts at 0); `Range.String` adds 1 to `From.Line` but not to `To.Line` in the multi-line form.
- _Model: claude-opus-5-5_
