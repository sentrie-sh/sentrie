## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #271
- Inside a derive body, casts, typed lets, parameter types and the return type resolve shape names in the caller's policy, while the index resolves them in the derive's own scope.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- No open design choice: use the derive's shape scope at every type-check site, the same way `isShapeScope` already does for `is` (#150).
- _Model: claude-opus-5-5_

### 14:47 - Resolve every derive type check in the derive's scope
- Moved `isShapeScope` from `runtime/eval_is_type.go` to `runtime/typeref.go` as `typeCheckScope`, next to `validateValueAgainstTypeRef`, since five files now need it.
- `invokeDerive` validates parameter and return types in `d.ShapeScope()`; `evalCast` and the typed-let branch of `evalIdent` validate in `typeCheckScope(ec, p)`; `lambdaCallable.Invoke` validates parameter and return types in `typeCheckScope(c.capture, site.Policy)`.
- Kept the derive body evaluated with `callerPolicy`: a lambda the caller passes in must keep resolving in its policy, and `typeCheckScope` already reads `ec.evalDerive` for every check in the body. Changing the body's policy was not needed.
- The lambda uses its capture context, not the call site's: a lambda written in a derive body and applied by `filter` from that derive has `site.Policy` = the caller's policy, so `site.EC` would only be right by accident, and a policy lambda applied inside a derive must not pick up the derive's scope. A mutation using `site.EC` fails the policy-lambda test.
- Audited the other paths: derive-to-derive calls go through `invokeDerive` with the callee's `ShapeScope`; the pipeline operator desugars to calls and casts; memoized calls go through the same target (the per-call-node memo key could previously cache a result computed under one caller's shapes for another caller; now the result no longer depends on the caller); fact validation in `executor.go` never runs in a derive; `LookupShape` is only used by `is`.
- Index side: `resolveIsExpressions` checks shape names only at `is` sites. Casts, typed lets, parameter and return types are never checked for shape existence by the index, and kind inference in `builtin_kind.go` uses `d.Policy` (nil for a namespace-level derive, so namespace shapes give an unknown kind). The index never uses the caller's scope, so validate and runtime no longer disagree, but a misspelt shape at those sites is still a runtime error. Left for a follow-up.
- Not changed: parameter constraint arguments are evaluated in the caller's context (`callerEC`) before the detached child exists, and the index does not walk constraint arguments for purity. Pre-existing; noted in the derive_invoke node.
- Tests: `TestDeriveTypeChecksUseDeriveScope` (five sites, caller with and without its own `Foo`, plus a value that is only the caller's `Foo` must fail) and `TestPolicyLambdaThroughDeriveUsesPolicyScope`. Reverting each site alone fails exactly its subtest.
- _Model: claude-opus-5-5_
