# 172-2-fmt-command

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-02 02:09 - Pick up layer 2 of issue #172
- Layer 2 of the #172 stack, cut from `172-1-sentrie-fmt` (PR #288), which holds comment retention in the parser and the `format` package.
- Problem: there is no command that applies the formatter to files.
- Decision: `sentrie fmt [paths...]` takes files and directories (recursing for `*.sentrie`), rewrites in place by default, `--check` exits non-zero and lists files that would change without writing, and `--diff` prints a unified diff without writing. A file that does not parse is reported with its error and makes the exit code non-zero; the other files are still processed. This is the top layer, so its PR closes #172.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.
