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
