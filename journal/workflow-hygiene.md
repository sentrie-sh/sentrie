# workflow-hygiene

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-13 16:14 - Bring the workflows in line with the GitHub Actions rule
- Follow-up from `agentrules-convention` (#181), whose PR body listed every gap between `.agentrules/github-actions.md` and `.github/workflows/`.
- Branched off `agentrules-convention`, not `main`, because this PR updates rule files that #181 adds. Worktree: `worktrees/workflow-hygiene`.
- Sized the timeouts from recent run history (`gh run view <id> --json jobs`): `lint` about 1 minute, `test` about 3, `docs-graph` under 30 seconds, the CLA, dependency and protected-files gates about 5 seconds each, `release` about 3 minutes, `brew-release` about 10 seconds.
  - Caps: `lint` 10, `test` 35 (unchanged), `docs-graph` 5, CLA and protected-files gates 5, dependency gate 10, `release` 30, `brew-release` 10.
- `docs-graph` moved to its own `docs-graph.yml`, so it can re-include `docs/codebase-graph/**` in a filter that otherwise drops prose. A graph-only PR now runs only the graph check.
- `ci.yml` and `docs-graph.yml`:
  - Filter on `**` minus `**.md`, `journal/**`, `.agentrules/**`, `.cursor/**` and `LICENSE`.
  - Listen for `ready_for_review`, gate every job on not-draft, and add `workflow_dispatch`.
  - Group concurrency per PR. `cancel-in-progress` is true only for `pull_request`, so `main` runs still finish and upload coverage.
- Path lists are repeated instead of using a YAML anchor. Anchors work in Actions now, but the repetition is easier to read.
- Layered caches in `test`: `setup-go` cache off, `actions/cache` for modules, and restore/save split for the build cache.
  - Restore order: own ref, then `main`, then the Go version.
  - Save on `main` always, and on a branch only when it did not restore its own entry.
  - The `gobuild-test-` prefix keeps any future job's build cache separate.
- PR gates (`cla-check`, `deps-policy-check`, `protected-files-maintainers`): not-draft gate, concurrency per PR, timeouts.
  - `deps-policy-check` now triggers only on `go.mod`, `go.sum` or its own file. It already did nothing for other PRs, but it still started a job.
- Conflict found and resolved compatibly: `deps-policy-check.yml` greps the PR body for `## Dependencies`, but the section list in `pr-conventions` (and #181's new template) says `## Dependency changes`.
  - #75 and #58, the last dependency PRs, used `## Dependencies`, which is why the check passed for them.
  - The check now accepts either heading. I tested the regex and `awk` range locally against four sample bodies.
- `protected-files-maintainers.yml` now also guards `AGENTS.md`, `.agentrules/**` and `.cursor/**`, and its comment lists them. Agent rules steer every agent's work, so they get the same review as workflows.
- Replaced the two em dashes in `ci.yml` step names.
- `release.yml`: timeouts only. It has no dispatch trigger (it reads its tag) and no concurrency group (never cancel a release).
- Validation: every workflow parses with Ruby's YAML loader. `actionlint` isn't installed, and I didn't download it without asking.
- The new triggers only run once this PR is ready for review, since drafts now run nothing. `workflow_dispatch` needs the workflow on the default branch, so it can't be tried from this branch either.
- Updated `.agentrules/github-actions.md` from target to description (workflow list, dispatch exception, cache keys, measured durations), the `deps-policy-check` note in `pr-conventions`, and the `AGENTS.md` keywords.
