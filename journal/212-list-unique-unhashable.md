## Binaek Sarkar (binaek89@gmail.com)

### 22:07 - Pick up #212
- Branch cut from `211-bound-cidr-expand`. #111 rewrites `builtin_collection_utils.go`'s error sites, so this stays in the chain 208 to 213 rather than branching from `main`.
- Plan: `list_unique` keys each exported element by a stable JSON encoding instead of using it as a Go map key, keeping first-occurrence order.
- _Model: claude-opus-5-5_

### 22:08 - Deduplicate by canonical key
- `writeCanonicalKey` writes a stable JSON-style key per exported element: sorted object keys, `int64` and `float64` of one value encode the same (goja exports `1` and `1.0` differently), `-0` folds into `0`, and NaN and the infinities become bare tokens. Plain `json.Marshal` was the first idea but fails on NaN and Infinity, which would have turned a working scalar list into a throw.
- Values with no canonical form (functions) throw `list_unique: element <i> cannot be compared`, per #111.
- The result now holds the caller's own values read back from the array (`src.Get(i)`) in a real `vm.NewArray`, rather than re-converted exported values. Found while testing: re-converting a Go map lost object key order (`{a:1,b:2}` came back as `{b:2,a:1}`), and object identity was lost too.
- Behaviour change: `[NaN, NaN]` now dedups to `[NaN]` (Set semantics); before, each NaN was kept because NaN never equals itself as a map key.
- Tests in a new `builtin_collection_utils_test.go`: scalars, strings, `1` vs `"1"`, `1` vs `1.0`, booleans and null, arrays, objects with reordered keys, nested, mixed, empty, non-finite numbers, a function throwing, and identity of kept objects.
- `.d.ts` and the `runtime.js.builtins` graph node document the comparison.
- _Model: claude-opus-5-5_
