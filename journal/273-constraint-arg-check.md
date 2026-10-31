## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #273
- A broken constraint argument, such as an invalid regexp, is only reported for values of the constraint's own kind, so `x is not string@regexp("[")` is `true` for a number.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer before pickup: check constraint arguments when the pack is indexed (count, kinds, and that a regexp compiles), reported by `sentrie validate`, and at runtime check the arguments before the value's kind so a broken argument errors on every input. Constraint names stay with #113.
- _Model: claude-opus-5-5_
