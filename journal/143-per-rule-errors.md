## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #143
- One failing rule fails the whole policy, and output and error order is nondeterministic (runtime/executor.go:232-278).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A failing rule carries a per-rule `error` field and every other rule still evaluates.
  - `serve` returns 207 Multi-Status when any rule failed (including all of them), and 200 otherwise.
  - Output and error order stays nondeterministic: the user declined a total order, so that part of the issue is closed as won't fix.
  - `exec` exits non-zero when any rule errored; its output still carries every rule's result.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_
