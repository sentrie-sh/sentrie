# 180-exponent-literals

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #180
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Number literals with an exponent (`1e5`, `2.5e-3`) fail to parse; extend lexer number scanning to accept `[eE][+-]?digits` and produce a float literal.
- No design choice left open by the issue.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
