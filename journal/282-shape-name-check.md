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
