## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #151
- A `with` binding naming a fact the target policy does not declare is silently ignored.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:48 - Reject unknown with fact names
- Added `checkImportWiths` (index/import_withs.go), run in `validate` right after `detectRuleCycle`; it reports each `with` name that is not a key of the imported policy's `Facts`, as `xerr.ImportUnknownFactError` (wraps `ErrIndex`) at the with clause's span, listing the declared facts sorted. `sentrie validate` now exits 1 on the issue's repro.
- Key space: `Policy.Facts` keys, which are fact aliases (`AddFact` does `p.Facts[fact.Alias]`). `ImportDecision` looks up `p.Facts[with.Name]` and `ExecRule` injects by the same keys, so the static and runtime checks agree. This follows #147's current resolution: if #147 changes the key, `Policy.Facts` changes with it and this check needs no edit. I did not fix #147.
- Runtime fallback: the `continue` in `ImportDecision` is now an invalid-invocation error sharing `xerr.ImportUnknownFactMessage` with the index check, so a pack built without the index still fails closed.
- Coverage of import sites: the parser only produces an `ImportClause` as a rule body (parser/rule.go), the same assumption `detectRuleCycle` makes, so only `rule.Body` is checked; nowhere else can an import appear. Noted in the graph node.
- Fixes to existing tests: lang_test fixture 0019 bound `user`/`resource` on targets with no facts, so the targets now declare them as optional facts. The runtime test TestImportDecisionSuccessWithWithInjection bound a deliberately unused `unused` name; that clause is removed, since it is now an error by design.
- Seen, not fixed: the with-clause span ends at column 0 when the expression is a map literal (the literal's end position), so the reported range reads like 9:44-9:0. The start is right.
- Kept the edit in runtime/imports.go confined to the with loop (plus the header and the xerr import) to limit conflicts with #136.
- _Model: claude-sonnet-5-5_
