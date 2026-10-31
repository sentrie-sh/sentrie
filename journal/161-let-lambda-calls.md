# 161-let-lambda-calls

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #161
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: `let f = (x) => {...}; f(2)` fails with an unresolved-import error; call resolution never considers a let-bound callable.
- Decision (default proposed to the user on 2026-10-31, not overridden): option 1, support it. A callee identifier that resolves to a `let` holding a callable is invoked with the same rules as a builtin lambda argument. The "unable to resolve import" message appears only when a real module alias is involved.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
