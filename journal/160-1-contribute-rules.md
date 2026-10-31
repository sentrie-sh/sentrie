# 160-1-contribute-rules

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #160
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Each rule has exactly one definition, so results cannot be built up from independent conditions.
- Decision (user, 2026-10-31): a new `contribute` keyword, e.g. `contribute deny when c { yield ... }`, adding to a declared rule. Result ordered by declaration, undefined contributions skipped, boolean rules OR their contributions. Contributions combine in a deterministic order (#143).
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
