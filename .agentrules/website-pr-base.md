# Website PR Base Branch

Every pull request in `sentrie-sh/website` targets the website's release branch, currently `v0.0.5`, never `main`.

The one exception is the release PR itself (`v0.0.5` into `main`), which ships the documentation once the matching Sentrie release is out.

**Why:** The website documents released behaviour. A docs change for a Sentrie PR describes a feature that has not shipped yet, so it waits on the release branch and reaches `main`, and the live site, together with the release.

**How to apply:**

- Cut website branches from `v0.0.5`, and pass `--base v0.0.5` to every `gh pr create --repo sentrie-sh/website`. `gh` defaults to `main` when `--base` is left out.
- Before handing off a batch of website PRs, check the bases: `gh pr list --repo sentrie-sh/website --state open --json number,baseRefName,headRefName`. Retarget a stray one with `gh pr edit <number> --repo sentrie-sh/website --base v0.0.5`.
- When the release PR merges and a new release branch is cut, update the branch name in this rule and its `AGENTS.md` entry in an `[AgentRule]` PR.

Related: [PR conventions](pr-conventions.md), [Issue workflow](issue-workflow.md), [gh CLI](gh-cli.md)
