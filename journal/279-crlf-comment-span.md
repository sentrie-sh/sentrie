## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #279
- In a CRLF file a comment token's span covers the `\r` before the newline, one column further than the same line in an LF file.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: stop a comment before a `\r` that is followed by `\n`, and pin the span with an LF and CRLF lexer test.
- _Model: claude-opus-5-5_
