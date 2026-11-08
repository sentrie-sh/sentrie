## Binaek Sarkar (binaek89@gmail.com)

### 07:34 - Pick up #282
- `sentrie validate` does not check shape names in casts, typed lets, or derive and lambda signatures, so a missing shape there only fails when the expression is evaluated.
- Picked up with #279 to #285 in one batch (follow-ups from #269 to #273); one pickup push and a draft PR per branch before implementation.
- Stacked on `273-constraint-arg-check`, which is not merged yet; the PR targets that branch.
- Plan: resolve every shape type reference at validation, in the scope the runtime uses, reusing the type position walk #273 added for constraint arguments, and report a missing or unexported shape at its position.
- _Model: claude-opus-5-5_

### 13:02 - Check every shape name in a type at validation
- Moved the type ref position walk out of `checkConstraintArgs` into `index/typeref_walk.go`: `forEachTypeRef` passes each type ref with the scope its shape names resolve in (`typeRefSite`: policy, namespace, and whether it is an `is` type), and `forEachNestedTypeRef` expands nullable, list, dict and record types. The constraint argument check and `checkIsTypeShapes` both use it now, so there is one walk.
- New `checkShapeRefs` (`index/shape_refs.go`) runs right after `resolveIsExpressions`. It resolves every `ShapeTypeRef` outside an `is` type with `LookupShape` and reports `shape '<name>' not found at <span>: index error` or `shape '<name>' not exported at <span>: index error`, in source order (shared `joinSiteErrors`, renamed from is_resolve's sort). The text is the one shape composition and the runtime already use.
- Scope: a shape's alias and field types resolve in the shape's scope, a derive's signature and body in the derive's (policy, or namespace for a namespace-level derive), everything else in its policy. `is` types are skipped because is_resolve already reported them in its own wording.
- Risk noted in the graph, not fixed: the runtime validates a shape field's type in the evaluating policy's scope, not the shape's, so a namespace shape whose field names a policy-only shape is rejected here though one policy could have evaluated it.
- The scan of every `.sentrie` file (sandbox, reset afterwards) found fixtures naming types that do not exist: `float` in example_pack/minimal.sentrie, `int` in example_pack/shapes.sentrie and lang_test 0029, `any` in lang_test 0019, `ShapeName` in lang_test 0020, and `dict[bool]` in runtime/import_envelope_test.go. All were fixture mistakes (each would fail at runtime once evaluated), so I fixed them: `number`, `document`, `boolean`, and a declared `shape ShapeName string` to keep 0020's shape-typed fact.
- `TestEvalIsTypeCheckErrorsPropagate` built a missing field shape on purpose to reach the runtime error; validation now rejects it, so the test asserts that and runs on the unvalidated index.
- _Model: claude-opus-5-5_

### 13:10 - Test the shape name check gaps
- Added three index suite tests to `index/shape_refs_test.go`: a missing shape nested in list, dict, record and nullable wrappers (fact, typed let, and a rule's lambda parameter and return type, each with its span); lambda shape names resolving in the scope of the thing holding them (a policy shape is visible in a lambda of a rule or policy derive and rejected in one of a namespace derive); and a missing shape reported once where positions overlap (a cast beside a known `is` type, an `is` type reported only in is_resolve's wording, namespace and policy shape fields).
- Proved they bite: removing the `checkShapeRefs` call fails 27 subtests; dropping dict value recursion in `forEachNestedTypeRef` fails the nested and constraint argument tests; giving a namespace-level lambda a policy site fails the scope tests.
- The base branch's constraint argument tests (all positions) still pass on the shared walk.
- Ruled out: the `site.isType` guard in `checkShapeRefs` and the shape dedupe in `forEachTypeRef` are not observable. An `is` type with a missing shape makes resolveIsExpressions return before checkShapeRefs runs, so removing the guard changes nothing; shapes are keyed by name, so no statement is visited twice. A source with both a bad `is` type and a bad cast reports only the `is` error until it is fixed (stage order), which the cast-beside-known-is test avoids.
- _Model: claude-sonnet-5-5_

### 13:15 - Address review: merge order and scope notes
- Review (opus) found the code correct once #276 (#271) and #291 (#281) land, but merging this PR first would reject packs that run today: a namespace derive whose signature, or a namespace shape whose field, names a shape only the calling policy declares (the runtime still uses the caller's scope for both on this base).
- Tied the graph notes on that disagreement (`index.shape_refs`, `index.typeref_walk`, `runtime.typeref_shape`) to #281, with the instruction to delete them when it lands, instead of stating them as lasting facts.
- PR description: Review notes now say to merge after #276 and #291 and name both cases; the bite count in Testing notes now matches the 13:10 entry (27 subtests).
- _Model: claude-opus-5-5_

### 2026-11-08 14:10 - Rebase onto the rebased #273 branch
- #273 was rebased onto main after #269 merged (as #274), so this branch moved onto it with `git rebase --onto`.
- #269 conflicted in two places, both appends: `cmd/validate_test.go` (kept #269's cast cycle test and this branch's missing shape test) and the CALLS rows in `docs/codebase-graph/index.validate.md` (kept both `index.expr_walk` and `index.shape_refs`). `index`, `runtime` and `cmd` tests and `make docs-graph` pass.
- _Model: claude-opus-5-5_

### 2026-11-08 15:05 - Rebase onto main after #271, #273 and #283
- #273 merged (as #278) under this branch, so it moved onto main with `git rebase --onto`; no textual conflicts.
- Combination conflict, not textual: two tests from branches already on main build a derive that names a shape it cannot see and expected `validate` to pass. This branch now rejects them at validation.
  - #283's `TestBuiltinCheckNamespaceDeriveCannotSeePolicyShape` now expects `shape 'Name' not found` and no `filter` error.
  - #271's `TestDeriveMissingShapeInDeriveScopeIsAnErrorNotACallerFallback` asserts validation reports `shape 'Nope' not found` four times, then still checks the runtime error on an unvalidated index.
- `index`, `runtime`, `cmd` and `lang_test` pass.
- _Model: claude-opus-5-5_
