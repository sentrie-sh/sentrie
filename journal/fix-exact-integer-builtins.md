# fix-exact-integer-builtins

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-09 18:09 - Fix main after exact integers met the new builtins
- `main` CI failed on the merge of #261 (exact integers, Closes #166). #268 (glob, object and units builtins) merged first, and its tests still expected integral results as `float64`.
- With #261 an integral value boxes as an exact integer, so `Any()` returns `int64`. That is the intended behaviour, so only the expectations change: `builtins/object_test.go` (the `object_get` default) and `runtime/stdlib_builtins_derive_test.go` (the `count` result and the integer values in the object builtin results). `parse_quantity` and `parse_duration` results stay `float64`.
- How it slipped through: the ship script's `gh pr checks --watch` returned on the PR's earlier skipped draft runs before the new run registered, so #261 merged without waiting for its own CI. The script now waits for the run on the pushed head SHA.
- `make build`, `make docs-graph` and `make test` pass.
- _Model: claude-opus-5-5, tokens: ~6k_
