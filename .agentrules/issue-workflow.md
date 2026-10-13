# Issue Workflow

When starting work on an issue, follow these steps in order. The sequence is built around a single origin touch: everything up to step 7 is local, per [Batch origin access](origin-access.md).

Asking an agent to pick up an issue authorizes this whole workflow, including its commits and its draft PR ([Commits](commits.md)).

1. Check for in-flight work with `gh pr list` before touching the area ([Check open PRs before starting work](check-open-prs.md)).
2. Confirm local `main` is current without a fetch: compare `gh api repos/sentrie-sh/sentrie/commits/main --jq .sha` against `git rev-parse main`. Only pull when they differ; never fetch speculatively.
3. Create the branch and its worktree in one command, off local `main`:
   `git worktree add -b <issue_number>-short-name worktrees/<issue_number>-short-name main`.
   Branches are named `<issue_number>-short-name`, and worktrees live under `worktrees/` ([Always work in a worktree](work-in-worktree.md)).
4. Write the first journal entry in the worktree at `journal/<issue_number>-short-name.md` ([Journal convention](journal-convention.md)).
5. Start `PR_DESCRIPTION_<issue_number>-short-name.md` in the worktree root ([PR description file](pr-description-file.md)).
6. Commit.
7. The one origin touch: `git push -u origin <branch>`, then `gh pr create --draft` (API-only once the branch is pushed), with the title and body taken from the PR description file. See [PR conventions](pr-conventions.md) for the required shape.

**Why:** Combines the worktree requirement with a concrete branch naming convention, and gets a draft PR up early for visibility, while keeping the human to a single authorization prompt rather than one per step.

For a change large enough to split, follow [Stacked PRs](stacked-prs.md): each layer repeats steps 3 to 7 with its own branch, worktree, journal and PR description file.

Related: [Stacked PRs](stacked-prs.md), [Always work in a worktree](work-in-worktree.md), [Check open PRs before starting work](check-open-prs.md), [Batch origin access](origin-access.md), [Journal convention](journal-convention.md), [PR conventions](pr-conventions.md), [PR description file](pr-description-file.md)
