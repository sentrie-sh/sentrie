---
id: box.set
type: Module / File
language: Go
file_path: box/set.go
tags: value-model, sets, canonical-order, collections
---

# Node: box Sets (Tagged Canonical Lists)

## 1. Architectural Role & Intent
`box/set.go` gives Sentrie its `set[T]` value ([#162](https://github.com/sentrie-sh/sentrie/issues/162)). A set is **not a new kind**: it is a `ValueList` that carries a set tag in the otherwise unused `u64` slot (the same trick as the decision envelope on dicts), with elements that are distinct and held in one canonical order. Choosing a tagged list over a distinct kind means every list-reading path - `in`, `contains`, `count`, indexing, iteration, JSON output, the JS boundary - reads a set unchanged, and a set always serialises as a sorted array, so output is deterministic. The file also owns the set algebra the builtins expose (union, intersection, difference, subset), done as linear merges over the sorted elements.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `box.set` | `DEPENDS_ON` | [[box.value]] | Builds tagged `ValueList` values and reads the typed accessors; strips a decision envelope's tag from each element. |
| `box.set` | `IMPORTS` | `std.cmp`, `std.slices`, `std.strings`, `std.math` | Canonical ordering, sorting, binary search, NaN detection. |
| [[box]] | `DEPENDS_ON` | [[box.set]] | `ContainsValue` binary-searches a set for a scalar needle. |
| [[builtins]] | `CALLS` | [[box.set]] | `set`, `union`, `intersection`, `difference`, `is_subset`, `is_superset` are thin wrappers: they convert each argument with `ToSet` (so an element error names its argument), then call `NewSet` or a `Set*` operation. |
| [[runtime.typeref_set]] | `CALLS` | [[box.set]] | `NewSetUnique` validates a plain list against `set[T]` and converts a bound list into a set. |
| [[runtime.eval_cast]] | `CALLS` | [[box.set]] | `as set[T]` converts a list with `NewSet`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewSet(xs []Value) -> (Value, error)` / `NewSetUnique(xs []Value) -> (Value, error)`
  - **Behavior:** Copy, validate, canonicalise, sort into canonical order. `NewSet` drops duplicates; `NewSetUnique` fails on the first one (`duplicate set element <v>`). Canonicalising rewrites every element at every depth: a decision envelope is stored as the plain dict it reads as, and `-0` as `0`, so the output does not depend on input order. A nested list keeps its set tag.
  - **Side Effects:** None; the input slice is never modified.
  - **Exceptions:** An element with no set order: `a set cannot hold a value of kind <kind>` (undefined, document, callable), `a set cannot hold an unknown value`, `a set cannot hold NaN`, checked recursively inside list and dict elements.

- **Signature:** `ToSet(v Value) -> (Value, error)` / `(Value).IsSet() -> bool`
  - **Behavior:** `ToSet` returns a set unchanged and builds one from a plain list with `NewSet`; any other kind is `value of kind <kind> is not a set or list`. `IsSet` reads the tag.
  - **Side Effects:** None.
  - **Exceptions:** As above.

- **Signature:** `SetUnion(a, b)`, `SetIntersection(a, b)`, `SetDifference(a, b) -> (Value, error)`; `SetIsSubset(a, b) -> (bool, error)`
  - **Behavior:** Both operands must be sets (convert with `ToSet`); the result is a set. Linear merges over the canonical order.
  - **Side Effects:** None.
  - **Exceptions:** `first operand is not a set` / `second operand is not a set`.

## 4. Operational Context & Gotchas
- **Statefulness:** Value semantics like every `box.Value`; the element slice is shared by copies and must be treated as immutable.
- **Performance/Scale Notes:** Construction is O(n log n); membership of a scalar needle is O(log n); the operations are O(n + m). A dict element's comparison sorts its keys on every call, so large sets of dicts sort slowly.
- **Dependencies Risk:**
  - **The canonical order must agree with `EqualValues`.** Order is by kind (null, bool, number, string, list, dict), then by value; lists compare element-wise then by length, dicts entry by entry in key order then by size. For the admitted elements this returns 0 exactly when `EqualValues` holds (so `-0` and `0` are one element). The one exception, a decision envelope equal to a bool by state (also inside a list, where `[impT]` equals `[true]`), is why `NewSet` strips the envelope tag at every depth, and why the binary-search fast path in `ContainsValue` is limited to scalar needles.
  - **The tag does not survive a rebuild.** `map`, `filter`, a module return, or a JSON round trip yields a plain list; only the set builtins, a `set[T]` binding, and `as set[T]` produce a set.
  - **A set equals a list with the same elements in the same order** under `EqualValues`, since both are lists; `set(["b", "a"]) == ["a", "b"]` holds and `== ["b", "a"]` does not.
