## Binaek Sarkar (binaek89@gmail.com)

### 21:55 - Pick up #218
- Branch cut from `168-3-deterministic-js`. #168 adds a nondeterminism walker over the same nodes; the two should share one walk.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 22:11 - Share the walk and cover every policy expression site
- #168's `forEachBuiltinCall` (`index/builtin_calls.go`) already walked fact defaults, policy lets, rule slots, export attachments and derive bodies for the nondeterminism and capabilities checks, with its own copy of `walkExpr`'s scoping. `checkBuiltinCalls` had a second walker over rules and derives only.
- Made `checkBuiltinCalls` a visitor over `forEachBuiltinCall` and deleted `walkExpr`. The visitor now receives the `kindCheckCtx` the call resolved in, since `checkBuiltinCall` needs the scope for argument kinds; `checkBuiltinCall` takes the already-resolved `decl` instead of resolving it again. Nondeterminism and capabilities visitors ignore the context.
- Found one more site neither walker reached: the `with` expressions of `rule r = import decision ... with f as <expr>`. `ast.ImportClause` fell into `forEachDeriveExprChild`'s default (error, swallowed). Added an explicit case in `visitBuiltinCalls` rather than in `forEachDeriveExprChild`, which the derive purity and cycle walks share and where an import clause can never appear. The runtime evaluates these in the importing rule's context (`runtime/imports.go` `ImportDecision`), so the rule scope is right. Effect: nondeterminism warnings and capability refusals now also cover with clauses.
- Ruled out type-constraint arguments: the parser accepts only literals there (`parseConstraintLiteral`), so no call can appear. Shapes hold no expressions.
- Added a namespace-level cancellation check to `forEachBuiltinCall` so the existing `TestCheckBuiltinCallsCancelledDuringLoops` cadence (namespace, policy, derive) still holds.
- Tests: arity failure with its span at a policy let, an optional fact default, an export attachment and an import with clause; a kind failure at a policy let; a let shadowing `count` stays unchecked across the policy; a nondeterminism warning in a with clause. The five failure tests fail against the old walker.
- _Model: claude-opus-5-5_

### 22:11 - Paused at the plan limit (handoff)
- The 5-hour usage window reached 98%, so work stops here per the usage guard.
- Done and committed (WIP commit): the walker change and all tests; `go test ./index/ ./lang_test/` passes.
- Remaining: update `docs/codebase-graph/index.builtin_check.md` (sections 3 and 4 still describe two loops and `walkExpr`; the performance note about one walk per rule slot), and the `forEachBuiltinCall` description in `docs/codebase-graph/index.md` (lines 49 and 54: add import with clauses); run `make docs-graph`; write the full `PR_DESCRIPTION_218-check-all-builtin-calls.md` (Review notes: with-clause coverage extends the nondeterminism warnings and capability refusals too); run `make test` and `make lint`.
- _Model: claude-opus-5-5_
