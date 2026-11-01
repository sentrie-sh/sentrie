## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #285
- The `@in` number constraint is defined but cannot be written, because `in` is a keyword and the parser expects an identifier after `@`.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer: the parser accepts a keyword token as a constraint name right after `@`, where it cannot be ambiguous, so `@in` works as documented and nothing is renamed.
- _Model: claude-opus-5-5_
