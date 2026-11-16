# 333-example-pack-engine

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-17 00:05 - Pick up #333
- Found during the v0.0.5 pre-release sanity pass on `main` at `47a77ac`. `make test`, `make lint` and the race runs were all green, including `./runtime/...` with `-race`, which CI skips.
- A binary built with `-X main.version=0.0.5` refuses `example_pack`: `pack requires sentrie >=0.1.0 <2.0.0, running 0.0.5`. The check came in with #330 (Closes #157). Dev builds only warn, so CI never sees this.
- `sentrie init` writes no `[engine]` block, so new packs are unaffected.
- Scope: the example pack's manifest only. The website `next` docs show the same `0.1.0` constraint in `packfile.md`, `cli-reference/index.md` and `cli-reference/serve.md`. That fix belongs in `sentrie-sh/website`.
- Test fixtures `loader/pack_test.go` and `cmd/serve_signal_test.go` also use `>=0.1.0 <2.0.0`. They are left alone: the tests never compare against a release-stamped version.
- Branch and worktree `worktrees/333-example-pack-engine` off local `main`; `gh api` confirmed local `main` matches origin.
- _Model: claude-opus-5-5_
