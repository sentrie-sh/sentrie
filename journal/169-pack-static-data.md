# 169-pack-static-data

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #169
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: A policy can only see per-request facts; there is no way to ship static reference data with a pack.
- Decision (default proposed to the user on 2026-10-31, not overridden): a `[data]` table in the pack manifest maps names to JSON files loaded once at pack load, optionally validated against a shape, exposed read-only. Access syntax settled by the implementer against existing namespace resolution and recorded here.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:37 - Access syntax and core implementation
- Access syntax options considered:
  - A reserved namespace binding `data.roles`: implicit dependency, and it takes the name `data` from every policy, where a fact or let named `data` is common, so existing packs could break.
  - A new lexer keyword `data`: explicit, but reserving the word breaks every identifier and field named `data` across the language.
  - Reusing `use { roles } from @data`: explicit, but `use` aliases are TypeScript module bindings, so the index, capability checks, nondeterminism checks and module loading would all need a special case, and `alias.fn()` is the module call form.
  - Chosen: a contextual `data` declaration in the policy header, `data <name> (':' <type>)? ('as' <alias>)?`, mirroring `fact`. `data` is recognised only at the head of a policy statement, where a bare identifier was always a parse error, so no identifier anywhere else changes meaning. The dependency is explicit per policy, the alias shares the policy identifier space (conflicts with facts, lets, rules and use aliases are index errors), and `as` resolves a clash with a fact of the same name.
- Shape validation: the optional type sits on the declaration, like a fact's, so it reuses the existing type references and shape visibility rules instead of an FQN string in TOML. It is checked once per typed declaration in `NewExecutor` (`runtime/pack_data.go`), so `sentrie validate`, `exec` and `serve` all fail before any request, and reads do no per-evaluation validation.
- Manifest: `[data]` maps an identifier-shaped name to a path (schema `propertyNames` pattern). `loader.loadPackData` runs at the end of `LoadPack`: rejects absolute paths and non-local paths (`filepath.IsLocal`), opens through `os.Root` so symlinks resolving outside the pack root fail, requires a regular file, bounds each file at 16 MiB (stat plus a limited read), rejects keyword names (unreachable from `data <name>`), and wraps every error as `data "<name>" (<file>): ...`.
- Decoding mirrors facts: `json.NewDecoder` into Go values then `box.FromBoundaryAny`, so #166's number change applies to both. Trailing content after the document is an error (facts ignore it).
- Values are boxed once at pack load into `PackFile.DataValues` and shared. Immutability: no policy operation mutates a value, and values cross into TypeScript only as copies (`box.TryToBoundaryAny` then `normalizeBoundaryForJS`).
- Derive purity: data is denied inside a derive body, paired: static `checkDeriveIdentifier` and runtime `evalIdent` both give "pack data is not available inside a derive; pass it as an argument". Reason: `data` is declared per policy, and derive bodies see only parameters, lets and derives; passing the value as an argument keeps a derive a function of its inputs.
- Static index: `checkPackData` (in `Validate`) requires every declaration to name a loaded manifest entry; the rule kind-check scope binds data aliases with the loaded value's kind, so `x is roles` and builtin kind checks see it as a value.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:47 - Tests, graph and docs
- Tests added: `loader/data_test.go` (path confinement, symlinks in and out of the root, size cap via a sparse file, malformed and trailing JSON, keyword names, schema rejections), `parser/data_test.go`, `index/pack_data_test.go` (manifest check, clashes, phase, derive refusal, static kind), `runtime/pack_data_test.go` (end-to-end reads, typed validation at executor build, immutability against a mutating TypeScript module, runtime derive guard). Runtime tests pass under `-race`.
- Problem: `dict[number]` did not reject a dict of lists in the mismatch test, because dict type validation checks only that the value is a dict and its constraints, not element types (this predates the branch). Switched the test to `list[string]`. Not filed here; the parent may want an issue for dict element validation.
- Problem: a derive declared in a policy must follow a body statement (`latePolicyHeaderErr("derive")` fires in the header phases). Data declarations move the phase to facts, the same as facts do, so the tests add a `let` before the derive.
- No `example_pack` or `lang_test` change: both validate sources with an index built without a pack (`TestValidateAllRepoPacksUnchanged`, `lang_test`), and such an index cannot satisfy a `data` declaration by design.
- Graph: new nodes `parser.data`, `index.pack_data`, `runtime.pack_data`; updated pack, loader, ast, grammar, parser.policy, index.policy, index.validate, index.builtin_kind, index.derive_purity, runtime.eval_ident, runtime.executor (NewExecutor now has a real error return, for pack data). `make docs-graph` passes. `make lint` panics locally (golangci-lint built with go1.26, toolchain go1.27); this is environmental.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-11-01 02:31 - End-to-end tests and CI gate
- Tests added: `cmd/pack_data_test.go` (exec on typed, untyped and aliased data with nested access and list membership; `data` still a fact and field name beside a data declaration; validate success and each failure: type mismatch, undeclared manifest name, derive refusal, missing file, absolute path, `..` escape, directory, malformed JSON, trailing content, symlink outside the root, file one byte over the cap; undeclared name is not visible; data passed into a derive as an argument), `api/pack_data_test.go` (decision endpoint reads data across requests), `runtime/pack_data_concurrency_test.go` (parallel reads and module copies under -race), `index/pack_data_phase_test.go` (a data declaration is a fact-phase statement; late data after the body is rejected).
- Bite proof: every production file touched fails the new tests when reverted to main, except `index/policy_stmt.go` and `index/builtin_kind.go`, which are bitten by the index tests (`TestPackDataIsAFactPhaseStatement`, `TestPackDataBindsLoadedKindStatically`). One interrupted bite run left `index/validate.go` reverted; it was restored from backup and the diffstat re-verified.
- Observation (pre-existing, not this branch): `Array.prototype.push` on an array passed to a TypeScript module has no effect on the array, for facts as well as data, so the `push` line in `TestPackDataImmutableFromModules` proves nothing; the index assignments, added key and delete in the same test do.
- No `lang_test` fixture (0057): that harness builds an index without a pack, which cannot satisfy `data`. Gate: make test, lint (0 issues), gofmt, docs-graph, -race on loader, parser, index, cmd, api, pack, ast and the runtime pack-data tests all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_
