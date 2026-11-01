## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #283
- The builtin kind check passes `d.Policy` (nil for a namespace derive) when resolving parameter types, so a shape-typed parameter of a namespace derive is unknown and the check is skipped.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Plan: resolve a namespace derive's parameter shapes in `d.Namespace`, matching the runtime's derive scope.
- _Model: claude-opus-5-5_

### 12:54 - Resolve a namespace derive's shapes in the kind check
- `newDeriveKindCheckCtx` now uses `d.ShapeScope()` (added for #150: the derive's policy, or a `Policy` holding only the derive's namespace) for the parameter kinds and as the ctx policy, instead of `d.Policy` which is nil for a namespace derive. Reused it rather than allocating a new stub per derive.
- Checked the other readers of the ctx policy: `lookupDerive` and `namesValue` test `k.derive` before the policy (the rules/derives lookups are skipped in a derive body), `shapeNamespace` returns the stub's namespace, which equals `d.Namespace`, and `LookupShape` / `resolveShapeReadOnly` only read `Shapes` and `Namespace`. A non-nil namespace-only policy is safe.
- Tests: namespace derive with a string-alias shape param reports the filter error at the argument span (fails without the fix), a list-alias shape param still validates, and a policy-level derive is unchanged (reject and accept).
- Graph: `index.builtin_kind` documents the derive ctx's shape scope.
- _Model: claude-sonnet-5-5_

### 13:09 - Regression tests for the namespace derive shape scope
- Added tests for: a typed `let` naming a namespace shape inside a namespace derive body, a cast `t as Name` to a namespace shape, a shape alias chain (alias of an alias) on a parameter, a namespace derive in `com/lib` called from `com/app` where `com/app` has a same-named shape of another kind (the derive's namespace decides, not the caller's), a namespace derive that names a policy-local shape (stays unresolved, so no leak of the calling policy's shapes), and a policy shape shadowing a namespace shape in a policy derive.
- Proved the bite by restoring `index/builtin_kind.go` from main: the typed let, cast, alias chain, param reject and caller-namespace tests fail. The policy shadowing test fails against a mutant that always uses a namespace-only scope; the policy-local-shape test is a guard against a leaking scope and passes on main by design.
- Ruled out: a shape-typed return type (the kind check never reads a derive's return type, call results of derives are unknown) and a pure lookalike-accept test (passes on main and on any plausible mutant).
- _Model: claude-sonnet-5-5_
