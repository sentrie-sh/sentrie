# Commits

## When to commit

- Do not commit unless explicitly told to.
- Asking an agent to pick up an issue counts as that instruction for the [Issue workflow](issue-workflow.md), including its commits and its draft PR. Outside that workflow, the rule above stands.
- Never commit `PR_DESCRIPTION_*.md` files ([PR description file](pr-description-file.md)). They are in `.gitignore`. Other Markdown, such as `docs/codebase-graph/` nodes, journals, `AGENTS.md` and `.agentrules/`, is committed with the change it belongs to.

## Commit scope

Prefer small, focused commits over large bundles. When practical, make one logical change per commit. Combine work into one commit only when the changes are inseparable, for example the same fix, the same slice of a feature, or the same mechanical rename. Edits grouped into one commit must be tightly coupled. Don't bundle unrelated changes.

## Commit messages

- Use short imperative subject lines.
- Don't use conventional-commit prefixes (no `refactor(scope):`, `feat(scope):` and so on). Prefer plain imperative messages: "Align X with Y", "Wire policy loading with parsing".
- Add a body only when it adds value.
- No agentic signatures or co-author trailers ([No agentic signatures](no-agentic-signatures.md)).

**Why:** Small, plainly named commits make review, bisecting and reverting easy. PR descriptions are local drafts, so committing them leaks one branch's description into another.

**How to apply:** Before each commit, read `git diff --cached --stat` and confirm it is one logical change with no `PR_DESCRIPTION_*.md` in it.

Related: [Issue workflow](issue-workflow.md), [PR conventions](pr-conventions.md), [No agentic signatures](no-agentic-signatures.md), [License headers](license-headers.md)
