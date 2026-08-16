---
id: runtime.eval_infix
type: Function / Endpoint
language: Go
file_path: runtime/eval_infix.go
tags: operators, arithmetic, kleene-logic, comparison, coercion
---

# Node: runtime.evalInfix (Binary Operators)

## 1. Architectural Role & Intent
Implements every binary operator in one switch: arithmetic, comparison, three-valued logic, membership, and regex matching. It evaluates both operands eagerly and applies a blanket undefined-propagation rule before dispatching, which makes the whole operator surface uniformly tolerant of missing data.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_infix` | `DEPENDS_ON` | [[box]] | `MustNumbers`, `EqualValues`, `ContainsValue`, `MatchesValue`, `TrinaryFrom`, and the constructors. |
| `runtime.eval_infix` | `DEPENDS_ON` | [[trinary]] | `And`, `Or`, `Not` implement Kleene logic; `xor` is composed from them. |
| `runtime.eval_infix` | `CALLS` | [[runtime.eval]] | Evaluates left then right operand. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_infix]] | All `ast.InfixExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalInfix(ctx, ec, exec, p, in *ast.InfixExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Evaluates both operands, returns `Undefined` if **either** is undefined, then dispatches on the operator string.
  - **Side Effects:** Whatever the operands do.
  - **Exceptions:** `divide by zero`; numeric coercion failures from `box.MustNumbers`; regex errors; `unsupported infix op: %s`.

- **Signature:** Arithmetic - `+`, `-`, `*`, `/`, `%`
  - **Behavior:** `+` is overloaded: if **either** side is a string it concatenates, rendering the other side via `String()`. The rest require numbers. `/` and `%` reject a zero divisor; `%` uses `math.Mod`, so it is float modulo, not integer remainder.
  - **Side Effects:** None.
  - **Exceptions:** `divide by zero`; coercion failure.

- **Signature:** Comparison - `==`, `is`, `!=`, `<`, `<=`, `>`, `>=`
  - **Behavior:** `==` and `is` are **the same operation**, both delegating to `box.EqualValues`. The four relational operators require numbers, so strings and dates cannot be ordered.
  - **Side Effects:** None.
  - **Exceptions:** Coercion failure on the relational forms.

- **Signature:** Logic - `and`, `or`, `xor`
  - **Behavior:** Coerces both sides with `box.TrinaryFrom` and applies Kleene operators. `xor` is derived as `(l or r) and not (l and r)`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** Membership and matching - `in`, `contains`, `matches`
  - **Behavior:** `in` and `contains` are the same predicate with operands swapped. `matches` applies a regex from the right operand to the left.
  - **Side Effects:** None.
  - **Exceptions:** Regex compilation or application errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Both operands are **always evaluated** - there is no short-circuiting for `and` or `or`. An expensive right operand runs even when the left already determines the result, and any side effect it has (a module call, a rule reference) happens regardless.
- **Dependencies Risk:**
  - **Undefined propagation overrides Kleene logic.** The blanket check runs before the switch, so `false and <undefined>` yields `Undefined` rather than `False`, even though Kleene logic defines that case. The three-valued semantics the language advertises are therefore not complete at the operator level - `Undefined` is a fourth state that absorbs everything.
  - **No short-circuit evaluation** means `a is defined and a.field > 0` still evaluates the right side. Guard patterns that look safe are not.
  - **`+` silently stringifies.** `1 + "a"` produces `"1a"` rather than a type error, because the string branch is checked before the numeric one. Combined with the lack of static type checking on operators, arithmetic typos degrade into concatenation.
  - **`==` and `is` are indistinguishable at runtime**, so the two spellings are pure syntax preference - but [[parser.precedence]] gives them different precedence entries, which can change grouping.
  - **`%` is floating-point modulo.** Since integers and floats collapse to one numeric kind in [[box]], `%` on values that look like integers still goes through `math.Mod`.
  - **The `default` branch is unreachable in practice** but is the only guard against [[parser.lookups]] registering an operator this switch does not implement - a mismatch would surface at request time.
