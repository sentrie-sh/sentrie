## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #299
- A `dict[T]` type never checks its values against `T` at runtime, so `{"a": "not a number"} is dict[number]` is `true`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: validate every value of a `dict[T]` against `T`, with an error naming the failing key as `list[T]` names an index, in every type position.
- _Model: claude-opus-5-5_
