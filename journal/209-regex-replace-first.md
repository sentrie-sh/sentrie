## Binaek Sarkar (binaek89@gmail.com)

### 22:02 - Pick up #209
- Branch cut from `208-remove-dead-base64`, which sits on `111-throw-builtin-errors`. #111 rewrites every error site in `builtin_regex.go` to `throwErr`, so this fix stacks on it rather than conflicting with it on `main`. Second layer of the chain 208, 209, 210, 211, 212, 213.
- Plan: `replace` replaces the first match only; `replaceAll` keeps replacing every match.
- _Model: claude-opus-5-5_

### 22:02 - Replace only the first match in `regex.replace`
- Go's `regexp` has no first-match replace. Added `replaceFirst`: `FindStringSubmatchIndex` for the leftmost match, then `ExpandString` for the replacement, so `$1`-style references behave exactly as in `ReplaceAllString` (used by `replaceAll`, unchanged). No match returns the input unchanged.
- Considered `ReplaceAllStringFunc` with a "first call only" flag: rejected, it does not expand `$n` references, so `replace` and `replaceAll` would disagree on the same replacement string.
- Empty matches: `replace("x*", "abc", "-")` gives `-abc` (leftmost empty match at 0), matching JavaScript's non-global `String.prototype.replace`.
- Tests in a new `builtin_regex_test.go` on `JSTestSuite`: first-only vs all, no match, `+` runs, capture references, anchors and empty matches, plus both forms throwing on a bad pattern.
- Breaking: anything relying on `replace` rewriting every match must switch to `replaceAll`. No in-repo fixture or example uses `@sentrie/regex`.
- `.d.ts` already promised first-only; added the `$1` note and the pre-#209 behaviour, and rewrote its legacy header. Graph: `runtime.js.builtins` gotcha added.
- _Model: claude-opus-5-5_
