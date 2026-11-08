# Website PR Base Branch

Every pull request in `sentrie-sh/website` targets `next`, the branch that holds documentation for the next publish, never `main`.

The one exception is the release PR (`next` into `main`), which publishes the documentation once the matching Sentrie release is out.

**Why:** `main` is the live site and documents released behaviour. A docs change for a Sentrie PR describes a feature that has not shipped yet, so it waits on `next` and reaches `main` together with the release. Naming the branch `next` rather than after a version keeps this rule, and every agent's `--base`, the same from one release to the next.

**How to apply:**

- Cut website branches from `next`, and pass `--base next` to every `gh pr create --repo sentrie-sh/website`. `gh` defaults to `main` when `--base` is left out.
- Before handing off a batch of website PRs, check the bases: `gh pr list --repo sentrie-sh/website --state open --json number,baseRefName,headRefName`. Retarget a stray one with `gh pr edit <number> --repo sentrie-sh/website --base next`.
- `next` outlives each release. The website repo squash-merges and deletes a merged PR's head branch, so merging the release PR would delete `next` and move the open docs PRs to `main`. Protect `next`, or restore it from `main` straight after the merge and retarget those PRs back, before merging the release PR.

Related: [PR conventions](pr-conventions.md), [Issue workflow](issue-workflow.md), [gh CLI](gh-cli.md)
