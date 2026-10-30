## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #236
- Lexer and token positions are inconsistent: a token's end is read after the next character, columns are 0-based on line one and 1-based after, and multi-line `Range.String` prints the end line one early.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_
