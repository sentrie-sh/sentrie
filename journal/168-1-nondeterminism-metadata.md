## Binaek Sarkar (binaek89@gmail.com)

### 14:40 - Plan the #168 stack
- #168 asks for operator capabilities (a deployment-level grant of builtins, `@sentrie/*` modules and local TypeScript modules) plus nondeterminism metadata, and optionally deterministic `Date.now`/`Math.random` shims. Well over 800 hand-written lines, so it ships as a stack:
  - `168-1-nondeterminism-metadata` (this branch): a `Nondeterministic` flag on `builtins.Decl`, a nondeterministic-export table for the Go and TypeScript `@sentrie/*` modules, derive rejection of nondeterministic builtins (static and runtime, paired per the evalDerive rule) and `sentrie validate` warnings.
  - `168-2-capabilities`: the capabilities JSON file, `--capabilities` on `exec`, `serve` and `validate`, the index-time check with spans and the paired runtime gates.
  - `168-3-deterministic-js`: optional seeded `Date.now`/`Math.random` shims behind a capabilities flag.
- Nondeterministic sources found: builtin `now()`; `@sentrie/time.now`; `@sentrie/uuid.v4`, `v6`, `v7`; `@sentrie/js.now` (`Date.now`) and `@sentrie/js.random` (`Math.random`). `@sentrie/jwt` does not read the clock (no `exp` checking), and no other Go module calls `time.Now` or a random source.
- Other agents touch `builtins/` (#153 adds `now_ms()`), `runtime/js/registry.go` `programFor` (#156) and `runtime/executor.go` binding code, so the JS metadata goes in a new `runtime/js/nondeterminism.go` instead of changing `RegisterGoBuiltin`.
- _Model: claude-opus-5-5_

### 14:47 - Nondeterminism metadata, derive gate and validate warnings
- Added `Decl.Nondeterministic` next to `DeriveSafe` instead of flipping `now` to `DeriveSafe: false`. The two flags answer different questions (within one execution, across executions), and #153's `now_ms` then needs only the one flag. `(*Decl).PermittedInDerive()` combines them, and both `index/derive_purity.go` (through `IsDeriveSafe`) and the runtime guard in `getTarget` read it, so the static and runtime gates cannot drift.
- Breaking: `now()` inside a derive is now refused. #168 asks for it, and memoized derive calls are cached across executions, so a clock inside a derive can go stale. Replaced `TestExecRuleDeriveUsesNowBuiltin` with a validation-failure test, and added a runtime-only guard test.
- Put the JavaScript export metadata in a new `runtime/js/nondeterminism.go` table rather than extending `RegisterGoBuiltin`, which other branches (#138, #156) also edit. `index` must not import `runtime/js` (goja and esbuild), so `NondeterminismWarnings` takes the lookup as a function argument and `cmd` passes `js.IsNondeterministicExport`. A runtime test keeps the table honest against what `NewExecutor` registers.
- Wrote `forEachBuiltinCall` as a separate walker instead of reusing `checkBuiltinCalls`. That check only walks rules and derives, and a capability check must also cover fact defaults, policy lets and export attachments. Layer 2 reuses the walker.
- Warnings go to stderr as `warning: <span>: <message>` and never fail `validate`. They are reported per `use` import, because the parser requires every `use` to name its imports.
- Found while reading: `use { x } from "@sentrie/js"` (quoted) is resolved by the registry as a pack-relative file, not the builtin. Several `lang_test` fixtures use that form; they only parse and index, so nothing fails. Left alone and reported.
- `make test` and `make lint` (golangci-lint v2.4.0 on go1.25.0) pass.
- _Model: claude-opus-5-5_

### 22:34 - Review fixes (pr-200)
- Walker gap: `visitBuiltinCalls` fell through to `forEachDeriveExprChild`, which has no case for `ImportClause`/`WithClause` and returns an error that was ignored, so `import decision ... with t as now()` was never visited. Added both cases to the walker itself rather than to `forEachDeriveExprChild`, which derive cycle detection and `checkBuiltinCalls` share (#218 is reworking the latter).
- Shadowing parity: `getTarget` never resolves a call through a local binding (derive, then builtin, then module alias), but the walker used `isBuiltinCall`, which lets locals shadow. So `(now) => { yield now() }` got no warning even though it calls the builtin at runtime. Chose to make the static side match the runtime: a new `runtimeBuiltinCallee` (derive, then builtin, no locals). Left `isBuiltinCall` alone; it is pre-existing and #218's area. Applied the same parity in `checkDeriveCall`: a builtin not `PermittedInDerive` is refused even when a parameter of the same name is in scope, exactly as the runtime guard refuses it.
- Local TypeScript: every `use` of a pack-local module now gets one warning that its code can read `Date.now()`/`Math.random()`. Chose this minimum over scanning the transpiled code, which would pull `runtime/js` into the warning pass; layer 2 adds a load-time `require` scan for the grant.
- Noted in `runtime/js/nondeterminism.go` that JWT `exp`/`nbf` exports (#197, #202) belong in the table once they land.
- Signing refused the first commit attempts ("agent refused operation"), so the work waited uncommitted until signing came back. A signing probe then created an empty local commit by mistake; I removed it with a soft reset before it was pushed or built on.
- `make test` and `make lint` pass.
- _Model: claude-opus-5-5_

### 11:34 - Restack onto main after #193 and #195 merged
- Rebased onto `main` (610352e). One conflict, in the Dependencies Risk list of `docs/codebase-graph/runtime.js.builtins.md`: kept #111's two throw bullets and this layer's wider non-determinism bullet (`uuid`, `time.now`, `js.now`, `js.random`).
- _Model: claude-opus-5-5_

### 14:01 - Restack onto main after #194 and the other bottom layers merged
- Rebased onto `main` (cfe249a). One conflict, in `docs/codebase-graph/runtime.js.registry.md`: #194's `CompiledPrograms` entry and this layer's nondeterminism-metadata entry are separate signatures, so I kept both, each with its own Side Effects and Exceptions lines.
- _Model: claude-opus-5-5_
