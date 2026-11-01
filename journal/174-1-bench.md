# 174-1-bench

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-01 07:02 - Pick up issue #174
- Picked up as part of a wave-2 batch of two issues (174 first slice, 172), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no way to measure policy evaluation cost or to catch performance regressions such as #140.
- Decision: this first slice adds `sentrie bench <policy/rule>` (facts input, `-count`, `-benchtime`, Go benchmark output so `benchstat` works) and Go benchmarks for the lexer, parser, index, executor and startup on `example_pack`. `exec --profile`, `exec --explain` and performance CI are left to later slices.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
