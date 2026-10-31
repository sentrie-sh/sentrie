## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #269
- The reference cycle check in `index/validate.go` skips index access, casts, transforms, comment wrappers, map keys and import `with` values, so a cycle through one of them passes `sentrie validate` and fails only at decision time.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer before pickup: replace the hand-written switch with a walk that visits the children of every expression kind, so a new expression kind cannot be missed again.
- _Model: claude-opus-5-5_

### 14:44 - Walk every expression kind in the cycle check
- Added `index/expr_walk.go`: `forEachExprChild`, a general child walker over every expression kind plus block statements, import clauses and `with` clauses. `forEachDeriveExprChild` did not fit: it errors on non-let block statements and has no import clause case, both needed here. `walkScopedExprs` is scope-tracking and builds on the derive walker, so neither was reused; neither was moved onto the new walker (not small and safe, and out of scope).
- `addNodes` now keeps only the cases that carry edge semantics (identifier edge; rule and let add a node and become the referrer) and descends into everything else through the walker. It returns an error and `detectReferenceCycle` reports it wrapped in `ErrIndex`; an unknown kind fails loudly instead of being skipped.
- A rule statement inside a block now also walks its default and when slots, not only its body (the parser only puts lets and comments in blocks, so this is the general-walker behaviour, not an observable change today).
- Import `with` values are now walked, with the importing rule as referrer.
- Deleted the commented-out `containsSelfReference` block: dead, in the area being rewritten.
- Tests: table in `index/validate_walk_test.go` (rule cycle, let cycle and a non-cyclic use for seven positions of the six kinds, plus import `with`); shown failing against the old `addNodes`. `TestForEachExprChildCoversEveryASTNode` parses the `ast` package source and fails when a new expression kind has no walker case.
- A trailing comment swallows the rest of its line, so rule-body test sources put the closing brace on its own line.
- Self-recursion (`rule r = r`, #155) untouched.
- New graph node `index.expr_walk`; updated `index.validate`, `index.derive_expr_walk`, overview catalog.
- _Model: claude-sonnet-5-5_

### 14:47 - Regression tests for the cycle walk
- Added to `index/validate_walk_test.go`: `TestValidate_ReferenceCycleAcrossMixedKinds` (three-node cycles that change kind each hop, cast inside index access inside a map key, `is defined` / `is empty` / `is T` left side, a value on the right of `is`, lambda bodies, `default` and `when` slots) and `TestValidate_AcyclicPackUsingEveryKindStillValidates` (shape operands, field names that equal rule names, comments, lambdas, import `with`, ternary and elvis, no cycle reported).
- Added `TestValidateCmdRejectsReferenceCycleThroughCastAndIndex` to `cmd/validate_test.go`, next to the existing #151 validate test: same CLI path, cycle through a cast and an index access.
- Bite check: with `index/validate.go` restored from main, the index kinds, `default`/`when`, mixed-kind and cmd tests fail; the `is` and lambda cases pass on main too (the old switch already covered them), so they are regression guards, not new-behaviour proofs. The acyclic test fails if the walker yields a field name as an identifier (checked by temporary mutation).
- Ruled out: a shape-operand false-positive test that bites. A shape name equal to a rule name is rejected earlier ("names both a shape and a value"), and an edge to a name that is not a rule is ignored by the DAG, so no input can show it.
- Ruled out: a built-binary test; the existing validate tests run the CLI in process, so the new one does too. The built binary was run by hand: cyclic pack exits 1 with `infinite recursion: a -> b -> a`, acyclic exits 0; the main-built binary exits 0 on the cyclic pack.
- _Model: claude-sonnet-5-5_

### 14:58 - Walker coverage test checks the children, review suggestions declined
- `TestForEachExprChildCoversEveryASTNode` now compares the identifier names each sample yields (recursively) with an expected set, so a kind given an empty `return nil` case fails. Bite check: replacing the cast case body with `return nil` fails the CastExpression sample; restored from HEAD.
- Declined review suggestion: prefix the unsupported-node error in `index/validate.go` with the policy and rule name. The path is unreachable from the parser today and the error already names the node type.
- Declined here: make `forEachDeriveExprChild` delegate to `forEachExprChild`. Out of scope for this issue; left as a follow-up.
- _Model: claude-sonnet-5-5_

