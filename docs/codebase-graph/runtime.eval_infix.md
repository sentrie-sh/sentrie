---
id: runtime.eval_infix
type: Function / Endpoint
language: Go
file_path: runtime/eval_infix.go
tags: operators, arithmetic, kleene-logic, comparison, coercion
---

# Node: runtime.evalInfix (Binary Operators)

## 1. Architectural Role & Intent
Implements every binary operator in one switch: arithmetic, comparison, three-valued logic, membership, and regex matching. The logic operators are routed first to `evalLogical`, which applies Kleene three-valued logic with an undefined operand read as unknown, and short-circuits `and`/`or`. Every other operator evaluates both operands eagerly and applies a blanket undefined-propagation rule before dispatching, which makes the rest of the operator surface uniformly tolerant of missing data.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_infix` | `DEPENDS_ON` | [[box]] | `Add`, `Sub`, `Mul`, `Div`, `Mod`, `CompareNumbers`, `StringValue`, `EqualValues`, `ContainsValue`, `MatchesValue`, `TrinaryFrom`, and the constructors. |
| `runtime.eval_infix` | `DEPENDS_ON` | [[trinary]] | `And`, `Or`, `Not` implement Kleene logic; `xor` is composed from them. |
| `runtime.eval_infix` | `CALLS` | [[runtime.eval]] | Evaluates left then right operand; for `and`/`or` the right only when the left does not decide the result. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_infix]] | All `ast.InfixExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalInfix(ctx, ec, exec, p, in *ast.InfixExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Sends `and`, `or` and `xor` to `evalLogical`. For every other operator it evaluates both operands, returns `Undefined` if **either** is undefined, then dispatches on the operator string.
  - **Side Effects:** Whatever the operands do.
  - **Exceptions:** `divide by zero`; operand and overflow errors from the `box` arithmetic helpers (`ErrIntegerOverflow`); regex errors; `unsupported infix op: %s`.

- **Signature:** Arithmetic - `+`, `-`, `*`, `/`, `%`
  - **Behavior:** `+` is overloaded: if **either** side is a string it concatenates, rendering the other side via `String()`. The rest require numbers and delegate to `box.Add`/`Sub`/`Mul`/`Div`/`Mod`: two exact integers stay exact and overflow is an error; a float operand makes the result a float. `/` and `%` reject a zero divisor; `/` of two integers is an integer only when exact (`7 / 2` is `3.5`).
  - **Side Effects:** None.
  - **Exceptions:** `divide by zero`; operand not a number; integer overflow.

- **Signature:** Comparison - `==`, `is`, `!=`, `<`, `<=`, `>`, `>=`
  - **Behavior:** `==` and the infix `is` are **the same operation**, both delegating to `box.EqualValues`. Only an `is` whose right-hand side is neither type syntax nor a bare or qualified name reaches here as an infix; `x is T` and `x is Name` are `ast.IsExpression` and go to [[runtime.eval_is_type]], whose `evalIsEquality` repeats this equality when the index resolved the name to a value; the two must stay in step. The four relational operators order two strings by byte order (`compareStrings`, Go's `strings.Compare`), which is not locale-aware: `"Z" < "a"` and `"10" < "9"`. The string branch is tried first and only when both operands are strings; any other pair, including a string and a number, falls through to `box.CompareNumbers` and keeps its operand error. `box.CompareNumbers` orders integers above 2^53 exactly and makes every comparison with NaN false. Dates have no kind of their own, so ISO-8601 strings order chronologically and other formats do not.
  - **Side Effects:** None.
  - **Exceptions:** Coercion failure on the relational forms when the operands are not both strings.

- **Signature:** Logic - `and`, `or`, `xor`
  - **Behavior:** `evalLogical` evaluates the left side and coerces it with `box.TrinaryFrom`, which maps undefined and null to `Unknown`. If the left decides the result (`False` for `and`, `True` for `or`) it returns that without evaluating the right side at all: no error, no module call, no trace child. Otherwise it evaluates the right side, coerces it the same way, and applies the Kleene operator, so `unknown and false` is `false` and `undefined or true` is `true`. `xor` always evaluates both sides and is derived as `(l or r) and not (l and r)`. These operators never return `Undefined`. Results re-box with `box.Trinary`, so a known result is a `ValueBool`, the same kind comparisons produce.
  - **Side Effects:** The right operand's, only when it is evaluated.
  - **Exceptions:** An operand's evaluation error propagates; a right-side error only when the left did not decide the result.

- **Signature:** Membership and matching - `in`, `contains`, `matches`
  - **Behavior:** `in` and `contains` are the same predicate with operands swapped. `matches` applies a regex from the right operand to the left.
  - **Side Effects:** None.
  - **Exceptions:** Regex compilation or application errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** `and` and `or` short-circuit, so an expensive or failing right operand (a module call, a rule reference) is skipped when the left already determines the result. When the left is `Unknown` or undefined the right still runs, since it can decide the result. Every other operator, `xor` included, evaluates both operands.
- **Dependencies Risk:**
  - **Undefined absorbs every operator except the logic ones.** The blanket undefined check runs only on the eager path, so arithmetic, comparison, `in`, `contains` and `matches` still return `Undefined` for an undefined operand, while `and`, `or` and `xor` read it as `Unknown`. A comparison on missing data therefore yields `Undefined`, and that reaches a logic operator as `Unknown`, not as an absorbing fourth state.
  - **Guard patterns are safe.** `a is defined and a.field > 0` does not evaluate the right side when `a` is undefined, and `false and x` / `true or x` hold for any `x`, including one that errors. The right side runs only when the left is `True` for `and`, `False` for `or`, or `Unknown`.
  - **`+` silently stringifies.** `1 + "a"` produces `"1a"` rather than a type error, because the string branch is checked before the numeric one. Combined with the lack of static type checking on operators, arithmetic typos degrade into concatenation.
  - **`is` is equality only when its right-hand side is a value.** With a type or a shape name on the right it is a conformance check ([[runtime.eval_is_type]]) that returns `false`, not `Undefined`, for an undefined left operand; with a value it keeps the equality semantics here, including undefined propagation.
  - **`%` depends on the representation.** Two exact integers use Go's integer remainder; any float operand uses `math.Mod`. Both truncate, so the sign follows the dividend either way.
  - **The `default` branch is unreachable in practice** but is the only guard against [[parser.lookups]] registering an operator this switch does not implement - a mismatch would surface at request time.
