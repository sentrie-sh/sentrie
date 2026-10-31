# 242-symlink-relative-requires

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #242
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: Nested relative `require`s in TypeScript modules fail when the pack root path goes through a symlink.
- No design choice left open by the issue; the fix must keep `@local` module path confinement intact.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:32 - Reproduce and root-cause
- Added `runtime/js/registry_symlink_root_test.go` and ran it against unchanged code. Nested chain `main -> ./a -> ./lib/b -> ../c` fails with `relative path is outside the packroot` both for an explicit `current -> releases/1` pack root and for a plain macOS `t.TempDir()` (`/var` -> `/private/var`). Mixing root spellings (registry on the link, first require from the real dir, and the reverse) fails the same way.
- Root cause: `confineToPackRoot` first ran a lexical `filepath.Rel` of the joined path against the **unresolved** `PackRoot`. A nested require joins onto the requiring module's `Dir`, which is already symlink-resolved, so the lexical check compared a resolved path with an unresolved root and failed before the resolved check ran. The resolved check alone already decides containment correctly.
- Found while writing confinement tests: extension probing in `getOrCreateModule` stats `path + ".ts"` / `".js"` **after** confinement, so a pack-local `leak.ts` symlink pointing outside the pack is read by `require('./leak')` (and `@local/leak`, `use ... from local leak`). The extension-less path does not exist, so it passes confinement, then probing follows the link. This is a real confinement bypass (fails open), pre-existing on `main`; `TestLoadRequire_ProbedExtensionLinkOutsideRejected` reproduces it.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:34 - Fix containment and confine extension probing
- `confineToPackRoot` now resolves `PackRoot` and the cleaned joined path, then decides containment only on that pair, returning the resolved path and its pack-relative form (so `relativeToLocal` stops resolving the root twice). Chose dropping the lexical check over caching a resolved root at construction: `NewRegistry` cannot return an error, and per-call resolution keeps a flipped `current` link behaving as before.
- Safety argument: `filepath.Join` cleans the path, so no `..` survives into resolution, and the resolved path is the one read, so a `root/link/../x` spelling reads `root/x`.
- Decision beyond the pickup entry: moved `.ts` / `.js` probing from `getOrCreateModule` into `relativeToLocal` and confined the probed file. `Path` keeps the probed (unresolved) name so `SourceName` and `Dir` are unchanged for valid packs. Cost: up to two stats per resolution instead of per key; `LoadRequire` memoises.
- Removed the `EvalSymlinks` workaround in `capabilities_test.go` that existed only because of this bug.
- Updated `docs/codebase-graph/runtime.js.registry.md` (containment, probing, perf note). `make docs-graph` passes.
- Verified: `go vet ./runtime/...`, `go test ./runtime/... ./lang_test/... ./cmd/...` all pass. Local `golangci-lint` panics on a toolchain mismatch (built with go1.26, file needs go1.27), unrelated to this change; left for the CI gate.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:37 - Test and CI pass
- Added `runtime/js/registry_symlink_confine_test.go`: probed-extension link-out rejected through all four entry forms (require relative, `@local`, `use local`, `use` relative) for .ts, .js, a link chain and a .ts link shadowing an in-pack .js; the same bypass from a nested module on a resolved root; dangling .ts falling through to .js; directory-link and directory-link-chain segments leaving the pack from a nested require (static and executed); `..` after a directory link reading the lexical parent; an in-pack link chain still followed.
- Added `cmd/symlink_pack_test.go`: `exec` and `validate` on a pack behind `current -> releases/1` with a nested TS chain, and `exec` refusing a `helper.ts` linked outside the pack (probed and explicit extension).
- Bite proof against `main`'s `registry.go`: the nested, probing and chain tests and both cmd tests fail; the directory-link, dangling-link and `..` tests are regression guards that already passed on `main` (it was fail-closed there). The nested-module probe test uses a resolved root so only probed-file confinement can refuse it.
- Gates: make test, lint (go1.25.0, 0 issues), gofmt, docs-graph, race on cmd and runtime/js, no em dash all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-10-31 14:42 - Address review: probed read path, root-itself probe
- Correction to the earlier "Fix containment and confine extension probing" entry and the `confineToPackRoot` comment: "the resolved path is the one read" is not true for a probed specifier. The resolved path is read, except a probed `.ts`/`.js`, which is confined and then read by its own name (`os.ReadFile` follows the link again). Reworded the comment and PR description; code unchanged on purpose, since storing the link target would change `SourceName`.
- Second fail-open closed by the same change, now documented and tested: a specifier resolving to the pack root itself (`.`, `./`, `@local/`, `@local/.`, `../<rootname>`, `use local` with no path) probed `<packroot>.ts`, a file next to the root, which `main` read unchecked. Added `TestRootItselfSpecifierDoesNotProbeSiblingOfRoot` to `registry_symlink_confine_test.go`; with `main`'s `registry.go` swapped in it fails (`.` resolves with no error), restored afterwards with `git diff main --stat` identical.
- Took the suggestions: dropped the duplicate `TestLoadRequire_ProbedExtensionLinkOutsideRejected` (covered for every entry form by the test agent's file), named the test agent's files in the PR, and said "OS-separated relative form" in the comment.
- Possible follow-up, out of scope: the `mods` cache is keyed by the extensionless `@local/...` key, so after a pack root link flip in a long-lived `serve` a cached module keeps pointing at the old release's resolved path.
- _Model: claude-opus-5-5, role: implementer subagent_
