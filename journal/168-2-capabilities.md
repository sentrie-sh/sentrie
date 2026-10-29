## Binaek Sarkar (binaek89@gmail.com)

### 14:47 - Start the capabilities layer
- Second layer of the #168 stack, cut from `168-1-nondeterminism-metadata`, which supplies the builtin call walker and the module metadata this layer builds on.
- Plan: a leaf `capabilities` package for the deployment-level JSON grant (allowed builtins, allowed `@sentrie/*` modules, whether local TypeScript modules are permitted); `--capabilities` on `exec`, `serve` and `validate`; an index-time check that fails with the offending span; paired runtime gates in builtin dispatch and in `js.Registry` `PrepareUse`/`LoadRequire`. An absent file grants everything, which is today's behaviour.
- _Model: claude-opus-5-5_

### 14:53 - Capabilities file, load-time check and runtime gates
- New leaf package `capabilities` (imports only `constants`), so `index`, `runtime`, `runtime/js` and `cmd` can all read the grant without a layering problem. Format: `version` (must be 1), `builtins`, `modules` (full `@sentrie/x` specifiers, since that is what pack authors write), `local_modules`. Pointer fields tell absent (grant all) apart from `[]` (grant none). Strict decoding: unknown fields, trailing data and bad entries are errors.
- Decided not to check names against this build's builtins and modules, as OPA does. A file written for a newer engine (for example one listing #153's `now_ms`) still loads, and a typo can only narrow the grant, which then fails loudly at load.
- A nil grant allows everything, unlike `pack.Permissions`, where nil must deny. Recorded the difference in the graph node so nobody copies one convention into the other.
- `index.CheckCapabilities` runs after `Validate` and reuses layer 1's `forEachBuiltinCall`. Uses are classified the way `js.Registry.resolveUse` classifies them (`@sentrie/...` is builtin; a quoted path or `@local/...` is local). Errors are `xerr.CapabilityError` with a span, wrapping a new `xerr.ErrNotGranted` as well as `ErrIndex`, sorted by position.
- Runtime halves: the builtin branch of `getTarget` checks `exec.grant()`, and `js.Registry.checkGrant` runs in `PrepareUse` and `LoadRequire` (not `programFor`, which #156 edits). `LoadRequire` serves every `require()`, so TypeScript cannot reach around the static check. Edits to `registry.go` and `executor.go` are one struct field each plus the two gate calls; the methods live in new `capabilities.go` files.
- CLI: `capabilitiesFlag()` and `loadAndCheckCapabilities()` in `cmd/capabilities.go`. Each of `exec`, `serve` and `validate` gains one `WithFlag`, one struct field and the load call before `NewExecutor`. That keeps the overlap with #142's `--eval-timeout` small. Rewrote the legacy headers of `exec.go` and `serve.go` on touch, as the license rule requires.
- Found: the span of a `use` statement prints as `3:3-3:0` (To column 0), so capability errors on `use` lines show a malformed end position. This predates the branch; reported rather than fixed.
- Pack `[permissions]` (#157) are not intersected with the grant yet; noted as a follow-up.
- `make test` and `make lint` pass.
- _Model: claude-opus-5-5_

### 22:35 - Review fixes (pr-203)
- `require()` in pack TypeScript is now checked at load, not first on the request that reaches it. `js.Registry.CheckUseRequires` compiles each `use` target and scans its transpiled code for static `require("...")` (esbuild emits every import that way), resolving each through `LoadRequire` so `checkGrant` applies, and follows pack-local requires transitively. `runtime.CheckModuleGrants(exec)` runs it for every `use` in sorted order. Chose a regex over transpiled output instead of a full JavaScript AST walk (goja has no visitor). It may over-match a `require("...")` inside a string literal, which fails closed, and dynamic `require(expr)` is still left to the runtime gate. An end-to-end test confirms `errors.Is(err, xerr.ErrNotGranted)` survives goja's exception wrapping.
- `cmd`: replaced `loadAndCheckCapabilities` with `newGrantedExecutor(ctx, idx, path, opts...)` (load, index check, `NewExecutor` with the grant, module scan). Each command now has one call where it had two blocks. `opts` leaves room for #219's `WithEvalTimeout`.
- Capabilities file errors: a `json.Decoder` token pre-pass (`checkTopLevelKeys`) rejects an empty file, a non-object value, unknown or differently-cased keys (exact spellings from the struct tags, so layer 3's `deterministic_js` is picked up automatically), duplicate keys, trailing data and a missing `version`. Syntax, type and key errors carry a 1-based `line:column`. `Load` now prefixes read and parse errors the same way (`<path>: capabilities: ...`).
- FORMAT.md: removed the malformed `use` span from the example (#216 fixes the parser range); said that forbidding clock and random sources also needs `local_modules: false`; documented the load-time `require` scan and the error positions.
- Inlined `joinLibFrom`.
- Not done here: the reviewer's `import ... with` and shadowed-name capability cases. They are fixed in layer 1's shared walker (`forEachBuiltinCall`), and this branch inherits that when layer 1 restacks. Cherry-picking the fix here would duplicate commits across the stack.
- Found: nested relative requires fail with "outside the packroot" when the pack root is behind a symlink (macOS temp dirs), because `ModuleSpec.Dir` is symlink-resolved but `PackRoot` is not. This predates the branch; the test resolves its temp dir, and the defect is reported.
- `make test` and `make lint` pass.
- _Model: claude-opus-5-5_

### 14:02 - Restack onto the rebased 168-1
- Carried onto 168-1 after it was rebased onto `main` (cfe249a). Code conflicts with #194 (105), both additive. In `runtime/executor.go` I kept `main`'s `constants` and `time` imports beside `capabilities`, and #194's dropped `puddle` import. In `runtime/js/registry.go` I kept #194's `resolved` LoadRequire memo and compile counter next to this layer's `grant` field.
- Checked that the memo cannot skip the grant: `checkGrant` runs in `loadRequire` on a memo miss, the grant is fixed at executor construction, and the check depends only on the module key, so a memoised resolution was already granted.
- _Model: claude-opus-5-5_

### 04:24 - Restack onto the rebased 168-1 after #205 merged
- Carried onto 168-1 after it was rebased onto `main` (1197589). #205 replaced the executor's module-binding perch with a lifetime module map and added `Executor.Close`, which conflicted in three places:
  - `runtime/executor.go`: kept `main`'s module map fields and added this layer's `capabilities` field.
  - `cmd/validate.go`: `validate` builds through `newGrantedExecutor` and then closes the executor, as #205 does, before the warnings from 168-1. `exec` and `serve` merged cleanly with #205's deferred `Close`.
  - `docs/codebase-graph/runtime.executor.md`: kept #205's `Close` entry and narrowed perch row next to this layer's `WithCapabilities` entry and `READS_FROM [[capabilities]]` row.
- A leak the merge exposed: `newGrantedExecutor` returned an error from `CheckModuleGrants` without closing the executor it had just built. It now closes it first, so a denied module grant does not leave the executor's modules loaded.
- Build, vet and `go test ./...` pass on the rebased head, and on 168-3 and 218 above it.
- _Model: claude-opus-5-5_
