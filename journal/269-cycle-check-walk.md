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
