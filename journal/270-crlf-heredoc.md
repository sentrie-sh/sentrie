## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #270
- A heredoc in a file with CRLF line endings never terminates: the terminator line is `EOT\r`, which never equals the tag, so the lexer reads to end of file.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer before pickup: strip a trailing `\r` from every heredoc line, so the value has LF endings, and allow trailing spaces or tabs after the terminator tag, as the opening tag line already does.
- _Model: claude-opus-5-5_
