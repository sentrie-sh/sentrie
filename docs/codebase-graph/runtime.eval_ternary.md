---
id: runtime.eval_ternary
type: Function / Endpoint
language: Go
file_path: runtime/eval_ternary.go
tags: control-flow, short-circuit, elvis, null-coalescing
---

# Node: runtime.evalTernary (Conditional and Elvis)

## 1. Architectural Role & Intent
Implements both `cond ? a : b` and the Elvis form `expr ?: fallback`. It is one of the very few genuinely **short-circuiting** constructs in the evaluator: exactly one branch is evaluated, which makes it the idiomatic way to guard an expensive or failure-prone sub-expression, since [[runtime.eval_infix]] evaluates both operands of `and`/`or` unconditionally.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_ternary` | `DEPENDS_ON` | [[box]] | `IsUndefined`, `IsNull`, `TrinaryFrom(...).IsTrue()`. |
| `runtime.eval_ternary` | `CALLS` | [[runtime.eval]] | Evaluates the condition and exactly one branch. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_ternary]] | `ast.TernaryExpression` nodes dispatch here, both forms. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalTernary(ctx, ec, exec, p, t *ast.TernaryExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Branches on the `Elvis` flag set by the parser.
  - **Side Effects:** Whatever the evaluated branch does.
  - **Exceptions:** Propagated from the condition or the taken branch.

- **Signature:** Elvis form - `expr ?: fallback`
  - **Behavior:** Evaluates the left side; if it is **undefined or null**, evaluates and returns the fallback; otherwise returns the left value. This is null-coalescing, not truthiness-based - a `false`, `0`, or empty string passes through unchanged.
  - **Side Effects:** The fallback is only evaluated when needed.
  - **Exceptions:** Propagated.

- **Signature:** Conditional form - `cond ? then : else`
  - **Behavior:** Coerces the condition with `box.TrinaryFrom` and takes the then-branch **only when the result is `True`**. `False` and `Unknown` both take the else-branch.
  - **Side Effects:** One branch only.
  - **Exceptions:** Propagated.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** The primary short-circuiting primitive in the language. Guarding with `?:` or `? :` is the only reliable way to avoid evaluating an expensive sub-expression, since logical operators do not short-circuit.
- **Dependencies Risk:**
  - **`Unknown` takes the else-branch.** Three-valued logic collapses to two at the conditional, so a condition that is genuinely indeterminate produces the same outcome as one that is definitively false. A policy that needs to distinguish them must test with `is defined` or compare against `unknown` explicitly.
  - **The two forms use different absence tests.** Elvis checks `IsUndefined || IsNull` - a structural absence test - while the conditional checks trinary truthiness. So `x ?: y` and `x ? x : y` are **not** equivalent: for `x = false`, Elvis returns `false` while the conditional returns `y`.
  - **The parser aliases the Elvis condition node.** Per [[parser.ternary]], the abbreviated form stores the same expression node in both the condition and then-branch slots. This evaluator reads only `Condition` for Elvis, so the aliasing is invisible here - but any pass that walks both slots would double-count.
  - **The branches are traced asymmetrically.** The Elvis path sets the result and returns explicitly; the conditional path attaches, sets the result, and returns `err` from the branch - so a branch error still records a result on the node.
  - **Branch precedence differs between the forms** at parse time, which can group a chained ternary differently than expected before evaluation ever begins.
