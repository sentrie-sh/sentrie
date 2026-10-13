# Plan Source Paths

When writing or generating plan Markdown (for example under `.cursor/plans/`, or any agent's plan file), link source files with paths that are valid for anyone who clones the repo.

- Use paths **relative to the plan's repository root** (the directory that contains `.git` for that project), for example `runtime/eval_call_test.go`, `cmd/exec.go`.
- **Don't** use machine-local absolute paths in links or prose: no `/Users/...`, `/home/...`, `C:\...`, or workspace prefixes that include a username or machine path.
- For Markdown links, use repo-root-relative targets, for example `[eval_call_test.go](runtime/eval_call_test.go)`.
- When a plan spans several repositories in a multi-root workspace, say which repo each path is relative to (a short label or a section per repo), and still avoid absolute paths.

**Why:** A plan with absolute paths only works on the machine that wrote it.

**How to apply:** Applies to plan bodies, todos, and any embedded file list, not only when editing an existing plan.

Related: [Journal convention](journal-convention.md)
