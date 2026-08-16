---
id: runtime.eval_lambda
type: Function / Endpoint
language: Go
file_path: runtime/eval_lambda.go
tags: closures, first-class-functions, capture-semantics
---

# Node: runtime.evalLambda (Closure Creation)

## 1. Architectural Role & Intent
Turns a lambda literal into a first-class runtime value. It performs no evaluation of the body — it only boxes the AST node together with a **reference** to the current execution context, deferring everything else to invocation. The reference-capture choice is what gives Sentrie lambdas late-bound lexical scoping.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_lambda` | `CALLS` | [[runtime.callable]] | `newLambdaCallable(lam, ec)` builds the closure. |
| `runtime.eval_lambda` | `DEPENDS_ON` | [[box]] | `box.Callable` boxes the closure as a value. |
| `runtime.eval_lambda` | `READS_FROM` | [[runtime.exec_ctx]] | Captures the current context by reference, without copying. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_lambda]] | `ast.LambdaExpression` nodes dispatch here. |
| [[builtins]] | `CALLS` | [[runtime.eval_lambda]] | Higher-order builtins receive the boxed result as a callback argument. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalLambda(ctx, ec, _ *executorImpl, _ *index.Policy, lam *ast.LambdaExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Constructs a `lambdaCallable` over the AST node and the current context, boxes it, and returns. The executor and policy parameters are **unused** — deliberately, since nothing about the definition site depends on them; they are resolved at invocation from the `CallSite`.
  - **Side Effects:** None. No body evaluation, no scope creation.
  - **Exceptions:** None — it cannot fail.

## 4. Operational Context & Gotchas
- **Statefulness:** Produces a value holding a live reference to mutable state. The closure is not a snapshot.
- **Performance/Scale Notes:** Trivial: one small allocation. All cost is at invocation, in [[runtime.callable]].
- **Dependencies Risk:**
  - **Capture by reference means late binding.** The comment marks this as a deliberate v1 choice. A lambda invoked after its defining scope has been mutated sees the **current** values, not those in effect at creation. Since [[runtime.eval_ident]] writes evaluated lets back into the context as locals, a lambda created before a let is first read and invoked after will observe the cached value — so behaviour can depend on evaluation order elsewhere in the policy.
  - **The captured context can outlive the expression that created it.** A lambda stored in a list or returned from a block keeps its whole parent chain alive, which is a retention consideration as much as a semantic one.
  - **Nothing validates the lambda at creation.** Parameter types, return type, and body are all checked only when invoked, so a lambda that can never be called successfully is created without complaint and fails at the call site — possibly in a different policy.
  - **Derives cannot yield callables** (enforced in [[runtime.derive_invoke]] and statically in [[index.derive_purity]]), which is what stops a lambda created inside a detached derive context from escaping into an impure scope.
  - **The trace node records only parameter names**, not the captured scope, so a decision trace cannot explain what a closure resolved a free variable to.
