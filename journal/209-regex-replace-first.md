## Binaek Sarkar (binaek89@gmail.com)

### 22:02 - Pick up #209
- Branch cut from `208-remove-dead-base64`, which sits on `111-throw-builtin-errors`. #111 rewrites every error site in `builtin_regex.go` to `throwErr`, so this fix stacks on it rather than conflicting with it on `main`. Second layer of the chain 208, 209, 210, 211, 212, 213.
- Plan: `replace` replaces the first match only; `replaceAll` keeps replacing every match.
- _Model: claude-opus-5-5_
