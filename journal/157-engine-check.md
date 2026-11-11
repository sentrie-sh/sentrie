## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #157
- The pack's `[engine]` version constraint is never checked.
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Escape-hatch flag `--skip-engine-check` on `exec`, `serve` and `validate`.
  - A dev or unversioned build passes with a warning.
  - Prereleases use STRICT semver: `0.0.5-rc1` fails `>=0.0.5` (user overrode the recommendation).
  - A mismatch under `validate` is an error.
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#90 (draft, base `next`); Sentrie PR #330.
- _Model: claude-opus-5-5_

### 2026-11-11 04:01 - Implement the engine check
- New `cmd/engine.go`: `checkEngine` runs after `LoadPack` in `exec`, `serve` and `validate`; flag `--skip-engine-check`. `main` hands the version over with `cmd.SetEngineVersion`. Uses Masterminds/semver already in go.mod, so no dependency change. Strict semver comes free from `Constraints.Check`.
- Found while testing: a binary built from a checkout carries a Go pseudo-version (`0.0.5-0.2026...-hash+dirty`), which strict semver would reject for `>=0.0.5`, breaking local dev. Treated as a dev build (warning, pass). PROVISIONAL (user AFK), needs confirmation: this rule; the error text suffix "(use --skip-engine-check to load it anyway)"; an unparseable binary version warns rather than fails; `bench` is not checked (not in the decision).
- Added `cmd/engine_test.go` (table test). Updated `docs/codebase-graph/cmd.md` and `main.md`.
- `make test` passes (log /tmp/sentrie-157-engine-check-impl-test.log).
- _Model: claude-sonnet-5-5_

### 2026-11-11 09:46 - Test the engine check end to end
- Extended `cmd/engine_test.go` (kept the implementer's `TestCheckEngine`): `TestEngineCheckWiredIntoExecServeValidate` drives the real CLI for `exec` and `validate` (satisfied, unsatisfied with the "(use --skip-engine-check to load it anyway)" suffix, strict `0.0.5-rc1`, `--skip-engine-check`, dev build, pseudo-version and unparseable version warn and pass) and for `serve` (blocks on success, so a broken policy is used: engine error first, skip lets the load error through). `TestBenchDoesNotCheckEngine` pins that `bench` is not checked (PROVISIONAL decision).
- Bite proof: `main`'s `exec.go`, `serve.go`, `validate.go` each make their wiring subtests fail; restored with `git checkout HEAD --`.
- Gate: make test, make lint (0 issues), make docs-graph, race on cmd and root all pass.
- _Model: claude-sonnet-5-5_

### 2026-11-11 09:49 - Docs for #157 (website#90)
- Website branch `157-engine-check`: documented the engine check on `packfile.md` (`[engine]` section and validation rule), the `--skip-engine-check` flag on `exec`, `serve`, `validate` and the CLI index, and a `common-errors.md` row. Behaviour taken from the implementation as it stands, including the PROVISIONAL items (pseudo-version is a dev build, unparseable version warns, `bench` unchecked).
- Verified against binaries built from this branch: ldflags `main.version=0.0.4` gives the mismatch error and exit 1 on exec, serve and validate, `--skip-engine-check` loads it; `0.0.5-rc1` fails `>=0.0.5`; `-buildvcs=false` shows the dev warning; a plain build shows the source-build warning.
- Found: pack snippets in `cli-reference/serve.md` and `cli-reference/index.md` use a pre-schema layout and `[engines]`, and fail `validate`. Left alone per the snippet rule and listed in the website PR description.
- `npm run build` passes (/tmp/website-157-engine-check-build.log). Website commits 4e83060, 2a63e45, 6ec2a4d, all signed.
- _Model: claude-sonnet-5-5_

### 2026-11-11 10:30 - Round 2 review fixes
- Fixed `pseudoVersionPrerelease` to also match Go's no-tag pseudo-version `v0.0.0-<time>-<hash>` (such a binary used to fail instead of warn).
- Strict semver per the user: the constraint is copied with `IncludePrerelease = true` before `Check`, so `0.0.5-rc1` fails `>=0.0.5` and passes `>=0.0.4` (Masterminds otherwise drops prereleases).
- Checked `err != nil` before `isVCSBuildVersion` for readability.
- Tests added to `TestCheckEngine` and the wiring test (rc1 vs `>=0.0.4`, no-tag pseudo-version); they failed before the fix (8 subtests) and pass now. Verified on ldflags builds: rc1 passes `>=0.0.4`, fails `>=0.0.5`; the no-tag pseudo-version warns.
- Website `packfile.md` and `common-errors.md` reworded; graph `cmd.md` updated. Gate: make test, lint, docs-graph, race on ./cmd, gofmt, website build all pass.
- _Model: claude-sonnet-5-5_
