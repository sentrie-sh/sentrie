# 164-stdlib-glob-object-units

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #164
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: First slice of the stdlib tracking issue: `glob.match`, object utilities (`get`, `keys`, `remove`, `union`, `filter`) and units parsing.
- Decision (default proposed to the user on 2026-10-31, not overridden): this PR is one slice of tracking issue #164, so its title says `Updates #164`. Native builtins via the `builtins` registry for these pure functions. X.509, YAML and JSON Schema are later slices.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:40 - Implement glob, object utilities and units
- Added eight native builtins in `builtins/glob.go`, `object.go`, `units.go`, registered in `builtins.go`; tests alongside, plus index validate-time and runtime end-to-end tests.
- Naming decision: the static checker (`runtimeBuiltinCallee`) resolves only identifier callees, so dotted names would skip validate-time checks. Used snake_case like `as_list`: `glob.match` -> `glob_match`, `object.get/keys/remove/union/filter` -> `object_get/keys/remove/union/filter`, units -> `parse_quantity` and `parse_duration`.
- Glob decision: own linear DP matcher instead of a library (stdlib only, no backtracking). Supports `*`, `**`, `?`, `[a-z]`, `[!a-z]`, `\`; rejects `{}` rather than silently diverging from OPA-style alternation. Empty delimiter list means no delimiters.
- Units decision: `parse_quantity` is Kubernetes quantity (n u m k M G T P E, Ki..Ei, decimal exponent), exact via `big.Rat`; `parse_duration` is Go `time.ParseDuration` returned as seconds (units align with `now()`). Split in two because `m` means milli in one and minutes in the other.
- Undefined arguments propagate, null and wrong kinds error; object utilities handle dicts (and lists in `object_get`), not host documents. `object_filter` is key-list based (OPA style), not a callback; the existing `filter` stays the list HOF.
- Problems: derive bodies cannot see policy lets, and dict literals need quoted keys; adjusted the end-to-end test. `golangci-lint` crashes in this environment, so only gofmt and vet were run.
- _Model: claude-sonnet-5-5, role: implementer subagent_

### 2026-10-31 14:41 - Test and CI pass for glob, object and units
- Added edge-case tests: `builtins/stdlib_edge_test.go` (glob `**` placement, delimiter runs, escapes, classes, unicode, caps exactly at and over the limit counted in runes; quantity exactness, negatives and malformed input; duration limits; object_get list and missing-intermediate paths; key ordering, duplicates, copies), `runtime/stdlib_builtins_derive_test.go` (all eight builtins inside derives, values compared across ten runs; undefined propagation and null errors through the dispatcher), `cmd/validate_stdlib_test.go` (`sentrie validate` rejects ten malformed calls), and `lang_test/0050-stdlib-glob-object-units.sentrie`.
- Bite proof: with the builtins files reverted to main, the runtime, cmd and lang_test tests fail and the builtins package no longer compiles (undefined glob caps); hand-made mutations of the caps, `**` delimiter handling, class delimiter handling, float rounding, sign handling, index validation and key sort each fail at least one test.
- Findings, not fixed (no production changes in this role): a `document`-typed value is rejected at validate time by `object_keys`, `object_remove`, `object_filter` and `object_union` (same as the existing `merge`), so these only accept dict-typed values; `parse_quantity("1e-400")` silently underflows to 0.
- Gate: make test, lint (0 issues), gofmt, docs-graph, race on builtins, index, cmd and lang_test all pass.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-10-31 14:58 - Round 1 review fixes
- `index/builtin_kind.go` `kindAllowed`: a static `document` is accepted where `dict` is allowed (a document is always a dict at runtime). Fixes `object_*`, and also `merge` and `count`, on document facts. Added index tests, including that a document is still rejected as a list.
- `parse_quantity`: a result that underflows to zero (`1e-400`) is now out of range; `1e-323` (subnormal) is accepted.
- `glob_match`: an empty delimiter list is now a runtime error. OPA's `glob.match` defaults `[]` to `["."]`, so treating it as "no delimiters" could fail open on a ported policy; rejecting avoids guessing. No validate-time check for a literal empty list: the checker has no per-builtin value hook, and adding one is out of scope for this slice. Description, graph node and tests updated.
- Builtin count corrected to 24 in the graph node and `ADDING_BUILTINS.md`.
- Renamed the language fixture `lang_test/0050-stdlib-glob-object-units.sentrie` to `0051-...` (#180 owns 0050); commit queue updated.
- Suggestion taken: `object_get` with a null object returns the default even for an empty path (a present key holding null still returns null). Suggestion taken: commit queue folds registration into one builtins commit so each commit compiles. PR wording changed from "linear" to "no backtracking blow-up".
- PR description Testing notes now record that lint ran (v2.4.0, 0 issues) and list the test agent's files; Review notes carry the document-kind change.
- _Model: claude-sonnet-5-5, role: implementer subagent_
