## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #279
- In a CRLF file a comment token's span covers the `\r` before the newline, one column further than the same line in an LF file.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: stop a comment before a `\r` that is followed by `\n`, and pin the span with an LF and CRLF lexer test.
- _Model: claude-opus-5-5_

### 12:53 - Stop comments before the CR of a CRLF line break
- `readComment` read up to `\n`, so in a CRLF file the comment span and the raw text included the `\r`; the span ended one column past the same line lexed with LF.
- Fix: the loop stops at a `\r` whose next rune is `\n` (`peekAhead`), leaving it for `skipWhitespace`. A lone `\r` not followed by `\n` stays in the comment, as before.
- Test `TestLexerCommentSpanIgnoresCRLF` lexes line, indented, trailing, empty and in-between comments with LF and with every `\n` turned into `\r\n` and compares kind, value, line, column and span length (offsets legitimately differ by the extra bytes). Confirmed it fails without the fix.
- Updated the lexer graph node (comments bullet).
- _Model: claude-sonnet-5-5_

### 13:08 - Test gaps around CRLF comments
- Added `TestLexerCommentEndOfFileAndFollowingTokensIgnoreCRLF`: comment as the last line with and without a final line break, a trailing comment on the last line, a file that is only a comment, and several comments followed by tokens (every token's kind, value, line, column and span length equal between LF and CRLF). Also pins the token after a CRLF comment at line 2, column 1, a lone \r at end of file (stays in the span, trimmed from the text) and \r\r\n (only the \r before the \n ends the comment).
- Ran the new and existing comment tests with `lexer/lexer.go` from main: both fail without the fix (the lone \r and next-token position checks are guards that also hold on main, as intended).
- Ruled out: testing offsets directly (they legitimately differ by the extra bytes); runtime-level tests, since the lexer span is the whole change.
- _Model: claude-sonnet-5-5_
