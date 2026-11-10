## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #149
- Invalid operands give silent results: `[10,20,30]["foo"]` is 10, `[10,20,30][1.7]` is 20, `count(5)` is undefined, and `"a" + 1` is `"a1"`.
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A string or non-integer list index is an error.
  - A missing dict key or an out-of-range integer index stays undefined.
  - `count` on anything but a list, dict or string is an error; `count(undefined)` stays undefined.
  - Only `count` is fixed here; a follow-up issue is filed to audit the other builtins for the same problem.
  - `+` with a string follows JavaScript for scalars: string + number or bool concatenates in either order (`"a" + 1` is `"a1"`); undefined propagates as undefined; a list or dict with a string is an error.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_
