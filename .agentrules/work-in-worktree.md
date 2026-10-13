# Always Work in a Worktree

Always do implementation work in a git worktree, not directly on the checked-out branch.

**Why:** Keeps the main working directory clean and avoids disrupting other in-progress work or branches.

**How to apply:** Before starting implementation work, create or reuse a git worktree rather than editing in the primary checkout. Worktrees live inside the primary checkout under `worktrees/<branch>`, which `.gitignore` excludes:

```sh
git worktree add -b <branch> worktrees/<branch> main
```

Remove a worktree once its PR has merged: `git worktree remove worktrees/<branch>`.

Related: [Check open PRs before starting work](check-open-prs.md), [Issue workflow](issue-workflow.md)
