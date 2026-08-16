---
id: runtime.builtin_call
type: Class
language: Go
file_path: runtime/builtin_call.go
tags: builtins, higher-order, interface-adapter, dependency-inversion
---

# Node: runtime.CallSite (Builtin Environment Adapter)

## 1. Architectural Role & Intent
The adapter that lets [[builtins]] stay independent of the runtime. `CallSite` bundles the execution context, executor, and policy into the frame handed to every builtin, and satisfies `builtins.Env` — the narrow interface through which a higher-order builtin such as `filter` can invoke a caller-supplied callback without importing the evaluator. It is the dependency inversion that keeps the builtin library free of import cycles.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.builtin_call` | `DEPENDS_ON` | [[builtins]] | Implements `builtins.Env`; the compile-time assertion `var _ builtins.Env = (*CallSite)(nil)` pins the contract. |
| `runtime.builtin_call` | `CALLS` | [[runtime.callable]] | `callableFromValue` then `Invoke` / `Arity`. |
| `runtime.builtin_call` | `READS_FROM` | [[runtime.exec_ctx]] | `ExecutionStart()` returns the context's root `CreatedAt`. |
| `runtime.builtin_call` | `DEPENDS_ON` | [[index]] | Carries the `*index.Policy` for downstream type-ref resolution. |
| [[runtime.eval_call]] | `CALLS` | [[runtime.builtin_call]] | Constructs a `CallSite` at each builtin dispatch. |
| [[builtins]] | `CALLS` | [[runtime.builtin_call]] | Higher-order builtins call back through the `Env` methods. |

## 3. Interface Contracts & Public Surface

- **Signature:** `CallSite` — `{ EC *ExecutionContext, Exec *executorImpl, Policy *index.Policy }`
  - **Behavior:** A plain frame struct. Note it holds the **unexported** `*executorImpl` rather than the `Executor` interface, so a `CallSite` can only be constructed from inside the package.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `(*CallSite).Call(ctx, fn box.Value, args []box.Value) -> (box.Value, error)`
  - **Behavior:** Unwraps a boxed callable and invokes it with this site. Both lambdas and derives are accepted transparently.
  - **Side Effects:** Full body evaluation of the callee.
  - **Exceptions:** `expected callable, got %s`; `internal error: callable payload is %T`; any invocation error.

- **Signature:** `(*CallSite).CallableArity(fn box.Value) -> (int, error)`
  - **Behavior:** Reports a callable's **required** arity so a builtin can validate a callback before invoking it per element.
  - **Side Effects:** None.
  - **Exceptions:** Same unwrapping errors as `Call`.

- **Signature:** `(*CallSite).ExecutionStart() -> time.Time`
  - **Behavior:** Returns the root context's creation time, giving time-dependent builtins a value that is **stable for the whole execution** rather than moving between calls.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** A short-lived value constructed per builtin dispatch; all real state lives behind its three pointers.
- **Performance/Scale Notes:** Allocation is trivial, but `Call` is the per-element path for collection builtins, so its cost is the callee's cost multiplied by collection size.
- **Dependencies Risk:**
  - **The `builtins.Env` assertion is load-bearing.** Adding a method to `Env` breaks compilation here, which is the intended safety net — but it also means the runtime must be able to satisfy any capability the builtin library wants, so widening `Env` widens what builtins can reach into.
  - **Time stability is a semantic guarantee, not an optimization.** Because `ExecutionStart` is pinned to context creation, two `now()`-style calls in one policy always agree — including across derive boundaries, since detached contexts inherit `createdAt`. Changing that would make policies nondeterministic within a single evaluation.
  - **`Call` does not re-check derive purity.** A builtin invoked inside a derive body receives a `CallSite` whose `EC` carries `evalDerive`, and it is the *callee's* evaluation path that enforces restrictions. A builtin that captured a callable and invoked it later under a different site would escape that marking.
  - **The site carries the policy, not the namespace**, so any builtin needing cross-namespace resolution must go through `Exec.index` rather than the frame.
