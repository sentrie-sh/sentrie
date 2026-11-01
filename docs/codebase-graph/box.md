---
id: box
type: System / Package
language: Go
file_path: box/
tags: value-model, type-system, boundary-marshalling, runtime-data
---

# Node: Box (Universal Boxed Value Model)

## 1. Architectural Role & Intent
`box` defines the single universal runtime value representation for the Sentrie language - a compact tagged union (`Value`) covering undefined, null, bool, number, string, trinary, list, dict, host document, and first-class callable. It exists so that the evaluator, builtins, constraints, and the JavaScript interop layer all speak one value dialect instead of passing raw `any` around, and so that the critical `undefined` vs `null` distinction (missing fact vs explicitly null fact) survives every boundary crossing. It also owns the semantic comparison operators (`EqualValues`, `ContainsValue`, `MatchesValue`) that back the language's infix operators.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `box` | `LAYERED_ON` | [[trinary]] | Boxes `trinary.Value` and projects any boxed value onto the Kleene lattice via `TrinaryFrom`. |
| `box` | `IMPORTS` | `std.encoding/json`, `std.regexp`, `std.math`, `std.strconv`, `std.cmp` | JSON marshalling of values and `json.Number` decoding; `regexp` backs `MatchesValue`; `math.Float64bits` packs floats into the union's `u64` slot; `strconv` parses numbers (`ParseNumber`). |
| [[box.value]] | `DEPENDS_ON` | [[box]] | The `Value` struct and its accessors are defined within this package. |
| [[builtins]] | `LAYERED_ON` | [[box]] | Every builtin signature accepts and returns `box.Value`. |
| [[runtime]] | `LAYERED_ON` | [[box]] | The evaluator's universal expression result type; also uses `Callable` to represent lambdas. |
| [[constraints]] | `LAYERED_ON` | [[box]] | Type/shape constraints validate incoming `box.Value` payloads. |
| [[index]] | `LAYERED_ON` | [[box]] | Static analysis carries literal and builtin-kind information as boxed values. |
| [[runtime.trace]] | `LAYERED_ON` | [[box]] | Trace tree nodes record evaluated `box.Value` results per step. |
| [[cmd]] | `LAYERED_ON` | [[box]] | CLI marshals final decision payloads to JSON via `Value.MarshalJSON`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `EqualValues(a: Value, b: Value) -> bool`
  - **Behavior:** Semantic equality; numbers compare by value across the exact integer and float representations (`5 == 5.0`), exactly (via `CompareNumbers`). Backs the `==` / `!=` infix operators. A `DecisionEnvelope` compared with a bool or trinary compares its state, so an unknown import equals neither `true` nor `false`. Any other kind mismatch is unequal, which is safe for booleans only because a known truth value always has the single kind `ValueBool` (see `Trinary` below).
  - **Side Effects:** None.
  - **Exceptions:** None - incomparable kinds return `false` rather than erroring.

- **Signature:** `ContainsValue(haystack: Value, needle: Value) -> bool`
  - **Behavior:** Backs the `contains` and `in` infix operators for string, list, and dict haystacks. Dict haystacks support only string-key lookup and dict-subset containment.
  - **Side Effects:** None.
  - **Exceptions:** None; unsupported haystack kinds return `false`.

- **Signature:** `MatchesValue(haystack: Value, pattern: Value) -> (bool, error)`
  - **Behavior:** Backs the `matches` operator; both operands must be strings and the pattern is compiled as a Go RE2 regexp.
  - **Side Effects:** Compiles a regexp on every call (no compilation cache).
  - **Exceptions:** Returns an error for non-string operands or an invalid pattern.

- **Signature:** `Add`, `Sub`, `Mul`, `Div`, `Mod(lhs: Value, rhs: Value) -> (Value, error)` / `Negate(v: Value) -> (Value, error)` (`box/number.go`)
  - **Behavior:** The arithmetic behind the infix and unary operators ([#166](https://github.com/sentrie-sh/sentrie/issues/166)). Two exact integers stay exact; any float operand promotes the result to a float. `Div` of two integers gives an integer when the division is exact and a float otherwise (so `7 / 2` is still `3.5`). `Mod` truncates like `math.Mod` (the result takes the dividend's sign).
  - **Side Effects:** None.
  - **Exceptions:** `left operand is not a number` / `right operand is not a number`; `divide by zero` for a zero divisor; an error wrapping `ErrIntegerOverflow` when integer arithmetic leaves the int64 range (including `MinInt64 / -1` and negating `MinInt64`). Overflow never wraps and never silently promotes to float.

- **Signature:** `CompareNumbers(lhs: Value, rhs: Value) -> (cmp: int, ordered: bool, err: error)`
  - **Behavior:** Exact three-way order across the integer and float representations; an int64 is never rounded through float64, so `2^53 + 1 > 2^53` and an integer compares correctly with a float that rounds to it. `ordered` is false when either side is NaN. Backs `<`, `<=`, `>`, `>=`, numeric `EqualValues` and the number constraints.
  - **Side Effects:** None.
  - **Exceptions:** Errors when either operand is non-numeric.

- **Signature:** `ParseNumber(s: string) -> (Value, error)` / `Int(int64)` / `Float(float64)`
  - **Behavior:** `ParseNumber` boxes a decimal that fits int64 as an exact integer and anything else `strconv.ParseFloat` accepts as a float; a magnitude float64 cannot hold returns `Undefined()` with the error. Used by `as number` casts and by `FromAny` for a `json.Number`. `Int` and `Float` box one representation explicitly.

- **Signature:** `CheckJSONNumbers(x: any) -> error`
  - **Behavior:** Walks a tree decoded with `json.Decoder.UseNumber` and reports the first number float64 cannot hold. With `UseNumber` the decoder no longer rejects `1e400`, so every fact decoder ([[cmd]] `exec`, [[api.handle_decision]]) calls this before the facts reach `FromAny`.
  - **Side Effects:** None.
  - **Exceptions:** `number <n> is out of range`.
  - **Side Effects:** None.
  - **Exceptions:** `strconv` errors.

- **Signature:** `ToBoundaryAny(v: Value) -> any` / `TryToBoundaryAny(v: Value) -> (any, error)`
  - **Behavior:** Unboxes a value tree for transport across a non-native boundary (JS interop in [[runtime.js]], module invocation, JSON output) while preserving `undefined` as an opaque sentinel token. `TryToBoundaryAny` is the strict form.
  - **Side Effects:** Allocates a fresh `[]any` / `map[string]any` tree.
  - **Exceptions:** `TryToBoundaryAny` returns `ErrCallableBoundary` if a callable appears anywhere in the tree; `ToBoundaryAny` instead degrades lossily to the string `"<callable>"`.

- **Signature:** `FromBoundaryAny(x: any) -> Value`
  - **Behavior:** Inverse of the above; reconstructs boxed values from boundary representations, restoring `undefined` from its sentinel. A `json.Number` (facts decoded with `json.Decoder.UseNumber`) boxes through `ParseNumber`, so an integral JSON number that fits int64 stays exact. A `Value` is returned as is (an explicit case; `FromAny` always did this), so an in-process caller (a `with` binding) keeps a `DecisionEnvelope` tag; JSON and JS data is never a `Value`, so it can never decode tagged.
  - **Side Effects:** Allocates a boxed tree.
  - **Exceptions:** None - unrecognized types are wrapped as `Document`.

- **Signature:** `IsBoundaryUndefined(x: any) -> bool`
  - **Behavior:** Detects the unexported undefined sentinel on the unboxed side of a boundary.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `TrinaryFrom(b: Value) -> trinary.Value`
  - **Behavior:** Fast Kleene projection that dispatches on the kind tag. Notably `undefined`/`null` → `Unknown`, empty list/dict → `False`, non-empty → `True`, callable → `True`, and a `DecisionEnvelope` → its carried state.
  - **Side Effects:** None; specifically avoids the intermediate allocation that `trinary.From(v.Any())` would incur.
  - **Exceptions:** None.

- **Signature:** `DecisionEnvelope(m: map[string]Value, state: trinary.Value) -> Value` / `(Value).DecisionState() -> (trinary.Value, bool)`
  - **Behavior:** Boxes an imported decision's envelope as an ordinary `ValueDict` whose otherwise unused `u64` slot carries a tag and the state (#136). It indexes, iterates and serializes as a plain dict; only `TrinaryFrom` and `EqualValues` read the tag. `Dict` never sets `u64`, so a user dict with a `state` key cannot be mistaken for an envelope. An out-of-range state is stored as `Unknown`.
  - **Side Effects:** None; no extra allocation.
  - **Exceptions:** None. The tag is per `Value`: a dict rebuilt from the entries, or one that crosses `TryToBoundaryAny`, is untagged; `FromBoundaryAny` of a `Value` keeps it.

- **Signature:** Constructors - `Undefined()`, `Null()`, `Bool[T ~bool]`, `Number[numeric]` (signed integers and unsigned up to `MaxInt64` box as exact integers, larger unsigned and floats as floats), `Int`, `Float`, `String[T ~string]`, `Trinary`, `List`, `Dict`, `Document[T any]`, `Callable`, `FromAny`
  - **Behavior:** Kind-specific boxing. `Trinary` normalises: `trinary.True`/`trinary.False` box as `ValueBool`, and only `trinary.Unknown` (or an out-of-range value, which `trinary` already reads as unknown) boxes as `ValueTrinary` ([#135](https://github.com/sentrie-sh/sentrie/issues/135)). So literals, logical operators, comparisons, facts and JS returns all yield the same boolean. `FromAny(trinary.Value)` goes through `Trinary` too. `Object`/`ObjectRef`/`SameObjectRef` are retained backward-compatible aliases for the `Document` family.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless package; `Value` is an immutable-by-convention struct passed by copy.
- **Performance/Scale Notes:** `Value` is deliberately a compact three-field union (`kind`, `u64`, `ref any`) - scalars (bool, number, trinary) are stored inline in `u64` with **zero heap allocation** (an exact integer's marker in `ref` is a zero-size struct, which does not allocate), while only strings, lists, dicts, documents, and callables touch `ref`. Prefer `TrinaryFrom` over `trinary.From(v.Any())` and prefer typed accessors over `Any()`, since `Any()` deep-copies list/dict trees. `MatchesValue` recompiles its regexp per invocation, so regex-heavy policies pay repeated compilation cost.
- **Dependencies Risk:** No external failure domain. The principal hazards are semantic: (1) `Value` copies share the underlying `ref` slice/map, so mutating a list or dict obtained from `ListValue`/`DictValue` **aliases into every copy** - treat them as read-only; (2) `ToBoundaryAny` silently degrades callables to a placeholder string, so any path that must reject callables at a boundary has to use `TryToBoundaryAny` and honour `ErrCallableBoundary`; (3) `MarshalJSON` renders `undefined` as JSON `null`, erasing the undefined/null distinction at the serialization edge - consumers of [[api]] output cannot recover it; (4) a `ValueTrinary` only ever holds `Unknown`, so code that switches on `ValueTrinary` to find a known boolean finds nothing - read truth values with `TrinaryValue()`, which accepts both kinds, and marshal `Unknown` as the JSON string `"unknown"`. (5) `NumberValue()` returns a float64 for both number representations, so it rounds an exact integer above 2^53; anything that compares, divides or tests parity must use `IntValue`, `CompareNumbers` or the arithmetic helpers instead. `Any()` returns `int64` for an exact integer, so host code that asserted `.(float64)` on every number must accept `int64` too.
