# 163-string-ordering

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #163
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: `<`, `<=`, `>`, `>=` accept only numbers; two strings should compare by byte order (`strings.Compare`), mixed kinds stay an error.
- No design choice left open by the issue.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
