## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #236
- Lexer and token positions are inconsistent: a token's end is read after the next character, columns are 0-based on line one and 1-based after, and multi-line `Range.String` prints the end line one early.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:49 - Fix lexer positions and Range.String
- Settled on 1-based lines and 1-based columns on every line, with an exclusive end (just past the last rune, on that rune's line). `Pos` already documented 1-based lines and columns, diagnostics are conventionally 1-based, and the exclusive end matches what the lexer already reported mid-line and what #216's `tokenEnd` produced.
- `lexer.readRune` now steps past the consumed rune before reading the next one and records `end` (the position just past it). `currentPosition` is the current rune's own position; the new `endPosition` returns `end`. A newline now sits at the end of its own line instead of column 0 of the next.
- Every token's end is taken after its last rune is consumed. `-`, `==`, `=>` and `|>` used to take the end before the final read, one column short of every other token; they now match.
- A heredoc no longer consumes the newline after its terminator tag, so its end is just past the tag; `skipWhitespace` eats the newline as before.
- `Range.String` prints lines as stored (no `+1`), so the single- and multi-line forms agree. A zero `To` still prints the start only.
- Test expectations that encoded the old base were updated: 0-based lines in index and xerr tests, first-line 0-based columns and the EOF one-short end in `parser/use_test.go`. The use span tests that pin the end just past the last token are unchanged in substance.
- Checked every other consumer of `Line`/`Column`: `lexer/err.go`, `parser/parse.go` and `parser/primary.go` print the raw fields and were one line low before; they are now right with no change. Synthetic ranges in index tests with `Column: 0` assert nothing about printing, so they were left.
- Checked every statement kind in a sandbox program: namespace, shape, export shape, title, description, version, tag, fact, let, rule and export decision all end just past their last token at end of line.
- Left alone: `parser/block.go` ends a block expression at the start of its `}` rather than past it. It is not a statement span and is outside this issue.
- _Model: claude-opus-5-5_
