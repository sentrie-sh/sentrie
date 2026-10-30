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

### 15:51 - Regression tests for #151
- index (index/import_withs_test.go): the issue repro (typo'd `usr` rejected, valid `user` on the same target accepted), error type via errors.Is(ErrIndex) and ImportUnknownFactError, message listing sorted facts, span start (file, 0-based line, 1-based column, and the offset lands on the `with` keyword), a target with no facts, a valid optional and a valid required fact, matching by alias, several withs where only the bad ones are reported (in source order), and a cross-namespace target.
- runtime (imports_test.go): ImportDecision called on a hand-built index, which never ran the check, returns an InvalidInvocation error for an unknown with name and sets the node error.
- cmd (validate_test.go): `sentrie validate` returns an error (non-zero exit) on the issue's repro, naming the typo'd fact and the declared ones.
- Span end columns are not asserted (lexer end-column bug, #236). Lines are 0-based in this lexer, columns 1-based.
- Proved by reverting index/validate.go and runtime/imports.go to main: the reject tests (index, runtime, cmd) fail. A mutation that rejects every with name fails the accept tests, so they guard against over-rejection.
- Header of runtime/imports_test.go rewritten to the SPDX form on touch.
- _Model: claude-sonnet-5-5_

### 15:53 - Review round 1 for #151
- Message now names what is accepted: `import binds unknown fact "user": policy t/q accepts fact aliases: u; fact "user" is bound by its alias "u"` (a target with no facts still reads `declares no facts`). `xerr.ImportUnknownFactMessage` takes an alias to declared-name map, built by `Policy.FactNamesByAlias`, so the index and runtime share it. Tests: index (aliased fact hint), xerr (table), and the existing message assertions in index, runtime and cmd updated.
- Correction to the 058f06f entry: it said a #147 change to the key needs no edit here. That is too strong. #147 may inject by name and bind by alias without re-keying `Policy.Facts`; this check, the runtime fallback and the message wording must follow #147's choice. The graph node says so now.
- Removed `xerr.ErrImportUnknownFact` (no callers; the index and runtime build `ImportUnknownFactError` directly).
- Moved the helper to index/policy.go as `FactNamesByAlias` (the sorted-names `FactNames` was no longer needed, the message sorts).
- Third copy of the ns/policy split: extracted `importTarget` in index/validate.go, used by `detectRuleCycle` and `checkImportWiths`. Did not fold the check into `detectRuleCycle`: it returns a graph plus the first failure and stops at a cycle, while the with check collects every miss. Left the runtime copy alone to stay clear of #136.
- Runtime fallback now returns `ImportUnknownFactError` (wraps `ErrIndex`, carries the with clause's span) instead of `ErrInvalidInvocation`, so the API does not map a pack-authoring error to 400. Runtime test asserts `ErrIndex`.
- _Model: claude-sonnet-5-5_

### 15:54 - Review approved
- Round-2 review approved. Took its one remaining suggestion: the runtime.imports graph node now says the error lists the fact aliases the target accepts, matching the message and the Exceptions line.
- Left as is, per the review: when two facts share a declared name under different aliases, the hint appears once per alias.
- _Model: claude-opus-5-5_

### 02:47 - Merge main after #236 and expect 1-based lines
- #236 landed first and makes lines 1-based, so merged `main` in. There was no textual conflict, but `index` failed as the PR description predicted: the span assertions in `index/import_withs_test.go` still expected 0-based lines.
- Bumped them by one (`8` to `9`, `10` to `11`, `2` to `3`) and reworded the comment. This is the same fix the scratch integration of the batch found (patch 0001). `index` and `runtime` pass.
- _Model: claude-opus-5-5_
