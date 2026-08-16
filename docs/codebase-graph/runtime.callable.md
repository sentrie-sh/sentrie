---
id: runtime.callable
type: Class
language: Go
file_path: runtime/callable.go
tags: closures, first-class-functions, higher-order, late-binding
---

# Node: runtime.Callable (Closure Abstraction)

## 1. Architectural Role & Intent
Unifies the two things Sentrie can invoke — an anonymous lambda and a named derive — behind one `Callable` interface, so higher-order builtins like `filter` and `map` accept either without knowing the difference. Lambda closures capture a **reference** to their defining execution context rather than a snapshot, making lexical lookups late-bound against the live parent chain.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.callable` | `DEPENDS_ON` | [[ast]] | Wraps `ast.LambdaExpression` and reads `RequiredLambdaArity`. |
| `runtime.callable` | `DEPENDS_ON` | [[index]] | `deriveCallable` wraps an `index.Derive`. |
| `runtime.callable` | `DEPENDS_ON` | [[box]] | Callables are carried as `box.Value` via `CallableRef`. |
| `runtime.callable` | `CALLS` | [[runtime.exec_ctx]] | Lambda invocation builds an **attached** child from the capture context. |
| `runtime.callable` | `CALLS` | [[runtime.derive_invoke]] | Argument padding/validation and `invokeDerive`. |
| `runtime.callable` | `CALLS` | [[runtime.eval_block]] | Lambda bodies are always block expressions. |
| `runtime.callable` | `CALLS` | [[runtime.typeref]] | Validates typed parameters and declared return types. |
| [[runtime.builtin_call]] | `CALLS` | [[runtime.callable]] | `CallSite.Call` and `CallableArity` unwrap and invoke. |
| [[runtime.eval_lambda]] | `CALLS` | [[runtime.callable]] | Boxes a lambda literal into a `lambdaCallable`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Callable` interface — `Arity() -> int`, `Invoke(ctx, site *CallSite, args []box.Value) -> (box.Value, error)`
  - **Behavior:** `Arity` reports the **required** parameter count, excluding optionals, which is what builtin callback-arity checks compare against.
  - **Side Effects:** Invocation evaluates a body.
  - **Exceptions:** Implementation-specific.

- **Signature:** `newLambdaCallable(lambda, capture *ExecutionContext) -> *lambdaCallable`
  - **Behavior:** Stores the AST node and a **live reference** to the defining context. Nothing is copied, so later mutations to the capture chain are visible to the closure.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*lambdaCallable).Invoke(ctx, site, args) -> (box.Value, error)`
  - **Behavior:** Pads and validates arguments, builds an attached child of the **capture** context (not the call site's context), force-binds parameters as locals, evaluates the body, then validates the declared return type if present.
  - **Side Effects:** Body evaluation, including any module calls the body performs.
  - **Exceptions:** Arity errors; parameter type-validation errors; `lambda return: …` on return-type mismatch.

- **Signature:** `(*deriveCallable).Invoke(ctx, site, args) -> (box.Value, error)` / `Arity() -> int`
  - **Behavior:** Delegates to `invokeDerive` with the call site's context, which builds a **detached** child. `Arity` is nil-safe and returns 0 for a nil derive.
  - **Side Effects:** Derive body evaluation under purity restrictions.
  - **Exceptions:** `internal error: missing derive`; anything `invokeDerive` raises.

- **Signature:** `callableFromValue(v box.Value) -> (Callable, error)`
  - **Behavior:** Unwraps a boxed callable reference and asserts it satisfies the interface.
  - **Side Effects:** None.
  - **Exceptions:** `expected callable, got %s`; `internal error: callable payload is %T`.

## 4. Operational Context & Gotchas
- **Statefulness:** `lambdaCallable` is a live closure over mutable state; `deriveCallable` is a stateless wrapper.
- **Performance/Scale Notes:** Each invocation allocates a child context and an argument slice. Higher-order builtins invoking a callable per element make this the hottest allocation path in the evaluator.
- **Dependencies Risk:**
  - **Late binding is a deliberate semantic choice with sharp edges.** Because the capture is a live reference, a lambda invoked after its defining scope has been mutated observes the **new** values, not those in effect at definition. Callbacks stored and invoked later can therefore see a different environment than the one they were written against.
  - **The two `Invoke` implementations use different contexts.** Lambdas build from `c.capture` (lexical scoping); derives build from `site.EC` (the caller). That asymmetry is correct — derives are context-free by design — but means a derive and a lambda passed to the same builtin resolve free identifiers differently.
  - **Lambdas escape derive purity only through the context they carry.** A lambda defined inside a derive body captures the detached context, so purity holds; a lambda defined outside and passed in would carry the full attached chain. Purity checking in [[index.derive_purity]] is what prevents that construction, not anything here.
  - **`lambdaCallable.Invoke` does not nil-check `c.lambda.Body`**, relying on the parser always producing a block. A lambda whose body failed to parse would panic rather than error.
  - **The derive `Arity()` nil-guard has no counterpart on the lambda side**, so the two implementations differ in robustness for no stated reason.
