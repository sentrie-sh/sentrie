## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #280
- A string literal that spans lines keeps the `\r` of a CRLF line break, so the same policy evaluates differently with CRLF and LF endings.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: drop a raw `\r` immediately followed by a raw `\n` inside a string literal, as heredocs do since #270; an escaped `\r` is unchanged.
- _Model: claude-opus-5-5_

### 12:57 - Drop the CR of a CRLF line break inside a string literal
- `readString` wrote every raw rune, so `let s = "a<CRLF>b"` was "a\r\nb" in a CRLF file and "a\nb" in an LF file; `s == "a\nb"` was false for CRLF.
- Fix: new helper `atCRLFBreak` (current is `\r` and the next rune is `\n`); `readString` skips writing that `\r`, both for a plain rune and for the default escape branch (backslash then a raw line break), so `\<CR><LF>` equals `\<LF>`. An escaped `\r` takes its own `case 'r'` and is unchanged; a lone `\r` is unchanged.
- The token span is unchanged: it still covers the dropped bytes, since it is source provenance.
- Tests: `TestLexerStringLiteralLineBreaks` (values, escapes, span) and `TestStringLiteralInACRLFFileEvaluatesLikeLF` (pack with LF and CRLF files, same results). The runtime test failed for CRLF without the fix. Namespace `string/...` is rejected (keyword), so the test uses `text/crlf`.
- Updated the lexer graph node (new gotcha).
- _Model: claude-sonnet-5-5_

### 13:08 - Test gaps around CRLF string literals
- Added `TestLexerStringLiteralCRLFPositions`: tokens after a three-line string literal have the same kind, value, line and column in CRLF as in LF (the string's own span is two bytes longer, one per dropped \r), an unterminated multi-line string reports the same error position text as LF, \r\r\n drops only the last \r, and a \r at end of file inside an unterminated string is kept and still reported as unterminated.
- Ran the new tests with `lexer/lexer.go` from main: the positions and \r\r\n subtests fail without the fix; the unterminated-string and end-of-file subtests are guards that also hold on main.
- Ruled out: more runtime tests, since the existing end to end test already covers LF vs CRLF evaluation and the rest is lexer-level.
- _Model: claude-sonnet-5-5_
