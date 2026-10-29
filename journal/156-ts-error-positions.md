## Binaek Sarkar (binaek89@gmail.com)

### 14:40 - Pick up #156 and check what goja does natively
- Branch `156-ts-error-positions` and its worktree cut from local `main` (b35fe4b). No open PRs touch `runtime/js`.
- Read the `runtime.js`, `runtime.js.registry`, `runtime.js.tscompile` and `runtime.modules` graph nodes, then `registry.go`, `tscompile.go` and `alias_runtime.go`.
- goja already maps positions through a source map: the parser reads a trailing `//# sourceMappingURL=` comment (a `data:` URL is decoded inline), sets it on the `file.File`, and every `Position()` lookup (stack frames, `Exception.Error()`, compile-phase `CompilerSyntaxError`) goes through it. Its `extractSourceMapLine` skips trailing lines that are empty or exactly `})`, which is the shape of a CommonJS wrapper.
- Sandbox experiment: esbuild `Transform` with `Banner` / `Footer` set to the factory wrapper and `SourceMapInline` emits the wrapper and the map together, with the map already offset for the banner line. goja then reported a throw on line 4 of a `.ts` file as `@local/lib/mod.ts:4:8`. So no frame rewriting and no new dependency are needed for runtime exceptions.
- The gap: goja **parse** errors are built from the parser's error list before the source map is attached, so they carry positions in the transpiled output and no `File`. Mapping them by hand would need `github.com/go-sourcemap/sourcemap` as a direct dependency. esbuild has already parsed the source and downlevelled it to ES2019, so a goja parse error on its output is rare. Decision: no new dependency; such errors name the module and say the position is in the transpiled output. Flag in Review notes.
- _Model: claude-opus-5-5_

### 14:43 - Map positions, wrap errors, stop caching failures
- `tscompile.go`: the factory wrapper moved out of `WrapAsIIFE` into esbuild's `Banner` / `Footer`, with `SourceMapInline` and `Sourcefile` set to the base name of the module's source name. esbuild offsets the map for the banner line itself, which the old same-line `{;` prefix never allowed. `WrapAsIIFE`, the `SourceMap` struct and `ModuleSpec.TranspiledMap` are removed: goja now consumes the map, so our own parse of it had no reader. `TranspileResult.Map` went with them.
- Every esbuild error is now returned (joined), each a new `*SourceError` (`errors.go`) formatted `file:line:col: message`. esbuild still stops at the first syntax error, so the multi-error test uses duplicate declarations, which it collects.
- `compileError` wraps goja errors: a compile-phase `CompilerSyntaxError` has a `File` whose `Position` goes through the source map; a parse error has none and is labelled as positioned in the transpiled output (see the 14:40 decision).
- `ModuleSpec.SourceName()` gives the reported name: the `@local/...` key with the resolved file's base name, so a `use` without an extension still reports `helpers.ts`. Pack-relative rather than the absolute path, so error text carries no machine-local paths and goja's URL resolution (absolute base, relative source) stays simple.
- Caching: `sync.Once` + `m.err` replaced by a per-spec `compileMu` and an `atomic.Pointer[goja.Program]` fast path. Traced the callers: `Registry.PrepareUse` from `executorImpl.bindUses` on every policy evaluation, and `LoadRequire` / `programFor` from `AliasRuntime.Require` on VM construction and nested `require`. All writes to `Program`, `TranspiledCode` and the failure fields happen under `compileMu`, and the atomic store comes last, so a lock-free reader that sees the program also sees those writes. Only successes are permanent. A failure is kept with the source text that produced it, so an unchanged broken module costs a file read per call rather than a transpile; a read error is never kept. Considered caching only successes outright, rejected because `bindUses` would re-transpile a broken module on every evaluation.
- Tests in `registry_errors_test.go` on `JSTestSuite`: syntax error at a known line and column, every esbuild error reported, a throw on line 5 reported as `@local/thrower.ts:5:` in both `Error()` and the full stack, recompile after a fix (and the same error value while unchanged, and a kept success after a later edit), read error retried, 16 concurrent first callers sharing one program, and the parse-error label. `go test -race ./runtime/js/` passes; `make test` and `make lint` (go1.25 golangci-lint v2.4.0) pass.
- Compiling modules during `sentrie validate` stays with #155; the registry node now links it.
- _Model: claude-opus-5-5_

### 22:25 - Review fixes for PR #198
- Review (`worktrees/_reviews/pr-198.md`) confirmed that mapped columns broke `SourceError`'s 1-based contract: goja's `file.File.Position` hands back the source map library's 0-based column unchanged, so `compileError` reported `delete y;` one short and dropped the column entirely for an error in column 1. This corrects my 14:40 and 14:43 note, which wrote it off as "can be one short" upstream: it is always one short, and esbuild-sourced errors from the same file were 1-based, so the two disagreed.
- Fix: `compileError` adds one to the mapped column. Every program `programFor` compiles carries esbuild's inline map, so the `se.File != nil` branch always sees a mapped position; the unmapped fallback (1-based) would only be hit for an offset in the banner or footer, which no compile error points at.
- New test `TestProgramFor_CompileErrorNamesSourcePosition`: `.js` modules with `delete y;` (line 5, column 8) and `with (o)` (line 3, column 1) pass esbuild and fail in goja's compiler, so they reach the mapped `CompilerSyntaxError` branch through `PrepareUse`. Checked that it fails (7 and 0) with the fix reverted.
- New test `TestProgramFor_ConcurrentCallersRecoverAfterFix`: 16 concurrent callers on a broken module all get an error; after the fix all get one shared program. Passes under `-race`.
- Graph: tscompile gotcha now says the skew remains only in goja's own stack frames. Registry gotchas gain two review suggestions: the source read under `compileMu` serializes concurrent callers on a broken module, and an in-place editor save can be read half-written and pinned as a success.
- Not taken: unexporting `ModuleSpec.Program` / `TranspiledCode`. Tests read `Program`, nothing else does, and unexporting is an API change beyond the review fix; left as is, with the publication rule documented on the fields.
- Merge plan with #194 (`105-isolate-pooled-vms`) and its stack recorded in the PR Review notes: `r.compiled.Add(1)` goes right after a successful `compileModule`, before `m.compiled.Store`, and one of the two `compiled` fields is renamed.
- `make test`, `make lint` (go1.25 golangci-lint v2.4.0), `make docs-graph` and `go test -race ./runtime/js/` pass.
- _Model: claude-opus-5-5_

### 14:01 - Restack onto main after #194 and the other bottom layers merged
- Rebased onto `main` (cfe249a). `runtime/js/registry.go` conflicted with #194 (105), as the merge plan in the PR description predicted. Resolved per that plan: this layer's mutex and atomic-pointer `programFor` replaces #194's `sync.Once` body. #194's compile counter now increments only after a successful `compileModule`, and I renamed it from `Registry.compiled` to `programsCompiled`, so it no longer shares a name with `ModuleSpec.compiled`. `CompiledPrograms()` is unchanged.
- `docs/codebase-graph/runtime.js.md`: merged the Statefulness and Performance notes, combining #105's per-call VMs with this layer's retryable compile and source-mapped stacks.
- _Model: claude-opus-5-5_

### 23:57 - Rebase onto main after the #154 and #168 stacks merged
- One conflict, in the `Registry` struct in `runtime/js/registry.go`: #203 added the `grant` field next to the compile counter this branch renames to `programsCompiled`. Kept the rename and the field.
- Vet, `go test -race ./...` and lint pass under Go 1.25, which CI uses.
- _Model: claude-opus-5-5_
