## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #134
- `is [not] defined` and `is [not] empty` parse and index but the evaluator has no case for them, so they fail with "unsupported expression node".
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:48 - Resolve an omitted optional fact as undefined
- `x is defined` must turn "missing" into `false`, so I traced how the runtime already represents a missing value: field access, out-of-range index and omitted optional derive/lambda parameters all yield `box.Undefined()` with no error.
- An omitted optional fact with no default is the odd one out: the executor never injects it, so `evalIdent` falls through to `identifier not found`. That made `extra is defined` and `extra.k is defined` errors for exactly the case the guard exists for.
- Fix in `evalIdent`, not in the new evaluators: after `GetFact` misses, a name declared in `p.Facts` as optional resolves to `Undefined`. Doing it at the identifier keeps field chains (`extra.k`) working through `accessField`'s undefined-receiver rule, rather than special-casing bare identifiers inside `is`.
- Ruled out swallowing operand errors inside `is defined`: that would hide typos (`identifier not found` for an undeclared name) and real failures (field access on a non-dict, failing calls). Only an undefined value means "not defined"; every error still propagates.
- Ruled out injecting `Undefined` into the facts map in the executor: `InjectFact` validates against the declared type and would need its own undefined exemption, a wider change.
- Derive path: the fact guard above already refuses every fact name inside a derive, and the new branch is also gated on `ec.evalDerive == nil`, so the derive resolution rule stays explicit.
- Side effect worth noting: any expression over an omitted optional fact now propagates `Undefined` (and a rule over it yields `Unknown`) instead of failing the request. No existing test relied on the error.
- _Model: claude-opus-5-5_
