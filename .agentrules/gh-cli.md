# gh CLI

All GitHub operations in this repo go through the `gh` CLI, not the web UI. The workflow rules in this directory assume it is installed and authenticated.

## Prerequisites

- Install `gh` and authenticate once with `gh auth login`.
- `gh auth status` reports your account, protocol, and token scopes if a command starts failing unexpectedly.
- Git remote operations may block on interactive authorization, unlike `gh`'s API calls: see [Batch origin access](origin-access.md).

## Commands the rules depend on

- `gh pr list` - check for open or in-flight PRs before starting work. See [Check open PRs before starting work](check-open-prs.md).
- `gh pr create --draft` - raise the draft PR at the end of the issue-start sequence. See [Issue workflow](issue-workflow.md) and [PR conventions](pr-conventions.md) for the required title and body shape.
- `gh pr edit --body-file` - refresh an open PR's body from the [PR description file](pr-description-file.md).
- `gh issue view <number>` - read an issue body before starting on it.
- `gh label list` - the authoritative list of labels; prefer this over trusting a hardcoded list in any doc.
- `gh api repos/sentrie-sh/sentrie/commits/main --jq .sha` - check whether local `main` is current without a fetch.

## Scope gotchas

A default `gh auth login` token does not carry every scope. If a command fails with a permissions error, check `gh auth status` before assuming the command is wrong:

- **GitHub Projects (v2)** needs the `project` scope. Without it every `gh project ...` call fails.
- **Pushing changes under `.github/workflows/`** through the GitHub API needs the `workflow` scope. Ordinary git pushes over SSH are unaffected.
- **Org-level writes** need `write:org`; read-only org access is enough for everything the rules here describe.

Related: [Batch origin access](origin-access.md), [Issue workflow](issue-workflow.md)
