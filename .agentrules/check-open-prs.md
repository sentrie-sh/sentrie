# Check Open PRs Before Starting Work

Always check open PRs before starting work.

**Why:** Avoid duplicating or conflicting with work already in flight on an open PR.

**How to apply:** At the start of any task in this repo, run `gh pr list` and look for open PRs relevant to the area being touched before beginning implementation. Also check `git worktree list`: a worktree under `worktrees/` may hold work that has not been pushed yet.

Related: [Always work in a worktree](work-in-worktree.md), [Issue workflow](issue-workflow.md)
