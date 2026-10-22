## Binaek Sarkar (binaek89@gmail.com)

### 23:30 - Pick up #243
- Branch cut from `225-evaluator-deadline`: it builds on #142's deadline plumbing in `serve`, and stacking on #225 avoids a conflict in the README "Evaluation Limits" section both edit.
- Design agreed with the maintainer in conversation:
  - header `Sentrie-Timeout` taking an RFC 3339 duration;
  - `--eval-max-timeout` as a ceiling that a request may go up to;
  - `--eval-timeout` stays the default.
- Calls made for the maintainer to confirm in review (listed in the issue):
  - a decimal fraction on seconds, which RFC 3339's grammar doesn't allow;
  - clamping over-ceiling requests rather than rejecting them;
  - startup errors for a ceiling below the default, and for a ceiling with an unbounded default.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_
