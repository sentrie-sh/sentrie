---
id: trinary
type: System / Package
language: Go
file_path: trinary/
tags: kleene-logic, three-valued-logic, policy-semantics, decision-model
---

# Node: Trinary (Kleene Three-Valued Logic Core)

## 1. Architectural Role & Intent
`trinary` implements the three-valued (Kleene) logic that is the semantic heart of Sentrie: a policy outcome is `True`, `False`, or `Unknown`, where `Unknown` propagates through boolean algebra rather than collapsing to false. It exists so that a policy evaluated against incomplete facts yields an explicit "cannot decide" verdict instead of a silently wrong allow/deny. Every rule evaluation, decision export, and boolean infix operator in the runtime ultimately funnels through this package's `And`/`Or`/`Not` truth tables.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `trinary` | `LAYERED_ON` | [[tokens]] | `FromToken` reads `tokens.Instance` to map `true`/`false`/`unknown` keyword tokens into `Value`. |
| [[box]] | `LAYERED_ON` | [[trinary]] | `box.Trinary` boxes a `trinary.Value`; `box.TrinaryFrom` derives the Kleene outcome of any boxed value. |
| [[ast]] | `LAYERED_ON` | [[trinary]] | Trinary literal AST nodes carry a resolved `trinary.Value`. |
| [[parser]] | `LAYERED_ON` | [[trinary]] | Parses trinary literal tokens into typed values at parse time. |
| [[runtime]] | `LAYERED_ON` | [[trinary]] | Boolean infix evaluation, `yield` semantics, and decision resolution delegate to `And`/`Or`/`Not`. |
| [[constraints]] | `LAYERED_ON` | [[trinary]] | Trinary type constraints validate against `trinary.Value`. |
| [[cmd]] | `LAYERED_ON` | [[trinary]] | CLI renders the final decision verdict and maps it to a process exit disposition. |
| [[runtime.decision]] | `CALLS` | [[trinary]] | Aggregates rule outcomes into the exported decision value. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Value` (enum: `False = -1`, `Unknown = 0`, `True = 1`)
  - **Behavior:** The tri-state outcome type. The `iota - 1` encoding makes `Unknown` the zero value, so an uninitialized `Value` is safely `Unknown` rather than accidentally `False`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `And(other: Value) -> Value`
  - **Behavior:** Kleene conjunction. `False` is absorbing (`Unknown and False == False`); `Unknown and True == Unknown`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Or(other: Value) -> Value`
  - **Behavior:** Kleene disjunction. `True` is absorbing (`Unknown or True == True`); `Unknown or False == Unknown`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Not() -> Value`
  - **Behavior:** Negation with `Unknown` as a fixed point (`Not(Unknown) == Unknown`).
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `From(v: any) -> Value`
  - **Behavior:** Universal coercion ladder: `nil`/`IsUndefined` → `Unknown`; `HasTrinary` delegates; `bool` maps directly; numerics are truthy when non-zero; strings try keyword parsing first then fall back to non-empty; pointers/interfaces dereference once and retry; reflected slice/array/map use length; all other values default to `True`.
  - **Side Effects:** Uses reflection on the fallback path.
  - **Exceptions:** None - the function is total and never errors, which means genuinely unmappable values silently become `True`.

- **Signature:** `Parse(s: string) -> Value`
  - **Behavior:** Textual parse accepting `1/t/T/TRUE/true/True`, `0/f/F/FALSE/false/False`, `-1/n/N/UNKNOWN/unknown/Unknown`.
  - **Side Effects:** None.
  - **Exceptions:** None - unrecognized input silently returns `Unknown`.

- **Signature:** `FromToken(t: tokens.Instance) -> Value`
  - **Behavior:** Converts a trinary keyword token to its value during parsing.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `MarshalJSON() -> ([]byte, error)` / `String() -> string` / `IsTrue() -> bool` / `Equals(other: Value) -> bool`
  - **Behavior:** Serialization and comparison surface. `IsTrue` is a strict identity check against `True` - it is **not** a truthiness coercion, so `Unknown` is not true.
  - **Side Effects:** None.
  - **Exceptions:** `MarshalJSON` errors are inherited from the encoder only.

- **Signature:** `HasTrinary` interface - `ToTrinary() -> Value`
  - **Behavior:** Extension point letting foreign types define their own Kleene projection, consumed by `From`.
  - **Side Effects:** Implementation-defined.
  - **Exceptions:** Implementation-defined.

- **Signature:** `IsUndefined` interface - `IsUndefined() -> bool`
  - **Behavior:** Lets a type declare itself absent, forcing `From` to yield `Unknown`. This is how [[box]]'s undefined sentinel keeps missing facts from collapsing to `False`.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; `Value` is an immutable integer-backed value type.
- **Performance/Scale Notes:** `And`/`Or`/`Not` are branch-only integer operations and are effectively free. `From` is the expensive path because it falls back to reflection; hot loops should prefer `box.TrinaryFrom`, which short-circuits on the boxed kind tag and avoids materializing intermediate `[]any`/`map[string]any`.
- **Dependencies Risk:** No external failure domain. The correctness risk is semantic: because both `From` and `Parse` are **total functions that never signal failure**, malformed input degrades to `True` (for `From`) or `Unknown` (for `Parse`) instead of raising. A policy fed a garbage fact can therefore produce a confident `True` verdict rather than an error - validate inputs upstream in [[constraints]] and [[index.validate]], not here.
