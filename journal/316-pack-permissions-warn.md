## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #316
- The `fs_read` and `net` pack permissions gate nothing.
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Warn at pack load and at `validate` when `fs_read` or `net` is non-empty (the permissions are not enforced yet).
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#91 (draft, base `next`); Sentrie PR #331.
- _Model: claude-opus-5-5_

### 2026-11-11 09:46 - Warn on unenforced fs_read and net
- Reproduced: `rg` finds no reader of `Permissions.FSRead` or `Permissions.Net`; only `CheckEnvAccess` is consulted. Both lists parse and gate nothing.
- Added `printUnenforcedPermissionWarnings` in `cmd/warnings.go`, the existing `warning: ...` stderr mechanism (as `validate` does for nondeterminism). Called right after `loader.LoadPack` in `exec`, `serve`, `validate` and `bench` (the four loader paths in `cmd/`). One line per non-empty list; nil permissions and empty lists are silent. Warnings never fail the command.
- Chose `cmd/` over `loader.LoadPack` so library callers and tests stay quiet and no logger is needed.
- Test: `TestUnenforcedPermissionWarnings` in `cmd/warnings_test.go` (nil, empty, env only, fs_read, net, both).
- Graph: `docs/codebase-graph/cmd.md` updated. `make docs-graph` and `make test` pass (log `/tmp/sentrie-316-pack-permissions-warn-impl-test.log`).
- _Model: claude-sonnet-5-5_

### 2026-11-11 10:27 - Test the permission warnings at the CLI
- Added `TestPackPermissionWarningsAtCLI` in `cmd/pack_permissions_warning_test.go`: validate, exec, bench and serve (serve against an unparsable policy so it returns after load) each run against packs with fs_read, net, both, empty lists, env only and no permissions. Asserts the exact stderr warning line per permission is present or absent, and that warnings do not fail the command. Covers every decision bullet (warn at pack load and at validate; empty stays silent). No PROVISIONAL items in the implementer entry.
- Note: `net` entries must be URIs under the pack schema, so the fixtures use `https://example.com`.
- Bite proof: restoring `cmd/{bench,exec,serve,validate}.go` from main fails all 12 positive subtests (4 commands x fs_read/net/both); negatives pass either way.
- _Model: claude-sonnet-5-5_

### 2026-11-11 10:30 - Docs for #316 (website#91)
- Rewrote the permissions reference: only `env` is enforced; `fs_read` and `net` are parsed and validated, gate nothing, and trigger the stderr warning on validate, exec, serve and bench. `net` entries must be URIs. Removed the invented defaults and the "violations fail at runtime" claim.
- Also changed packfile.md, guides/deployment-and-security.md (OS-level confinement advice), troubleshooting/common-errors.md and reference/functions.md. No docs change needed in cli-reference/index.md (example only).
- Snippets run against the branch binary: complete pack file example, security page examples, guide example (validate and exec exit 0, warnings on stderr); a non-URI `net` is rejected. No pre-existing broken snippets found; untouched TOML fragments on packfile.md were not run.
- Website build passes (`/tmp/website-316-pack-permissions-warn-build.log`). Website commits 5ee99f2, 6da75d5, 7faf538, 6ef6ef6, all signed.
- _Model: claude-sonnet-5-5_
