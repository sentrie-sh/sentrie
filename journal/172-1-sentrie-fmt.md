# 172-1-sentrie-fmt

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-01 07:02 - Pick up issue #172
- Picked up as part of a wave-2 batch of two issues (174 first slice, 172), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no canonical formatter for .sentrie files, so style varies, diffs are noisy and docs examples drift.
- Decision: `sentrie fmt [paths...]` rewrites files in place, with `--check` (non-zero exit when anything would change) and `--diff`. It prints from the AST, so the parser keeps leading and standalone comments as well as trailing ones. A golden test over `lang_test/` and `example_pack/` checks that formatting is idempotent and that the output parses to an equivalent AST. If the change grows past about 800 lines it is split into a stack (comment retention and printer, then the command).
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
