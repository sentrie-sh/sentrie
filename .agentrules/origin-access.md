# Batch Origin Access

Assume every git operation that touches `origin` may require interactive authorization from a human. Depending on how a contributor's credentials are held, `git push`, `git fetch`, `git pull`, and `git ls-remote` can block on a hardware key touch, a biometric prompt, a passphrase, or a 2FA challenge.

Treat remote operations as something that cannot be assumed to run unattended. Two consequences follow.

## Batch origin access

Each origin operation may cost a human interruption, and a blocked one can stall a task indefinitely if nobody is at the machine. So:

- Group remote operations together rather than scattering them through a task. One fetch at the start and one push per batch of branches that are due beats four of each spread across the work.
- Batching groups pushes that are due at the same time. It never moves a push that a workflow requires at a fixed point. The issue pickup push and draft PR ([Issue workflow](issue-workflow.md) step 7) come before implementation, even when several issues are picked up together: they share one push, made before any of them is implemented. "Push once at the end" is the misreading to avoid.
- Never run a speculative or "just to be safe" fetch. Check whether the information is already available locally first.
- Do the local work for the step at hand first, then reach for origin once there is a real reason to. For an issue pickup, the local work is the branch, journal entry, PR description file and commit, not the implementation.
- If a sequence genuinely needs several origin touches, say so up front so the human knows to stay available, rather than surprising them with prompts one at a time.

## Prefer `gh` for reads

`gh`'s API-backed commands talk to GitHub over HTTPS using its own auth token rather than your git credentials, so they generally complete without an interactive prompt. Anything that shells out to git against `origin` may not.

Prefer the `gh` read over the git equivalent whenever both would answer the question:

- `gh pr list` / `gh pr view` instead of fetching to inspect branches
- `gh issue list` / `gh issue view` for issue state
- `gh api ...` for anything else readable through the API, for example `gh api repos/sentrie-sh/sentrie/commits/main --jq .sha` to check whether local `main` is current
- `gh label list` instead of guessing at labels

`gh pr create` against an already-pushed branch is API-only. It only touches origin if gh has to push the branch itself.

Related: [gh CLI](gh-cli.md), [Issue workflow](issue-workflow.md), [Check open PRs before starting work](check-open-prs.md)
