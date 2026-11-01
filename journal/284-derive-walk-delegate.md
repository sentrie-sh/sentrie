## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #284
- `forEachDeriveExprChild` duplicates the general `forEachExprChild` walker #269 adds, so the two can drift when an expression kind is added.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `269-cycle-check-walk`, which is not merged yet; the PR targets that branch.
- Plan: make the derive walker delegate to `forEachExprChild`, keeping only what differs for derives, and cover it with the walker coverage test.
- _Model: claude-opus-5-5_

### 12:58 - Delegate the derive walker to the general walker
- Diffed `forEachDeriveExprChild` against `forEachExprChild`. Identical (children, order, nil handling aside) for every expression kind: call, infix, unary, ternary and elvis, lambda body, field and index access, list, map (key and value), cast, transform, is defined, is empty, is (Right only when set), both comment wrappers, leaves.
- Differences, and the decision for each:
  - Block statements: derive yields a `let`'s value and errors on any other statement (`derive walk: unsupported statement`); general yields every statement node. Intentional, kept as an override: it mirrors `walkDeriveBlock`, so derive analysis fails closed on a block holding a rule or a comment.
  - Import, `with` and attachment clauses: derive errors (`unsupported expression`), general descends. Intentional, kept: a derive body is a pure expression and these are rule-body forms; `builtin_calls.go` handles `ImportClause` itself before reaching the walker.
  - Unknown kind message: `derive walk: unsupported expression` vs `reference walk: unsupported node`. Kept the derive wording by mapping the general walker's own error (it can only be about `e`, since children are reached by the caller's recursion).
  - Nil children: derive yielded nil, general skips them. Accident, dropped: all three consumers (`walkDeriveExprDFS`, `walkDeriveExprSeen`, `walkScopedExprs`) already return on nil, so this changes nothing observable.
- No behaviour change to derive purity or cycle checks. Verified the built binary gives the same result before and after on a derive with a line comment.
- Found, not fixed (out of scope, the walker is unchanged here): a derive block with a line comment before a `let` or the yield is rejected by purity ("derive body may only contain let declarations before yield (got *ast.CommentStatement)") because `walkDeriveBlock` rejects `CommentStatement`. The parser allows `LineComment` in blocks, so this looks like a bug for a follow-up issue.
- Tests: moved the samples and expected identifiers of `TestForEachExprChildCoversEveryASTNode` into `astWalkSamples` / `astWalkExpected`, and added `TestForEachDeriveExprChildCoversEveryASTNode`, which runs every expression kind through the derive walker and compares with the same expectations, plus the derive overrides. Mutating the let override makes it fail.
- Updated the `index.expr_walk` and `index.derive_expr_walk` graph nodes, and the sentences in `index.builtin_check` and `index.derive_purity` that described the old switch. `go test ./...` and `GOTOOLCHAIN=go1.25.0 make lint` pass.
- _Model: claude-sonnet-5-5_
