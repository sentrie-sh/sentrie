---
id: runtime.decision
type: Class
language: Go
file_path: runtime/decision.go
tags: decision-model, trinary, coercion, output-contract
---

# Node: runtime.Decision (Decision Envelope)

## 1. Architectural Role & Intent
The output contract of a rule: a Kleene `State` paired with the underlying `Value` that produced it. `DecisionOf` is the single coercion point where an arbitrary evaluated value becomes a policy outcome, and its bias is deliberate — anything absent or unrepresentable becomes `Unknown` rather than `False`, so a failure never masquerades as a definitive allow or deny.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.decision` | `DEPENDS_ON` | [[box]] | Inspects `IsUndefined`, `IsNull`, `TrinaryValue`, and falls back to `box.TrinaryFrom`. |
| `runtime.decision` | `DEPENDS_ON` | [[trinary]] | `trinary.Value` is the state type; `Unknown` is the failure-safe default. |
| [[runtime.executor]] | `CALLS` | [[runtime.decision]] | Wraps rule bodies, `default` clauses, and the error fallback. |
| [[runtime.imports]] | `READS_FROM` | [[runtime.decision]] | Flattens `State` and `Value` into the import envelope dictionary. |
| [[cmd]] | `READS_FROM` | [[runtime.decision]] | Renders the decision as CLI output. |
| [[api]] | `READS_FROM` | [[runtime.decision]] | Serializes it as the JSON response payload. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Decision` — `{ State trinary.Value \`json:"state"\`, Value box.Value \`json:"value"\` }`
  - **Behavior:** Both fields are serialized, so consumers see the outcome **and** the value it came from — a `False` from a boolean and a `False` derived from an empty list are distinguishable downstream.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `(Decision).ToTrinary() -> trinary.Value`
  - **Behavior:** Returns `State`. A **value receiver**, so it works on both values and pointers.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `DecisionAttachments` — `map[string]box.Value`
  - **Behavior:** The named metadata computed from a rule export's `attach` clauses.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `DecisionOf(val box.Value) -> *Decision`
  - **Behavior:** Three-way coercion. Undefined or null yields `Unknown` while **preserving the original value**; an already-trinary value passes its state through; anything else is coerced via `box.TrinaryFrom`, which applies the language's truthiness rules.
  - **Side Effects:** None.
  - **Exceptions:** None — it is a total function.

## 4. Operational Context & Gotchas
- **Statefulness:** Immutable value type.
- **Performance/Scale Notes:** Allocation-only; one small struct per rule outcome.
- **Dependencies Risk:**
  - **Coercion is total and therefore silent.** `DecisionOf` cannot fail, so a rule body that returns a string, a list, or a dictionary still produces a decision via truthiness rather than a type error. Whether that value *should* be a decision is not checked anywhere — [[index.validate]] does not constrain rule body types either.
  - **`Unknown` is overloaded.** It means "the guard was not true", "the value was null or undefined", and "evaluation failed" (via the executor's error fallback). Consumers cannot distinguish these from the `State` alone; the `Value` field and the accompanying error are the only discriminators.
  - **`ExecutorOutput.ToTrinary()` dereferences `Decision` without a nil check**, so a caller holding a partially-constructed output can panic — the executor always populates it, but nothing in the type enforces that.
  - **Truthiness semantics live in [[box]], not here.** Changing `box.TrinaryFrom` silently changes what every non-trinary rule body decides.
