# GitHub Actions

Workflows live under `.github/workflows/`: `ci.yml` (`lint`, `docs-graph`, `test`), `cla-check.yml`, `deps-policy-check.yml`, `protected-files-maintainers.yml`, and `release.yml` (tag pushes only). Pin `uses:` to a major tag (`@v4`, `@v5`), not a patch or a commit SHA. A tool version passed as an input (such as golangci-lint's `version: v2.4.0`) is pinned exactly.

Changes to these files are `[Workflow]` PRs ([PR conventions](pr-conventions.md)), and only maintainers may make them (`protected-files-maintainers.yml` guards `.github/**`).

## Run CI only when it tells us something

Sentrie is a public repository, so standard GitHub-hosted runners cost nothing. Minutes are not the constraint; speed and noise are. Every unnecessary run occupies the shared concurrent-job limit, delays the run that matters, and adds red or yellow checks and bot comments to PRs that nobody needs to read. The local toolchain ([Build and test](build-and-test.md)) proves a change; CI is the gate before merge, not the development loop. So:

- **Draft PRs run nothing.** Every PR-triggered job is gated on `!github.event.pull_request.draft`, and the `ready_for_review` event starts it. Marking a PR ready for review is what runs CI; later pushes to a ready PR run it again.
- **One workflow per job that needs its own path filter.** GitHub has no per-job path filter, so a job whose inputs differ from its neighbours' lives in its own file with its own `paths` filter. A filter with no matching changed file creates no run at all.
  - On `pull_request`, a filter is matched against the whole PR diff, not the commit just pushed. A journal-only push to a ready PR that changes code still runs every workflow the PR's code matches. Batch journal and prose commits into the next code push, or push them while the PR is a draft.
- **Prose-only paths are ignored.** Go CI skips changes that touch only `**.md`, `journal/**`, `.agentrules/**`, `.cursor/**` and `LICENSE*`, with `docs/codebase-graph/**` re-included for the `docs-graph` job because it validates those nodes. When you add a new prose-only location, add it to the filter.
- **On demand:** every workflow has `workflow_dispatch`, so `gh workflow run <file> --ref <branch>` runs one check without the others.
- **Superseded runs are cancelled.** `concurrency` with `cancel-in-progress: true`, grouped on workflow and ref, cancels the in-progress run when a new commit lands on the same PR. Never cancel `release.yml` runs.
- **Every job sets `timeout-minutes`.** A hung job otherwise runs for six hours. Size the cap from recent run times with headroom, not from a guess.
- **No new always-on workflows.** Anything added under `.github/workflows/` must justify each trigger. Prefer `workflow_dispatch` or a schedule over `push`, filter by path, gate on drafts, set `timeout-minutes`, and skip jobs whose inputs are absent rather than failing.
- `main` has no branch protection and no required status checks, which is what makes a skipped run safe. If that ever changes, the required checks must be ones that run on every PR, or docs-only and draft PRs will be unmergeable.

## Caches

A PR can restore only caches saved on its own ref or on `main`, never another PR's. Keep caches layered, each keyed on what invalidates it:

- **Go modules** (`GOMODCACHE`), keyed on `go.sum`. `actions/cache` saves only on an exact miss, so a PR saves one only when its dependencies differ from `main`'s.
- **Go build cache** (`GOCACHE`: compiled packages and test results), restored from the most specific level available: this ref's own earlier run with the same dependencies, then `main`'s with the same dependencies, then the latest for the Go version. `main` saves; a branch saves only when it did not restore its own, and only from a passing run.
- **Tool caches** (such as golangci-lint's), keyed on the tool's version, never folded into the build cache key.

Rules for changing them:

- A stale cache is never a correctness risk: Go keys every build cache entry on a hash of its exact inputs, so a fallback restore only saves time. Never add a step that trusts cached output without Go's own checks.
- Caches are ephemeral. Old PR entries are left to GitHub's least-recently-used eviction; don't add cleanup jobs for them.
- Once `ci.yml` uses layered caches, turn `setup-go`'s own `cache` off. Its single `go.sum`-keyed entry cannot be layered and duplicates the build cache.

## Habits that stop runs repeating

- **Mark one stack layer ready at a time.** Only the bottom open layer of a stack is ready for review ([Stacked PRs](stacked-prs.md)). A layer marked ready early runs CI against its parent branch, then again against `main` when it becomes next, so the first run buys nothing. Before marking several PRs ready at once, even when asked to, say how many runs that starts.
- **Check past durations before a long dispatch.** Before dispatching a workflow that takes more than a few minutes (the `test` job has a 35-minute cap), look at its recent run times (`gh run list --workflow <file>`, then the job's step times). If it ran close to its timeout, find out why before starting another run.
- **Mark a PR ready instead of dispatching its workflows.** Readiness runs every workflow whose filter matches, and GitHub cannot merge a draft, so a PR has to be marked ready before it merges anyway. Dispatching `ci.yml` on a draft and then marking it ready runs the same commit twice. Dispatch by hand only for a workflow readiness does not run.

Related: [Build and test](build-and-test.md), [Stacked PRs](stacked-prs.md), [PR conventions](pr-conventions.md)
