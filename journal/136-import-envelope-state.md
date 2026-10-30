## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #136
- An imported decision evaluates to a non-empty dict envelope, which is always truthy, so imports fail open.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:49 - Tag the import envelope and read its state as a truth value
- Marked the envelope with `box.DecisionEnvelope(m, state)`: a normal `ValueDict` whose `u64` slot (unused by dicts) holds a tag bit plus the state offset by one. `(Value).DecisionState()` reads it back.
- Why this representation: no new `ValueKind`, so every `DictValue`, `.field`, iteration, `contains`, JSON and boundary path keeps working unchanged; no struct growth and no extra allocation. `box.Dict` never sets `u64`, so a user dict `{"state": false}` cannot carry the tag and keeps ordinary dict truthiness. Ruled out a new kind (every dict switch in box, runtime, builtins and js would need a second case) and sniffing a `state` key (forgeable by user data).
- Truth positions: `box.TrinaryFrom` returns the tagged state, which covers the rule's own decision (`DecisionOf`), `and`, `or`, `xor`, `not`, ternary conditions, `when` gates, `as bool`/`as trinary` casts and the collection predicate builtins, since all of them go through it. An unknown import stays `Unknown`.
- `==`: `EqualValues` compares a tagged envelope with a bool or trinary by state. So `imp == true` is true only for a true import, `imp == false` only for a false one, and an unknown import equals neither. Kept dict equality was rejected because `imp != false` would then be true for every import, a fail-open guard. `!=` stays the plain negation, so `imp != false` is true for an unknown import, the same as for a plain `unknown`.
- Exported value: inside the policy a bare-import rule keeps the envelope as its value, so `imp.state`, `imp.value` and `imp.<attachment>` still work through rule references. `ExecRule` reports `decision.reported()`, which unwraps the envelope to its `value`, so `export decision of imp` reports the imported decision (state and value agree) and nested imports stay flat. Imported attachments are not re-exported unless the importing rule attaches them.
- Limits: the tag is per `Value`; a dict rebuilt from the entries, or one that crosses `TryToBoundaryAny` (a `with` fact, a JS call), is a plain dict again. `let x: bool = imp` still fails the bool type check (fail closed).
- Independent of #135: the tag carries a `trinary.Value`, and equality accepts either a bool or a trinary operand, so it holds whether known booleans are `ValueBool` or `ValueTrinary`. Did not touch the `with` loop (#151).
- Verified with a scratch pack: false and unknown imports decide false and unknown directly and under `and`/`or`/`not`/ternary, and a user dict with a `state` key stays truthy. Added only minimal box and runtime tests; the regression matrix is the test agent's.
- _Model: claude-opus-5-5_
