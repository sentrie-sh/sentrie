# 335-serve-test-engine

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-17 00:46 - Pick up #335
- The v0.0.5 release run (43936928299) failed in `go test ./...`: six `TestCmdTestSuite` serve tests in `cmd/serve_signal_test.go` exit with `pack requires sentrie >=0.1.0 <2.0.0, running 0.0.5`.
- Cause: the test builds the real binary with a plain `go build`. Go stamps the main module version from git, which is exactly `v0.0.5` on the tagged commit, and `version/version.go` uses it. Off a tag it is a pseudo-version that the engine check only warns about, so `main` and PR CI stayed green.
- Supersedes a claim in #334 (`333-example-pack-engine`): its journal and PR said the `>=0.1.0 <2.0.0` fixtures in `loader/pack_test.go` and `cmd/serve_signal_test.go` could stay because "the tests never compare against a release-stamped version". That is wrong for `cmd/serve_signal_test.go`, which runs a built binary. It still holds for `loader/pack_test.go`, which only parses the manifest.
- Reproduced locally with `main` at the `v0.0.5` tag: `go test ./...` fails in `cmd` only, with the same six tests.
- Plan: drop the `[engine]` block from `signalTestPack`; these tests are about signals, health and body limits, not the engine check.
- Release plan agreed with the user: merge this, then delete and re-push the `v0.0.5` tag on the fix commit. The Go proxy and sumdb already hold `v0.0.5` at `8b4718a`, so module users keep that commit; the difference is a test file only.
- Branch and worktree `worktrees/335-serve-test-engine` off local `main` (`8b4718a`), which matches origin.
- _Model: claude-opus-5-5_

### 2026-11-17 00:52 - Drop the fixture's engine block and check it at a tag
- Removed the `[engine]` block from `signalTestPack` in `cmd/serve_signal_test.go`. The fixture now has only `[schema]` and `[pack]`.
- Checked it where it fails: a throwaway clone of this branch, the change committed there unsigned (scaffolding, never pushed), and a local-only `v0.0.5` tag on that commit.
  - `go version -m` on a plain `go build` there reports `mod github.com/sentrie-sh/sentrie v0.0.5`, the same stamp the release job gets.
  - `go test -count=1 ./...` passes, all 23 packages with tests, including `cmd`.
- Did not change `loader/pack_test.go`: it only parses the manifest and never runs the engine check.
- _Model: claude-opus-5-5_
