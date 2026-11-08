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

### 2026-11-08 01:14 - Test and CI pass: cache key coverage, bite proof, gate
- Coverage audit found one gap among the five scenarios: the module cache key. Added `TestModuleCacheKey_SameFileAcrossRootSpellingsAndSpecifiers` to `registry_symlink_root_test.go`: one file reached through the symlinked, real and resolved root spellings and through `./lib/b.ts`, `./lib/../lib/b.ts`, `./lib/./b.ts`, `@local/lib/b.ts` yields one `@local/lib/b.ts` key (no `..`, not absolute) and one `CompiledPrograms()`.
- Bite proof against `main`'s `registry.go` (restored afterwards): fail on main: the new cache-key test, `TestNestedRelativeRequires_PackRootBehindSymlink`, `_MixedRootSpellings`, `_UnresolvedTempDir`, `TestNestedRequire_InPackLinkChainAllowed`, `TestProbedExtensionLinkOutsideRejectedForEveryEntryForm`, `_FromNestedModule`, `TestRootItselfSpecifierDoesNotProbeSiblingOfRoot`, `TestCheckUseRequiresScansLocalTypeScript` (capabilities workaround removed), and both cmd tests. Regression guards (pass on main): `TestResolveUse_LocalPackRootBehindSymlink`, `TestResolveRequire_SymlinkedRootRelativeEscapeRejected`, `_SymlinkedRootInnerLinkOutsideRejected`, `TestLoadRequire_InnerLinkInsidePackRootAllowed`, `TestProbedExtensionDanglingLinkFallsThroughToJS`, `TestNestedRequire_DirectoryLinkSegmentOutsideRejected`, `_DirectoryLinkOutsideRejectedWhenExecuted`, `_DotDotAfterDirectoryLinkReadsLexicalParent`.
- Gate green: make test, make lint (golangci-lint v2.4.0 built with Go 1.25), gofmt, make docs-graph (no diff), go test -race ./cmd/... ./runtime/js/..., no em dash, no attribution lines, SPDX year 2026.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-11-08 01:20 - Review pass: approved, suggestions
- Opus reviewer returned `VERDICT: APPROVE` with no Critical or Warning findings. It verified the symlinked root (both spellings, macOS `/var`), in-pack directory links, every outward link form including probed `.ts`/`.js` and the root-itself specifier, nested `../` requires, and the cache key (relative to the resolved root, never `..` or absolute, one compile across root spellings). Error text names only the pack-relative path, as on `main`.
- Took: pinned the extensionless key in `TestModuleCacheKey_SameFileAcrossRootSpellingsAndSpecifiers` (`./lib/b` and `@local/lib/b` give `@local/lib/b` and probe to `b.ts`), listed that test in the PR description, and added two pre-existing gotchas to `runtime.js.registry.md`: stale specs after a pack root link flip, and the separate key for an extensionless specifier.
- Declined: appending the probed extension to the key so `./a` and `./a.ts` share one spec. `main` behaves the same, and changing module identity is out of scope for #242; it belongs in its own issue.
- Declined: resolving the pack root once per `relativeToLocal` call. A probe hit now costs one `os.Stat` and one extra symlink walk of the root per unmemoised resolution (`PrepareUse` runs per evaluation). It is bounded and small, and the better fix is to memoise `PrepareUse` the way `LoadRequire` is, which is its own change.
- _Model: claude-opus-5-5, subagents: 1x sonnet (test and CI), 1x opus (review)_

### 2026-11-08 01:31 - Review re-check approved; cost correction
- The same reviewer re-checked the suggestion delta and returned `VERDICT: APPROVE`.
- Correction to the 01:20 entry: a probe hit costs up to two `os.Stat` calls plus one extra confinement, which is two symlink walks (the root and the probed file), not one. The PR description's review note and the graph node's performance note now say so. The decline stands: memoising `PrepareUse` is the fix, in its own change.
- Gate rerun on Go 1.25.0 after the delta: make test, make lint (golangci-lint v2.4.0, 0 issues), gofmt, make docs-graph, `go test -race ./cmd/... ./runtime/js/...` all pass.
- _Model: claude-opus-5-5, subagents: 1x opus (review re-check)_
