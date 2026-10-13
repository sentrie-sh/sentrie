# Batch Origin Access

Assume every git operation that touches `origin` may require interactive authorization from a human. Depending on how a contributor's credentials are held, `git push`, `git fetch`, `git pull`, and `git ls-remote` can block on a hardware key touch, a biometric prompt, a passphrase, or a 2FA challenge.

Treat remote operations as something that cannot be assumed to run unattended. Two consequences follow.

## Batch origin access

Each origin operation may cost a human interruption, and a blocked one can stall a task indefinitely if nobody is at the machine. So:

- Group remote operations together rather than scattering them through a task. One fetch at the start and one push at the end beats four of each spread across the work.
- Never run a speculative or "just to be safe" fetch. Check whether the information is already available locally first.
- Do the local work first (branching, edits, commits, journal entries), then reach for origin once there is a real reason to.
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
