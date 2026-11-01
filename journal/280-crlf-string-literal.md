## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #280
- A string literal that spans lines keeps the `\r` of a CRLF line break, so the same policy evaluates differently with CRLF and LF endings.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: drop a raw `\r` immediately followed by a raw `\n` inside a string literal, as heredocs do since #270; an escaped `\r` is unchanged.
- _Model: claude-opus-5-5_
