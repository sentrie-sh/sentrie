# agentrules-convention

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-13 14:42 - Survey and conflict resolution
- Task: move Sentrie's agent conventions from `AGENTS.md` and `.cursor/rules/` into the dryft-os `AGENTS.md` + `.agentrules/` layout, and adopt the dryft-os working rules adapted to Sentrie. Conventions only: no product code, no workflow files.
- Surveyed `AGENTS.md`, `.cursor/rules/*.mdc`, `.cursor/agents/*.md`, `CONTRIBUTING.md`, `Makefile`, `go.mod`, `.goreleaser.yml`, `.github/workflows/*`, `.github/pull_request_template.md`, `THIRD_PARTY_POLICY.md`, `docs/codebase-graph/overview.md`, the top-level layout, `git log --oneline -50` and the last 30 merged PRs.
- `gh pr list`: no open PRs, so nothing overlaps this work.
- Local `main` matched `gh api repos/sentrie-sh/sentrie/commits/main` (`d7a66e3`), so no pull was needed.
- Repo facts that shape the rules:
  - The repo is public, so standard GitHub runners cost nothing. The GitHub Actions rule argues from speed and noise, not money.
  - Squash merge only, merged branches are deleted, and `main` has no branch protection.
  - `go.mod` pins `go 1.25.0` with no `toolchain` line.
  - The Makefile has only `build`, `clean-git`, `docs-graph` and `gen`. It has no `test` or `lint` target. `make build` stamps a wall-clock date and `treeState=dirty`, so the dryft-os reproducibility section was dropped.
  - No persisted generated ids. `uuid` is used only for the process instance id, the request id and the policy-facing JS `uuid` builtin, so `id-generation` was skipped.
  - `protected-files-maintainers.yml` covers `CLA.md`, `LICENSE`, `LICENSE-DUAL.md`, `THIRD_PARTY_POLICY.md` and `.github/**`. It does not cover `AGENTS.md`, `.agentrules/**` or `.cursor/**`.
  - The `sandbox.go` template fails `go vet -tags sandbox sandbox.go` (`declared and not used: ctx`). Logged as a follow-up, not fixed here (product code).
- Conflicts the prompt did not resolve, put to the user:
  - Worktree location: the prompt said `../sentrie-<branch>`, but all 12 existing worktrees live under `worktrees/<branch>`, and `main` has an uncommitted `.gitignore` entry for `worktrees/`. User chose `worktrees/<branch>`. This branch's worktree is `worktrees/agentrules-convention`, and the `.gitignore` entry ships in this PR.
  - Legacy license headers: the cursor rule said both "update legacy headers to the SPDX form" and "only new files use the SPDX form". The user said they want to move away from cursor rules to generic agent rules, which did not pick a side. History decided it: every fix PR from #122 to #132, plus #93 and #94, rewrote touched legacy headers to the two-line SPDX form. The rule now says rewrite on touch.
  - `.cursor/agents/config-writer.md` sent new conventions to `.cursor/rules/` or a central `agents.md`. User chose to repoint it to `.agentrules/` + `AGENTS.md` in this PR.

### 2026-10-13 14:50 - Rule files, index and pointers
- Wrote 25 rule files under `.agentrules/`, each with a title, the rule, **Why:**, **How to apply:** and a `Related:` line.
  - Migrated from Sentrie (wording kept except where a resolution changed it): `codebase-knowledge-graph`, `sandbox`, `pr-description-file`, `license-headers`, `commits`, `go-test-conventions`, `evalderive-resolution`, `plan-source-paths`.
  - Adapted from dryft-os: `work-in-worktree`, `origin-access`, `trace-data-flow-for-races`, `parallelize-with-subagents`, `token-efficiency`, `journal-convention`, `no-em-dash`, `no-agentic-signatures`, `check-open-prs`, `issue-workflow`, `pr-conventions` (merged with Sentrie's PR rules), `stacked-prs`, `github-actions`, `gh-cli`, `build-and-test`, `go-project-layout`, `current-libraries`.
  - Skipped: `id-generation` (no persisted generated ids), plus `db-migrations-sqlc`, `opentofu`, `event-payload-compatibility` and `epic-story-task-backlog` (dryft-os only, per the task).
- Rewrote `AGENTS.md` as an index: Sentrie's opening line, the dryft-os preamble, then the eight dryft-os sections in order. "Deploy and release" has no rules yet.
- Problem found while verifying the example commands in `token-efficiency`: `go test ./lang_test/` inside the worktree failed with "main module (github.com/sentrie-sh/sentrie) does not contain package .../worktrees/agentrules-convention/lang_test".
  - Cause: an ambient `~/go.work` lists `./sentrie/sentrie` (the primary checkout), so `go` resolved the module to the primary checkout, not the worktree.
  - Fix in the rules: `build-and-test` now requires `GOWORK=off` on every `go` invocation when a `go.work` sits above the checkout, and every command in `token-efficiency` carries it. The Makefile doesn't set it yet (follow-up), so the rule says `GOWORK=off make docs-graph`.
  - With `GOWORK=off`, both example test commands pass, and so does `make docs-graph`.
- `.cursor/rules/*.mdc` now keep only their front matter (`description`, `globs`, `alwaysApply`) and one line linking the `.agentrules/` file.
- Repointed `.cursor/agents/config-writer.md`: rules go to `.agentrules/<kebab-name>.md` with an `AGENTS.md` bullet, and `.cursor/rules/` holds only thin pointers for glob attachment. `code-reviewer` and `test-writer` are untouched and are referenced from `parallelize-with-subagents` as prompt bases.
- Resolutions the task specified:
  - "NEVER commit any `*.md`" is narrowed to "never commit `PR_DESCRIPTION_*.md`", and that pattern is now in `.gitignore`.
  - `PR_DESCRIPTION_<branch>.md` stays the drafting file and feeds `gh pr create --body-file`, with the first-line title stripped and passed as `--title`.
  - Picking up an issue authorizes the issue workflow's commits.
  - The cursor-signature ban became the general `no-agentic-signatures` rule.
- Em dash sweep: one literal U+2014 got into `no-em-dash.md` through an escape sequence, and one came from a copied `config-writer.md` line. Both were replaced. The em dashes already in the untouched lines of `config-writer.md` were left alone.

### 2026-10-13 14:51 - Commits and draft PR
- Seven commits, one logical change each: `.gitignore`, migrated rules, adopted rules, the `AGENTS.md` index, the Cursor rule pointers, `config-writer`, and this journal.
- No attribution trailers, per `no-agentic-signatures`.
- `PR_DESCRIPTION_agentrules-convention.md` is drafted and is ignored by the new `.gitignore` pattern (checked with `git check-ignore`).
- The one origin touch comes next: `git push -u origin agentrules-convention`, then `gh pr create --draft` with the title and body taken from the PR description file.

### 2026-10-13 16:04 - Rebase onto main
- Draft PR #181 opened. Then `main` moved to `11070ea` (#183: ExecPolicy panic-recovery test fix, `runtime/executor_test.go` and `docs/codebase-graph/runtime.executor.md`).
- No overlap with this branch's files. `git rebase origin/main` applied all seven commits cleanly.
- `GOWORK=off make docs-graph` still passes after the rebase.
- Force-pushed with the lease pinned to the PR head GitHub reported.

### 2026-10-13 16:10 - Drop Cursor rules, prose em dashes, PR template
- The maintainer asked to translate the Cursor rules into agent rules and to take care of the follow-ups.
- Read with the earlier answer ("move away from cursor rules to generic agent rules"), that meant deleting `.cursor/rules/*.mdc` outright rather than keeping pointers.
  - Each rule's scope already lives in its `.agentrules/` text (`go-test-conventions`: every `**/*_test.go`; `evalderive-resolution`: `runtime/eval_call.go` and `runtime/eval_ident.go`).
  - This supersedes the thin-pointer step from the 14:50 entry.
- `config-writer.md` no longer mentions pointer files. It says not to create `.cursor/rules/` files, and to state path scope in the rule text.
- `code-reviewer.md` now checks changed code against the rules indexed in `AGENTS.md` instead of `.cursor/rules/`.
- `.cursor/agents/` stays. Those files are subagent briefs, not rules.
- Replaced every em dash in prose files with " - ": `CONTRIBUTING.md`, `builtins/ADDING_BUILTINS.md`, `.cursor/agents/*.md`, and the `parser.not` and `api.problem_details` graph nodes.
  - Em dashes in source files (`.go` comments and error strings, `grammar/grammar.ebnf`, `grammar/grammar.peg`) and in `ci.yml` step names are left for their own PRs.
- `.github/pull_request_template.md` now has the six sections from `pr-conventions`, and keeps the dependency guidance.
- Follow-up split, following this branch's own rules:
  - workflow hygiene and protected paths: `[Workflow]` PR, based on this branch;
  - Makefile and `sandbox.go` tooling: no-issue tooling PR, based on this branch;
  - `golang.org/x/exp` to std, and source em dashes: product code, so each gets an issue and a PR off `main`.

### 2026-10-14 01:56 - Sync after follow-up merges
- Merged by the maintainer:
  - #187 (tooling) into this branch, its base, as `af3afc7`. This PR now carries the Makefile and `sandbox.go` changes too.
  - #188 (`x/exp` to std, closes #184) and #189 (em dashes and gofmt, closes #185) into `main`.
- One fetch of `main` and this branch. Fast-forwarded to `origin/agentrules-convention`, then `git rebase origin/main`. All 12 commits applied cleanly, and `make docs-graph` passes.
- Rule text the merges made stale:
  - `build-and-test` said some files are not gofmt-clean. `gofmt -l .` is now empty, so the rule says to keep it that way.
  - `current-libraries` gains #188 as a precedent, plus the note that the standard `maps.Keys` and `maps.Values` return iterators.
- Only `ci.yml` still has em dashes, and #186 removes them.
- Rebased #186 (`workflow-hygiene`) onto this branch with `--onto` the old head `4e2ebfa`.
