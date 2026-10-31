# 165-1-union-any-types

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #165
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no union type and no top type other than `document`, which is dict-shaped.
- Decision (default proposed to the user on 2026-10-31, not overridden): `T1 | T2` in any typeref position, a value validates if it matches any member, constraints bind to the member they follow; `any` accepts every value including scalars. Discriminated unions are out of scope. When no member matches, report the closest member's failure.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
