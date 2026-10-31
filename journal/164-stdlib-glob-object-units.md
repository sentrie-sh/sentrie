# 164-stdlib-glob-object-units

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #164
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: First slice of the stdlib tracking issue: `glob.match`, object utilities (`get`, `keys`, `remove`, `union`, `filter`) and units parsing.
- Decision (default proposed to the user on 2026-10-31, not overridden): this PR is one slice of tracking issue #164, so its title says `Updates #164`. Native builtins via the `builtins` registry for these pure functions. X.509, YAML and JSON Schema are later slices.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
