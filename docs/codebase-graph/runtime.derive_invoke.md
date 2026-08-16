---
id: runtime.derive_invoke
type: Function / Endpoint
language: Go
file_path: runtime/derive_invoke.go
tags: purity, sandboxing, arity, recursion-detection, type-validation
---

# Node: runtime.invokeDerive (Derive Invocation and Argument Binding)

## 1. Architectural Role & Intent
The runtime half of derive purity. It pads and type-validates arguments, then evaluates the derive body inside a **detached** execution context that has no facts, no lets, no modules, and no parent — so a derive physically cannot read policy state even if a check elsewhere were bypassed. It also enforces the two invariants that keep derives composable: no callable may be returned, and recursion is caught via the cloned reference stack.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.derive_invoke` | `DEPENDS_ON` | [[index.package]] | Reads `index.Derive` — its lambda, FQN, and scope. |
| `runtime.derive_invoke` | `DEPENDS_ON` | [[ast]] | `RequiredLambdaArity`, `ParamTypes`, `ParamOpts`, `ReturnType`. |
| `runtime.derive_invoke` | `CALLS` | [[runtime.exec_ctx]] | `DetachedChildContext`, `PushRefStack`/`PopRefStack`, `SetLocal`. |
| `runtime.derive_invoke` | `CALLS` | [[runtime.eval_block]] | Derive bodies are always block expressions. |
| `runtime.derive_invoke` | `CALLS` | [[runtime.typeref]] | `validateValueAgainstTypeRef` for typed parameters and returns. |
| [[runtime.callable]] | `CALLS` | [[runtime.derive_invoke]] | `deriveCallable.Invoke` delegates here; lambdas share the argument padder. |
| [[runtime.eval_call]] | `CALLS` | [[runtime.derive_invoke]] | Direct derive calls by name or FQN. |

## 3. Interface Contracts & Public Surface

- **Signature:** `padAndValidateCallableArgs(ctx, ec, exec, p, lam, args, argKind string) -> ([]box.Value, error)`
  - **Behavior:** Shared by lambdas and derives. Rejects too many arguments against total parameter count and too few against **required** arity, pads missing optionals with `box.Undefined()`, then validates each typed parameter — **skipping** optional parameters that were left undefined. Error messages name the parameter and are prefixed with `argKind` so lambda and derive failures read differently.
  - **Side Effects:** None.
  - **Exceptions:** `too many arguments: want at most %d, got %d`; `not enough arguments: want at least %d, got %d`; `%s argument %q: …` on type failure.

- **Signature:** `invokeDerive(ctx, callerEC, exec, callerPolicy, d *index.Derive, args) -> (box.Value, error)`
  - **Behavior:** Validates arguments, creates a detached child, marks it with `evalDerive = d`, pushes the derive's FQN onto the cloned reference stack, force-binds parameters as locals, evaluates the body, rejects a callable result, then validates the declared return type.
  - **Side Effects:** Body evaluation under purity restrictions; reference-stack mutation on the child only.
  - **Exceptions:** Argument errors; `xerr.ErrInfiniteRecursion` via `PushRefStack`; `derive cannot yield a callable value`; `derive return: …`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per call; every invocation gets a fresh detached context.
- **Performance/Scale Notes:** Each call clones the reference stack and allocates a context, an argument slice, and a locals map. A derive used as a callback inside a collection builtin pays this per element, which is the dominant cost of derive-heavy policies.
- **Dependencies Risk:**
  - **The detached context is the runtime purity boundary**, and it is enforced structurally rather than by checks: with no facts map, no lets, and no parent, there is simply nothing for a fact lookup to find. [[index.derive_purity]] provides the good error message; this provides the guarantee.
  - **Recursion detection is per-branch, not global.** `PushRefStack` operates on the *child*, whose stack was cloned from the caller. Sibling calls therefore cannot see each other's frames — correct for detecting true cycles, but it means a derive invoked many times in a fan-out pattern is never flagged regardless of depth.
  - **`callerPolicy` is passed through to body evaluation.** A derive's body is evaluated with the **caller's** policy for type-ref and shape resolution, so a derive that references a shape resolves it against whoever called it rather than where it was declared. Cross-policy derive reuse can therefore resolve shapes differently per call site.
  - **The no-callable-return rule is enforced twice** — statically by `yieldHasNoLambda` in [[index.derive_purity]] and dynamically here — because a derive could otherwise return a closure over its detached context and leak it into an impure scope.
  - **Optional parameters are padded with `undefined`, not with a declared default.** Derives have no default-parameter syntax, so an omitted optional is always undefined and the body must handle it.
  - **`argKind` exists purely for message wording**, so a shared failure in the padder reads as "lambda argument" or "derive argument" depending on the caller — useful when tracing which invocation path failed.
