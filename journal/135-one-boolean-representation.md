## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #135
- Literals and logical operators box booleans as `ValueTrinary` while comparisons, facts and JS box `ValueBool`, so `true == true` can be false, `false` is truthy in TypeScript, and JSON output mixes strings and booleans.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:52 - Box known truth values as bool
- Normalised in one place: `box.Trinary` now returns `Bool(true)`/`Bool(false)` for `trinary.True`/`trinary.False` and a `ValueTrinary` only for `Unknown` (an out-of-range `trinary.Value` boxes as `Unknown`, matching how `trinary` already reads it). Every producer goes through it: literals (`runtime/eval.go`), `and`/`or`/`xor` (`eval_infix.go`), `not`/`!` (`eval_unary.go`), `as trinary` casts, `FromAny(trinary.Value)`, and the imported decision envelope's `state`. Comparisons, facts and JS returns already boxed `ValueBool`, so `box.EqualValues` and `ContainsValue` needed no change.
- Chose constructor normalisation over patching each producer so no future caller can build a known `ValueTrinary`. Ruled out issue option 2 (kind-aware equality) as agreed.
- `Value.TrinaryValue()` now accepts `ValueBool` too (True/False) and `ValueTrinary` (Unknown). Consumers that read truth values through it (the five trinary constraints, `DecisionOf`, `validateAgainstTrinaryTypeRef`) keep working when a known value is a bool; `validateAgainstTrinaryTypeRef` dropped its separate `BoolValue` branch. Truthiness consumers (ternary, `when`, `all`/`any`/`filter`, decisions) already used `TrinaryFrom`, which handles both kinds.
- Audited `builtins/collection.go` `distinct` keying: the `tri:` key now only ever sees `Unknown`, and `true` from a literal and from a comparison share the `bool:` key, which is the fix.
- Static kinds (`index/builtin_kind.go`): a `true`/`false` literal now resolves to `ValueBool` and `unknown` to `ValueTrinary`, mirroring the runtime. A `trinary`-typed value keeps static kind `ValueTrinary`, now meaning "bool or unknown"; `kindAllowed` admits it only where both kinds are allowed. No builtin parameter lists either kind today, so diagnostics are unchanged.
- JSON shape: a decision whose rule body is a logical operator or a `true`/`false` literal now has `"value": true|false` (JSON boolean) instead of the string `"true"`/`"false"`. `state` keeps its string enum `"true"`/`"false"`/`"unknown"`. An `unknown` value still marshals as the string `"unknown"`. Inside policies, an imported decision's `state` is now a bool for a known state, so `d.state == true` holds; #136 decides how the envelope itself reads as a truth value.
- Behaviour change worth review: `true as number` / `(not false) as number` now give `1`/`0` through the existing bool branch; before, a literal or logical result errored with "cannot cast trinary to number". `unknown as number` still errors.
- Existing tests changed because they encoded the old split:
  - `box/box_internal_test.go`: `Bool(true).TrinaryValue()` used to report not-ok; the wrong-kind probe is now `Number(1)`. Added `TestTrinaryBoxesKnownTruthAsBool`.
  - `index/builtin_kind_internal_test.go`: a `true` literal resolves to `ValueBool`; added `unknown` and `kindAllowed` cases.
  - `runtime/eval_unary_test.go`, `eval_infix_test.go` (xor), `eval_chain_test.go` (unary dispatch), `eval_branches_test.go` (`not true`): expected `trinary.True/False` from `.Any()`, now a Go bool / `box.Bool`.
  - `runtime/eval_cast_test.go` `TestEvalCastNumberFromBool` and `eval_branches_test.go`: expected `false as number` and `(!false) as number` to error; they now yield `0`/`1`, and the error is asserted for `unknown as number` instead.
  - `runtime/imports_test.go`, `eval_chain_gap_wave2_test.go`: envelope `state` is now `box.Bool`.
  - `runtime/value_test.go`: `FromAny(trinary.True)` is `ValueBool`; `Trinary(False).Any()` is `false`; added the `Unknown` cases.
- Verified the issue repro with the CLI: rules `a`, `b`, `c`, `not false` and `(true and true) == (1 == 1)` all decide `true` with `"value": true`.
- Rewrote legacy license headers on touched files.
- _Model: claude-opus-5-5_

### 15:58 - Pass truth values into JavaScript as primitives
- `normalizeBoundaryForJS` (`runtime/modules.go`) now enumerates what it converts and returns an error instead of falling through for truth values: `bool` crosses as a JavaScript `boolean`, `trinary.Unknown` as `undefined`, `trinary.True`/`False` (unreachable after the boxing change, handled so the boundary stays total) as `boolean`, and an out-of-range `trinary.Value` fails the call with `trinary value %d cannot cross into JavaScript`. Before, a trinary crossed as a goja-wrapped Go int, which JavaScript saw as a truthy `object`.
- Null, numbers and strings pass through; a host document still gets goja's reflection wrapping, unchanged and out of scope here.
- The return direction needed no change: a JS `boolean` already comes back as `box.Bool`, and JS `undefined` comes back as Sentrie `undefined`, not `unknown`.
- CLI check against a TypeScript module: `kind(false)` is `"boolean:false"`, `kind(unknown)` is `"undefined:undefined"`, `add(true, false)` is `1`.
- Added one minimal unit test for the mapping and the error; the cross-producer matrix and JS boundary tests are the test agent's.
- _Model: claude-opus-5-5_

### 15:56 - Cross-producer boolean matrix and JS and JSON boundary tests
- Added `runtime/bool_representation_test.go` on `RuntimeTestSuite`, three tests that run real policies through `newTempPackExecutor` with a local TypeScript module:
  - Matrix: nine producers (literal, comparison, document fact, `trinary` fact, JS return, `not`, `and`, derive return typed `trinary`, `as trinary` cast), every ordered pair, under `==`, `in [x, 99]` and `[x, 99] contains y`. Each rule yields the four yes/no combinations and the test checks `[true, false, false, true]`; the failure message names the operator and the two producers.
  - JS boundary: `typeof` is `"boolean"` for true and false from every producer, `a ? 1 : 0` is 1/0, unknown from a literal, `not unknown` and `true and unknown` arrives as `undefined` (and picks 0), and a list or dict holding truth values converts recursively.
  - JSON: `json.Marshal` of `ExecPolicy` outputs (the encoder `exec --output json` and the serve API use): `value` is a JSON boolean for `not false`, `not true`, a comparison, a literal and `and`; `state` keeps `"true"`/`"false"`/`"unknown"`; `unknown` keeps the string value.
- Proved they bite: with `box/value.go` from main all three fail (JSON `value` strings, matrix mismatches, JS `object`); with `normalizeBoundaryForJS` returning a trinary unchanged, the JS test fails on the `unknown` cases. Restored both afterwards.
- Ruled out: imports (#136), a standalone `true as number` test (already covered by the changed `eval_cast_test.go`), and per-pair subtests (768 rules; one test with named messages is faster and still pinpoints the pair).
- Syntax notes for future tests: facts must precede `use`, and a dict literal needs quoted keys.
- _Model: claude-sonnet-5-5_

### 02:49 - Merge main after #134 and correct the eval_is node
- #134 landed first. Its `runtime.eval_is` graph node says the negated `is` forms box differently (a `box.Trinary` from `not` against the predicates' `box.Bool`). This branch makes `box.Trinary` box a known outcome as `ValueBool`, so that gotcha is false once both are in.
- Merged `main` in (clean) and rewrote the gotcha: both forms give a bool, and only an operand that errors to `unknown` gives a `ValueTrinary`. Same text as the batch's scratch integration (patch 0003). `make test` and `make docs-graph` pass.
- _Model: claude-opus-5-5_
