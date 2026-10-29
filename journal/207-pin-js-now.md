## Binaek Sarkar (binaek89@gmail.com)

### 21:55 - Pick up #207
- Branch cut from `main`. Independent of the open stacks: only `runtime/js/ts_src/_sentrie_js.ts` changes.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 22:00 - Remove the `now` export from `@sentrie/js`
- Decision: remove `now` rather than re-pin it. It returned `Date.now` (wall-clock milliseconds, a fresh instant per call). `@sentrie/time.now()` returns the pinned execution start in seconds, so re-pointing `now` at it would silently change every value by a factor of 1000. A removed export fails loudly instead: `use { now } from @sentrie/js` errors at bind time with "missing required export" (`jsBindingConstructor` in `runtime/executor.go`, untouched here).
- Breaking for any pack that imports `now` from `@sentrie/js`; the migration is `use { now } from @sentrie/time` (seconds, pinned).
- Searched for users: no `.d.ts` exists for `@sentrie/js`, and no `lang_test` fixture, `example_pack` policy or doc imports `now` from it, so nothing else needed changing beyond the graph node.
- Rewrote the legacy headers on `_sentrie_js.ts` and `builtin_ts_src_test.go` to the two-line SPDX form.
- Tests: `TestBuiltinJSTS_NoNowExport` (source contains no `now` export and no `Date.now`), and `RuntimeTestSuite.TestSentrieJSNowExportRemoved` in a new `runtime/sentrie_js_module_test.go`, which binds the real embedded module and checks `now` fails with "missing required export" while `round` still binds. Kept it out of `executor_test.go` to avoid conflicts with the other open stacks.
- Graph: `runtime.js.builtins` now records that the `js` module deliberately has no `now`.
- _Model: claude-opus-5-5_

### 10:23 - Rebase onto main after #205 and #202 merged
- Rebased onto `main` (1197589) ahead of merge. No textual conflict, but `TestSentrieJSNowExportRemoved` called `exec.jsBindingConstructor`, which #205 replaced with `getModuleBinding` (same arguments, same "missing required export" error), so the runtime package no longer compiled. The test now binds through `getModuleBinding`.
- Build, vet and `go test ./...` pass on the rebased head.
- _Model: claude-opus-5-5_

### 23:16 - Rebase onto main after the #154 stack merged
- Rebased onto `main` (0e909ba). One conflict, in the "Non-deterministic" and "TypeScript" bullets of `docs/codebase-graph/runtime.js.builtins.md`: kept main's note that `time.now()` is pinned to the execution start alongside this branch's note that the `js` module has no `now`.
- _Model: claude-opus-5-5_

### 23:28 - Rebase onto main after #200 merged
- #200 (#168 layer 1) merged first, so this branch carried the overlap. Its nondeterminism table in `runtime/js/nondeterminism.go` listed `now` among the `js` module's nondeterministic exports. With `now` removed here, the entry is `{"random"}`, and `TestNondeterministicExportLookup` now asserts `@sentrie/js` `now` is not listed. `runtime.js.registry` and `runtime.js.builtins` drop `js.now` from their lists.
- The doc conflict in `runtime.js.builtins.md` keeps #200's list of exports (without `js.now`) and this branch's note that the module has no `now`.
- _Model: claude-opus-5-5_
