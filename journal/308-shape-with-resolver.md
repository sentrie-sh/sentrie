## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #307 #308
- `Shape.resolveDependency` (index/shape.go) resolves a `with` base by its last segment with later scopes overriding, while `detectShapeCycle` (index/validate.go) looks a policy shape's base up only in `ns.Shapes` by full FQN text. Both will call one shared resolver.
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - A qualified base that its namespace does not export is an error ("not exported").
  - An unqualified base resolves policy, then namespace, then visible exports, first match wins.
  - An unqualified base that matches only another namespace's export is an error: it must be qualified (the user chose this over keeping it allowed).
- Docs PR: sentrie-sh/website#77 (branch `308-shape-with-resolver` on `next`). Sentrie PR: #317.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 15:52 - One with-base resolver for hydration and the cycle check
- Re-reproduced both issues on the branch binary: #307 (`shape not found: A` for a policy shape `with` a sibling, and for `with com/base/A`) and #308 (`B with com/base/A` took its fields from the local `A`).
- Added `(*Index).resolveWithBase` in `index/shape.go`, the only `with` resolution. `resolveDependency` (now `resolveDependency(idx)`, it reads `s.Policy` instead of an `inPolicy` argument that commit always passed as nil) and `detectShapeCycle` (through a new `addShapeBaseEdge`) both call it.
- Behaviour, per the recorded decisions: a qualified base resolves only in the namespace it names and must be exported there (own namespace included, matching `LookupShape`); an unqualified base tries the policy, then the namespace, first match wins; an unqualified base held only by another namespace's export fails with "must be qualified: use <ns>/<name>". The "visible exports" step therefore only ever produces that error.
- The old cross-namespace fallback compared `ns.FQN` with the shape's own namespace after `ResolveShape(ns, ...)`, so it never checked export at all; the new resolver checks it.
- Existing tests that encoded the old behaviour, changed minimally: `TestShapeDependency_CompositionWithUnexportedShapeCrossNamespace` (expected the bug, now expects "not exported"), `TestValidate_ChildNamespaceShapeComposesParentShape` (fixture now exports `ParentS`), `TestShapeResolveDependencyFailsWhenComposedShapeNotExported_Cov` (base now qualified), two message assertions (`error resolving shape`, `shape not found` to `not found`), `TestShapeDependency_ResolveDependency_UsesPolicyShape` (sets the shape's `Policy`), and the two `isShapeDependencyNamespaceMiss` classifier tests removed with the helper. Legacy headers rewritten on the three touched test files.
- Graph: `index.shape`, `index.validate`, `index.resolve`, `index.commit` updated; `make docs-graph` passes.
- `GOWORK=off go test ./index/... ./runtime/... ./lang_test/... ./cmd/... ./loader/...` green. Regression tests and the full gate are the test agent's.
- _Model: claude-opus-5-5_
