---
id: box.value
type: Class
language: Go
file_path: box/value.go
tags: value-model, tagged-union, type-system, memory-layout
---

# Node: box.Value (Tagged Union Value Type)

## 1. Architectural Role & Intent
`box.Value` is the concrete tagged-union struct that every Sentrie expression evaluates to. It encodes eleven value kinds in three fields - a `ValueKind` tag, a `uint64` inline payload for scalars, and an `any` reference for heap-backed kinds - so that the evaluator can pass values by copy without allocating for numbers, booleans, or trinary states. Its defining responsibility is preserving the four-way distinction between *invalid*, *undefined*, *null*, and a real value, which is what allows Sentrie to report "I could not decide" rather than guessing.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `box.value` | `DEPENDS_ON` | [[trinary]] | `TrinaryValue()` unpacks the `u64` slot into a `trinary.Value`; `TrinaryFrom` projects any kind onto the Kleene lattice. |
| `box.value` | `DEPENDS_ON` | [[box.set]] | `IsSet()` reads the set tag that `box/set.go` stores in a list's `u64` slot; a set has kind `ValueList` (#162). |
| `box.value` | `DEPENDS_ON` | [[box]] | Member of the `box` package; shares the boundary-marshalling helpers in `box/utilities.go`. |
| [[runtime.eval]] | `CALLS` | [[box.value]] | Every expression evaluation constructs and destructures `Value` instances. |
| [[builtins]] | `CALLS` | [[box.value]] | Builtin implementations use the typed accessors to validate argument kinds. |
| [[constraints]] | `CALLS` | [[box.value]] | Constraint checks read `Kind()` and the typed accessors to enforce declared types. |
| [[runtime.js]] | `CALLS` | [[box.value]] | JS interop converts to/from boundary representations through `ToBoundaryAny`/`FromBoundaryAny`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Kind() -> ValueKind`
  - **Behavior:** Returns the union discriminant. This is the canonical dispatch point - always branch on `Kind()` rather than type-asserting `Any()`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `IsInt() -> bool` / `IntValue() -> (int64, bool)` / `IntegralValue() -> (int64, bool)`
  - **Behavior:** A number is one kind (`ValueNumber`) with two representations ([#166](https://github.com/sentrie-sh/sentrie/issues/166)): an exact integer (int64 bits in `u64`, a zero-size `exactIntTag` in `ref`) or a float (float64 bits in `u64`, `ref` nil). `IntValue` reports only an exact integer; `IntegralValue` also accepts a float with no fractional part inside the int64 range (used for `distinct` keys so `5` and `5.0` collide).
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `BoolValue() -> (bool, bool)` / `NumberValue() -> (float64, bool)` / `StringValue() -> (string, bool)` / `TrinaryValue() -> (trinary.Value, bool)` / `ListValue() -> ([]Value, bool)` / `DictValue() -> (map[string]Value, bool)`
  - **Behavior:** Kind-checked typed accessors following the comma-ok idiom; the boolean is `false` when the tag does not match, and the value slot is zero. `TrinaryValue` is the one accessor that spans two kinds: it reports `True`/`False` for a `ValueBool` and `Unknown` for a `ValueTrinary`, since a known truth value is always boxed as a bool ([#135](https://github.com/sentrie-sh/sentrie/issues/135)).
  - **Side Effects:** None. `ListValue`/`DictValue` return the **live backing container**, not a copy. `NumberValue` converts an exact integer to float64, which rounds above 2^53.
  - **Exceptions:** None - never panics on kind mismatch.

- **Signature:** `DocumentRef() -> (any, bool)` / `CallableRef() -> (any, bool)`
  - **Behavior:** Unwraps the opaque host payload. `CallableRef` returns a reference that is deliberately opaque here and interpreted only by [[runtime.callable]], keeping `box` free of any dependency on the evaluator.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `SameDocumentRef(other: Value) -> bool`
  - **Behavior:** Reference-identity comparison for two `Document` values, used where structural equality is inappropriate or too expensive.
  - **Side Effects:** None.
  - **Exceptions:** None. Returns `false` if either operand is not a `Document`.

- **Signature:** `IsValid() -> bool` / `IsUndefined() -> bool` / `IsNull() -> bool` / `IsCallable() -> bool`
  - **Behavior:** Tag predicates. `IsValid()` distinguishes the zero-value `ValueInvalid` (a `Value{}` that was never constructed) from a legitimately `Undefined` value.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Any() -> any`
  - **Behavior:** Fully unboxes into plain Go types for host consumption; lists and dicts are recursively converted. An exact integer unboxes as `int64`, a float as `float64`.
  - **Side Effects:** **Deep-copies** list and dict trees into fresh `[]any` / `map[string]any`. Callables degrade to the placeholder string `"<callable>"`. Both `Undefined` and `Null` return Go `nil`, erasing the distinction.
  - **Exceptions:** None.

- **Signature:** `String() -> string`
  - **Behavior:** Human-readable rendering used in diagnostics and CLI output; an exact integer renders as plain digits; non-scalar kinds fall through to `fmt.Sprintf("%v", Any())`.
  - **Side Effects:** Triggers the deep copy in `Any()` for lists and dicts.
  - **Exceptions:** None.

- **Signature:** `MarshalJSON() -> ([]byte, error)`
  - **Behavior:** JSON serialization for API and CLI output. `Undefined` is emitted as JSON `null`. An exact integer is emitted as plain digits, never exponent notation. A callable, at any depth, is emitted as `{"kind":"callable"}` (not the `"<callable>"` string `Any()` gives), so a decision or trace holding a lambda still encodes ([#141](https://github.com/sentrie-sh/sentrie/issues/141)).
  - **Side Effects:** Deep copy via the unexported `jsonAny()`, which is `Any()` except for callables.
  - **Exceptions:** Propagates `encoding/json` errors (for example a non-finite number).

- **Signature:** `IsSet() -> bool`
  - **Behavior:** Reports whether a `ValueList` carries the set tag; see [[box.set]]. A set is not a separate kind, so `Kind()` is `ValueList` and `ListValue()` returns its distinct, canonically sorted elements.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `ValueKind` enum - `ValueInvalid`, `ValueUndefined`, `ValueNull`, `ValueBool`, `ValueNumber`, `ValueString`, `ValueTrinary`, `ValueList`, `ValueDict`, `ValueDocument`, `ValueCallable` (with `ValueObject` aliasing `ValueDocument`)
  - **Behavior:** The closed kind set. `ValueInvalid` is the zero value, so an unassigned `Value` is detectably invalid rather than masquerading as undefined.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Value type, copied by assignment; there is no identity or lifecycle. It is safe to share across goroutines **only** for scalar kinds - see aliasing note below.
- **Performance/Scale Notes:** Numbers live in `u64` either as int64 bits (exact integer) or as `math.Float64bits` (float); the `ref` slot tells them apart. Integer literals, integral JSON facts that fit int64 and Go integers are exact; integers beyond int64 fall back to float. Scalar construction and access are allocation-free; `Any()`, `String()`, and `MarshalJSON()` are the allocating paths and should be kept out of evaluation hot loops.
- **Dependencies Risk:** The dominant hazard is **shared mutable backing storage**: `ListValue()` and `DictValue()` hand back the same slice/map that every copy of the `Value` references, so an in-place write is visible through all copies and across evaluation frames. Treat returned containers as immutable and rebuild via `List`/`Dict` when a modification is needed. Secondarily, `ValueInvalid` (zero value) reaching the evaluator usually indicates a missing constructor call upstream rather than a legitimate undefined fact - check `IsValid()` when diagnosing "invalid" in output.
