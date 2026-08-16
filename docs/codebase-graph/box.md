---
id: box
type: System / Package
language: Go
file_path: box/
tags: value-model, type-system, boundary-marshalling, runtime-data
---

# Node: Box (Universal Boxed Value Model)

## 1. Architectural Role & Intent
`box` defines the single universal runtime value representation for the Sentrie language — a compact tagged union (`Value`) covering undefined, null, bool, number, string, trinary, list, dict, host document, and first-class callable. It exists so that the evaluator, builtins, constraints, and the JavaScript interop layer all speak one value dialect instead of passing raw `any` around, and so that the critical `undefined` vs `null` distinction (missing fact vs explicitly null fact) survives every boundary crossing. It also owns the semantic comparison operators (`EqualValues`, `ContainsValue`, `MatchesValue`) that back the language's infix operators.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `box` | `LAYERED_ON` | [[trinary]] | Boxes `trinary.Value` and projects any boxed value onto the Kleene lattice via `TrinaryFrom`. |
| `box` | `IMPORTS` | `std.encoding/json`, `std.regexp`, `std.math` | JSON marshalling of values; `regexp` backs `MatchesValue`; `math.Float64bits` packs numbers into the union's `u64` slot. |
| [[box.value]] | `DEPENDS_ON` | [[box]] | The `Value` struct and its accessors are defined within this package. |
| [[builtins]] | `LAYERED_ON` | [[box]] | Every builtin signature accepts and returns `box.Value`. |
| [[runtime]] | `LAYERED_ON` | [[box]] | The evaluator's universal expression result type; also uses `Callable` to represent lambdas. |
| [[constraints]] | `LAYERED_ON` | [[box]] | Type/shape constraints validate incoming `box.Value` payloads. |
| [[index.package]] | `LAYERED_ON` | [[box]] | Static analysis carries literal and builtin-kind information as boxed values. |
| [[runtime.trace]] | `LAYERED_ON` | [[box]] | Trace tree nodes record evaluated `box.Value` results per step. |
| [[cmd]] | `LAYERED_ON` | [[box]] | CLI marshals final decision payloads to JSON via `Value.MarshalJSON`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `EqualValues(a: Value, b: Value) -> bool`
  - **Behavior:** Semantic equality including cross-kind numeric comparison. Backs the `==` / `!=` infix operators.
  - **Side Effects:** None.
  - **Exceptions:** None — incomparable kinds return `false` rather than erroring.

- **Signature:** `ContainsValue(haystack: Value, needle: Value) -> bool`
  - **Behavior:** Backs the `contains` and `in` infix operators for string, list, and dict haystacks. Dict haystacks support only string-key lookup and dict-subset containment.
  - **Side Effects:** None.
  - **Exceptions:** None; unsupported haystack kinds return `false`.

- **Signature:** `MatchesValue(haystack: Value, pattern: Value) -> (bool, error)`
  - **Behavior:** Backs the `matches` operator; both operands must be strings and the pattern is compiled as a Go RE2 regexp.
  - **Side Effects:** Compiles a regexp on every call (no compilation cache).
  - **Exceptions:** Returns an error for non-string operands or an invalid pattern.

- **Signature:** `MustNumbers(lhs: Value, rhs: Value) -> (float64, float64, error)`
  - **Behavior:** Coerces both operands for arithmetic and relational operators.
  - **Side Effects:** None.
  - **Exceptions:** Errors when either operand is non-numeric; surfaced by [[runtime]] as a typed evaluation failure.

- **Signature:** `ToBoundaryAny(v: Value) -> any` / `TryToBoundaryAny(v: Value) -> (any, error)`
  - **Behavior:** Unboxes a value tree for transport across a non-native boundary (JS interop in [[runtime.js]], module invocation, JSON output) while preserving `undefined` as an opaque sentinel token. `TryToBoundaryAny` is the strict form.
  - **Side Effects:** Allocates a fresh `[]any` / `map[string]any` tree.
  - **Exceptions:** `TryToBoundaryAny` returns `ErrCallableBoundary` if a callable appears anywhere in the tree; `ToBoundaryAny` instead degrades lossily to the string `"<callable>"`.

- **Signature:** `FromBoundaryAny(x: any) -> Value`
  - **Behavior:** Inverse of the above; reconstructs boxed values from boundary representations, restoring `undefined` from its sentinel.
  - **Side Effects:** Allocates a boxed tree.
  - **Exceptions:** None — unrecognized types are wrapped as `Document`.

- **Signature:** `IsBoundaryUndefined(x: any) -> bool`
  - **Behavior:** Detects the unexported undefined sentinel on the unboxed side of a boundary.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `TrinaryFrom(b: Value) -> trinary.Value`
  - **Behavior:** Fast Kleene projection that dispatches on the kind tag. Notably `undefined`/`null` → `Unknown`, empty list/dict → `False`, non-empty → `True`, callable → `True`.
  - **Side Effects:** None; specifically avoids the intermediate allocation that `trinary.From(v.Any())` would incur.
  - **Exceptions:** None.

- **Signature:** Constructors — `Undefined()`, `Null()`, `Bool[T ~bool]`, `Number[numeric]`, `String[T ~string]`, `Trinary`, `List`, `Dict`, `Document[T any]`, `Callable`, `FromAny`
  - **Behavior:** Kind-specific boxing. `Object`/`ObjectRef`/`SameObjectRef` are retained backward-compatible aliases for the `Document` family.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless package; `Value` is an immutable-by-convention struct passed by copy.
- **Performance/Scale Notes:** `Value` is deliberately a compact three-field union (`kind`, `u64`, `ref any`) — scalars (bool, number, trinary) are stored inline in `u64` with **zero heap allocation**, while only strings, lists, dicts, documents, and callables touch `ref`. Prefer `TrinaryFrom` over `trinary.From(v.Any())` and prefer typed accessors over `Any()`, since `Any()` deep-copies list/dict trees. `MatchesValue` recompiles its regexp per invocation, so regex-heavy policies pay repeated compilation cost.
- **Dependencies Risk:** No external failure domain. The principal hazards are semantic: (1) `Value` copies share the underlying `ref` slice/map, so mutating a list or dict obtained from `ListValue`/`DictValue` **aliases into every copy** — treat them as read-only; (2) `ToBoundaryAny` silently degrades callables to a placeholder string, so any path that must reject callables at a boundary has to use `TryToBoundaryAny` and honour `ErrCallableBoundary`; (3) `MarshalJSON` renders `undefined` as JSON `null`, erasing the undefined/null distinction at the serialization edge — consumers of [[api]] output cannot recover it.
