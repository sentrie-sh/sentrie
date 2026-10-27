## Binaek Sarkar (binaek89@gmail.com)

### 14:34 - Pick up #138
- `@sentrie/crypto.sha256` called `sha256.New().Sum(input)`, which appends the empty-string digest to the input bytes instead of hashing them. The output carried the plaintext.
- Chose to fix rather than remove the export: removing it breaks any pack that imports it, and the fix is one line.
- Output encoding matches `@sentrie/hash.sha256` (lowercase hex) so the two modules agree on the same function name.
- Wrong argument count throws (`panic(vm.NewGoError(...))`) rather than returning an Error object, following the convention #111 is moving every builtin to.
- _Model: claude-opus-5-5_

### 14:37 - Tests and graph
- Added `JSTestSuite` cases: known answers for `""` and `"abc"`, agreement with `hash.sha256`, the digest not containing the input, and arity throwing so `try`/`catch` catches it.
- Updated `runtime.js.builtins` gotchas: `crypto` no longer has the return-don't-throw hazard; `jwt` still does until #111.
- _Model: claude-opus-5-5_

### 11:34 - Restack onto main after #193 and #195 merged
- Rebased onto `main` (610352e). One conflict, in the Dependencies Risk list of `docs/codebase-graph/runtime.js.builtins.md`: kept #111's "`crypto` and `jwt` are security-critical surfaces" bullet (failures throw) and this layer's `crypto.sha256` bullet, and dropped the stale return-don't-throw wording.
- _Model: claude-opus-5-5_
