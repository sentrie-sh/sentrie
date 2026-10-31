## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #270
- A heredoc in a file with CRLF line endings never terminates: the terminator line is `EOT\r`, which never equals the tag, so the lexer reads to end of file.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer before pickup: strip a trailing `\r` from every heredoc line, so the value has LF endings, and allow trailing spaces or tabs after the terminator tag, as the opening tag line already does.
- _Model: claude-opus-5-5_

### 14:44 - Strip CR from heredoc lines and accept a padded terminator
- `readHereDoc` now compares `strings.TrimRight(line, " \t\r")` with the tag, so a CRLF terminator (`EOT\r`) and one with trailing spaces or tabs both end the heredoc. Leading whitespace still does not.
- Body lines drop only a trailing `\r` (`strings.TrimSuffix`), so the value has LF endings; other trailing whitespace in a body line is kept, as before.
- Token end: the whole terminator line is read before it is known to be the terminator, so `l.end` is past the trailing whitespace by then. Ruled out resetting `l.end` by hand; instead the line loop records the end just past the line's first `len(tag)` runes (the tag is ASCII, so bytes equal runes) and `readHereDoc` returns it as the token's end. The trailing whitespace and newline stay for `skipWhitespace`, so later positions are unchanged.
- `readRune` counts `\r` as an ordinary column and only `\n` starts a line, so CRLF positions needed no other change.
- An unterminated heredoc now reports `heredoc <<<EOT has no terminator line "EOT": unterminated string literal: at <file>:<line>:<col>` at the `<<<`, not at end of file. The parser still wraps the error token as `no prefix parse function found for '...'`; changing that is in `parser`, outside this function, so left alone.
- Tests in a new `lexer/heredoc_test.go` rather than appending to `lexer/lexer_test.go`, which open PR #258 appends to. Two CRLF and padded-terminator cases added to `TestLexerTokenPositions`, and a padded heredoc added to the LF and CRLF program in `TestProgramStatementSpans`. All new cases fail on the old lexer.
- `lexer/coverage_errors_test.go` takes the new three-value return; its legacy header rewritten to SPDX on touch.
- _Model: claude-opus-5-5_
