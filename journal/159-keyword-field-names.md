## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #159
- Reserved words cannot be field names, so common JSON keys cannot appear after `.` (parser/access.go:32) or as shape fields (parser/shape.go:106).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Any keyword is accepted as a name after `.`, as a shape field name and as an attach name.
  - Keywords are not accepted as bare dict keys.
- Docs PR: sentrie-sh/website#82 (branch `159-keyword-field-names` on `next`). Sentrie PR: #322.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_
