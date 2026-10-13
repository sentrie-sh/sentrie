# PR Description File

When you make any changes, update the `PR_DESCRIPTION_<branch_name>.md` file at the root of the checkout, describing all the changes in the branch comprehensively. It is the working draft of the pull request body, and `gh pr create` reads from it.

## Scope

`PR_DESCRIPTION_<branch_name>.md` MUST describe only the changes introduced in this branch, compared to the base branch (usually `origin/main`). Do not include or summarize work that already exists on the base or that was merged from other branches.

- To determine scope, list the commits and file changes with `git log origin/main..HEAD` and `git diff origin/main..HEAD`, then document only those.
- Do not include code or logic that was added in the branch, then removed and never re-added.
- Summarize in PR description format: short, bulleted points that are easy to scan, one main idea per bullet. Prefer sub-bullets over paragraphs.

## Shape

- **First line is the PR title:** a single `#` heading whose text is the full title string. Do not add a separate `# Title` section or any second top-level heading. The title rules are in [PR conventions](pr-conventions.md).
- **Then the sections**, in order: Summary, What this PR does, Changes by area, Review notes, Testing notes, Dependency changes.
- **Changes by area:** format each area as a `###` subheading (for example `### Runtime evaluator and execution`) followed by bullets, not as bold bullet titles.
- **Review notes:** keep them current. They must cover every critical area reviewers should concentrate on.

## Raising the PR from it

Strip the first-line title and pass the rest as the body:

```sh
title=$(head -n1 PR_DESCRIPTION_<branch>.md | sed 's/^# //')
tail -n +2 PR_DESCRIPTION_<branch>.md > /tmp/pr-body.md
gh pr create --draft --title "$title" --body-file /tmp/pr-body.md
```

After later edits, update the open PR the same way with `gh pr edit --body-file`.

## Never committed

The file is a local draft. `PR_DESCRIPTION_*.md` is in `.gitignore` and must never be committed ([Commits](commits.md)).

**Why:** Drafting the description as the work happens keeps it accurate. Keeping it out of git stops one branch's description from leaking into another.

**How to apply:** Update the file in the same session as every change, and rebuild the PR body from it rather than editing the body on GitHub.

Related: [PR conventions](pr-conventions.md), [Commits](commits.md), [Journal convention](journal-convention.md)
