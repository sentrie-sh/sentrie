# 166-1-exact-integers

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #166
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: All numbers are float64, so integers above 2^53 lose precision in literals and JSON facts.
- Decision (default proposed to the user on 2026-10-31, not overridden): keep integers that fit in int64 exact (JSON facts via `UseNumber`, integer literals the same); `int op int` stays exact and overflow is an error; mixed int/float promotes to float; exact integers serialise without exponent notation.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
