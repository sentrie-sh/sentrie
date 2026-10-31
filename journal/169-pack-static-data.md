# 169-pack-static-data

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #169
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: A policy can only see per-request facts; there is no way to ship static reference data with a pack.
- Decision (default proposed to the user on 2026-10-31, not overridden): a `[data]` table in the pack manifest maps names to JSON files loaded once at pack load, optionally validated against a shape, exposed read-only. Access syntax settled by the implementer against existing namespace resolution and recorded here.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
