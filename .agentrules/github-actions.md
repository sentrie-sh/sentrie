# GitHub Actions

Workflows live under `.github/workflows/`:

- `ci.yml`: `lint` and `test` (race split, coverage, Codecov upload).
- `docs-graph.yml`: `make docs-graph`.
- `cla-check.yml`, `deps-policy-check.yml`, `protected-files-maintainers.yml`: PR gates.
- `release.yml`: GoReleaser on a pushed `v*.*.*` tag.

Pin `uses:` to a major tag (`@v4`, `@v5`), not a patch or a commit SHA. A tool version passed as an input (such as golangci-lint's `version: v2.4.0`) is pinned exactly.

Changes to these files are `[Workflow]` PRs ([PR conventions](pr-conventions.md)), and only maintainers may make them (`protected-files-maintainers.yml` guards `.github/**`, `AGENTS.md`, `.agentrules/**` and `.cursor/**`).

## Run CI only when it tells us something

Sentrie is a public repository, so standard GitHub-hosted runners cost nothing. Minutes are not the constraint; speed and noise are. Every unnecessary run occupies the shared concurrent-job limit, delays the run that matters, and adds red or yellow checks and bot comments to PRs that nobody needs to read. The local toolchain ([Build and test](build-and-test.md)) proves a change; CI is the gate before merge, not the development loop. So:

- **Draft PRs run nothing.** Every PR-triggered job is gated on `!github.event.pull_request.draft`, and every PR workflow listens for `ready_for_review`. Marking a PR ready for review is what runs CI and the gates; later pushes to a ready PR run them again.
- **One workflow per job that needs its own path filter.** GitHub has no per-job path filter, so a job whose inputs differ from its neighbours' lives in its own file with its own `paths` filter. That is why `docs-graph` has its own file, and why `deps-policy-check.yml` runs only when `go.mod` or `go.sum` changes. A filter with no matching changed file creates no run at all.
  - On `pull_request`, a filter is matched against the whole PR diff, not the commit just pushed. A journal-only push to a ready PR that changes code still runs every workflow the PR's code matches. Batch journal and prose commits into the next code push, or push them while the PR is a draft.
- **Prose-only paths are ignored.** `ci.yml` and `docs-graph.yml` filter on `**` minus `**.md`, `journal/**`, `.agentrules/**`, `.cursor/**` and `LICENSE`. `docs-graph.yml` re-includes `docs/codebase-graph/**`, because it validates those nodes, so a graph-only change runs it and nothing else. Patterns apply in order, and the last match wins. When you add a new prose-only location, add it to both filters.
- **On demand:** `ci.yml` and `docs-graph.yml` have `workflow_dispatch`, so `gh workflow run <file> --ref <branch>` runs one check without the others. The PR gates read the pull request they run on, and `release.yml` reads its tag, so they have no dispatch trigger.
- **Superseded runs are cancelled.** `concurrency`, grouped on workflow and PR number, cancels the in-progress run when a new commit lands on the same PR. Runs on `main` are not cancelled (`cancel-in-progress` is true only for `pull_request`), so every merge still uploads its coverage. `release.yml` has no concurrency group and must never be cancelled.
- **Every job sets `timeout-minutes`.** A hung job otherwise runs for six hours. Size the cap from recent run times with headroom, not from a guess. In October 2026, `lint` took about 1 minute (cap 10), `test` about 3 (cap 35, for the race-split steps), `docs-graph` under 1 (cap 5), the PR gates seconds (caps 5 to 10), and `release` about 3 (cap 30).
- **No new always-on workflows.** Anything added under `.github/workflows/` must justify each trigger. Prefer `workflow_dispatch` or a schedule over `push`, filter by path, gate on drafts, set `timeout-minutes`, and skip jobs whose inputs are absent rather than failing.
- `main` has no branch protection and no required status checks, which is what makes a skipped run safe. If that ever changes, the required checks must be ones that run on every PR, or docs-only and draft PRs will be unmergeable.

## Caches

A PR can restore only caches saved on its own ref or on `main`, never another PR's. `ci.yml`'s `test` job keeps layered caches, each keyed on what invalidates it:

- **Go modules** (`~/go/pkg/mod`), keyed on Go version and `go.sum`, through `actions/cache`. It saves only on an exact miss, so a PR saves one only when its dependencies differ from `main`'s. `docs-graph.yml` shares the same key.
- **Go build cache** (`~/.cache/go-build`: compiled packages and test results), keyed `gobuild-test-<os>-go<version>-<go.sum hash>-<ref>-<run id>`. The restore takes the most specific match: this ref's own earlier run with the same dependencies, then `main`'s with the same dependencies, then the latest for the Go version. `main` saves on every passing run. A branch saves only from a passing run, and only when it did not restore its own entry (once per dependency set).
- **golangci-lint** keeps its own cache through `golangci-lint-action`, keyed on the linter version.

Rules for changing them:

- A stale cache is never a correctness risk: Go keys every build cache entry on a hash of its exact inputs, so a fallback restore only saves time. Never add a step that trusts cached output without Go's own checks.
- Caches are ephemeral. Old PR entries are left to GitHub's least-recently-used eviction; don't add cleanup jobs for them.
- Keep `setup-go`'s own `cache` off. Its single `go.sum`-keyed entry cannot be layered and duplicates the build cache.
- A job that compiles a different set of packages gets its own build-cache prefix (like `gobuild-test-`), so two jobs never race to save the same key.
- A workflow that adds a new cached tool or download gives it its own cache keyed on that tool's version, not folded into the build cache key.

## Habits that stop runs repeating

- **Mark one stack layer ready at a time.** Only the bottom open layer of a stack is ready for review ([Stacked PRs](stacked-prs.md)). A layer marked ready early runs CI against its parent branch, then again against `main` when it becomes next, so the first run buys nothing. Before marking several PRs ready at once, even when asked to, say how many runs that starts.
- **Check past durations before a long dispatch.** Before dispatching a workflow that takes more than a few minutes (the `test` job has a 35-minute cap), look at its recent run times (`gh run list --workflow <file>`, then the job's step times). If it ran close to its timeout, find out why before starting another run.
- **Mark a PR ready instead of dispatching its workflows.** Readiness runs every workflow whose filter matches, and GitHub cannot merge a draft, so a PR has to be marked ready before it merges anyway. Dispatching `ci.yml` on a draft and then marking it ready runs the same commit twice. Dispatch by hand only for a workflow readiness does not run.

Related: [Build and test](build-and-test.md), [Stacked PRs](stacked-prs.md), [PR conventions](pr-conventions.md)
