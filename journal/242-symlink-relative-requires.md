# 242-symlink-relative-requires

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #242
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Nested relative `require`s in TypeScript modules fail when the pack root path goes through a symlink.
- No design choice left open by the issue; the fix must keep `@local` module path confinement intact.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
