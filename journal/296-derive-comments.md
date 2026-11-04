## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #296
- A line comment inside a derive's body makes `sentrie validate` reject the derive (`derive body may only contain let declarations before yield (got *ast.CommentStatement)`), although blocks accept comments everywhere else.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: skip comment statements wherever a derive block's statements are walked (purity, cycle check, derive expression walker), with tests for comments before, between and after `let`s and in a lambda body inside a derive.
- _Model: claude-opus-5-5_
