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
