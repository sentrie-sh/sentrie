---
id: builtins
type: System / Package
language: Go
file_path: builtins/
tags: standard-library, purity, higher-order-functions, argument-validation, registry
---

# Node: builtins (Native Function Registry)

## 1. Architectural Role & Intent
The single registry of Sentrie's native functions and the declarative machinery that validates their arguments. Its defining design decision is the narrow `Env` interface: builtins receive only three capabilities from the engine, and widening that interface is explicitly documented as the review event for whether builtins remain derive-safe. The package therefore inverts the usual dependency — builtins do not know about the runtime, the runtime adapts itself to them.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `builtins` | `LAYERED_ON` | [[box]] | Every signature, argument, and result is a `box.Value`; `ValueKind` drives kind checking. |
| `builtins` | `LAYERED_ON` | [[xerr]] | `InjectedError` lets the `error` builtin short-circuit with a user message. |
| [[runtime.eval_call]] | `CALLS` | `builtins` | Looks up `Table`, runs `Decl.Precheck`, then `Decl.Impl`. |
| [[runtime.builtin_call]] | `INHERITS_FROM` | `builtins` | `CallSite` implements the `Env` interface. |
| [[index.builtin_check]] | `READS_FROM` | `builtins` | Static arity and kind checking reads `Table` and each `Decl.Sig`. |
| [[index.derive_purity]] | `READS_FROM` | `builtins` | `IsDeriveSafe` and `DeriveSafeNames` gate what may appear in a derive body. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Env` interface — `Call(ctx, fn box.Value, args []box.Value) (box.Value, error)`, `CallableArity(fn box.Value) (int, error)`, `ExecutionStart() time.Time`
  - **Behavior:** The complete capability set. `Call` lets higher-order builtins invoke lambdas and derives; `ExecutionStart` supplies the pinned clock so `now()` is stable within an execution.
  - **Side Effects:** `Call` runs arbitrary policy code.
  - **Exceptions:** Propagated from the callee.

- **Signature:** `Table map[string]*Decl`
  - **Behavior:** Fifteen builtins. Collection higher-order functions — `all`, `any`, `filter`, `first`, `collect`, `reduce`, `distinct`; list shaping — `as_list`, `normalise_list`, `flatten`, `flatten_deep`, `count`, `merge`; and two specials — `now` and `error`. The map key **must** equal `Decl.Name`, enforced by a test.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Decl` — `{Name, Description, Sig, DeriveSafe, Impl}`
  - **Behavior:** A builtin's complete declaration. `Sig` is rich enough for the static checker to validate call sites without executing anything: per-parameter allowed kinds, optionality, required callable arities, a per-parameter mismatch policy, and custom error strings.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Decl).Precheck(env Env, args []box.Value) -> (handled bool, val box.Value, err error)`
  - **Behavior:** Validates arity then per-argument kinds. The `handled` flag is the interesting part: when a parameter declares `MismatchUndefined`, a kind mismatch **short-circuits the whole call to `Undefined`** rather than erroring, so `count(<not a collection>)` yields undefined instead of failing. Undefined and null arguments always bypass kind checking and reach the implementation.
  - **Side Effects:** May call `env.CallableArity`.
  - **Exceptions:** The signature's `TooFewError`, `TooManyError`, `KindError`, or `CallableArityError`.

- **Signature:** `IsDeriveSafe(name string) -> bool` / `DeriveSafeNames() -> []string`
  - **Behavior:** Purity gate. `DeriveSafe` means deterministic **within** a single execution, not across executions — the doc comment calls this out explicitly, which is why `now` qualifies.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** `Table` is a package-level map, built at initialisation and read-only thereafter. Builtins are stateless.
- **Performance/Scale Notes:** `Precheck` runs on every call and is O(arguments). Higher-order builtins invoke `env.Call` once per element, and each of those is a full lambda or derive evaluation with its own trace subtree — so `filter` over a large list is the dominant cost in most policies that use it.
- **Dependencies Risk:**
  - **A kind mismatch can be silently ignored.** In `precheckArgKinds`, when a defined argument's kind is disallowed and the policy is `MismatchError`, the code reports an error only if `KindError` is set **or** the allowed kinds include `ValueCallable`. A parameter with allowed kinds but no `KindError` and no callable in the list **falls through with no error and no short-circuit**, passing the wrong-kinded value to the implementation. Whether any current declaration hits this depends on every `ParamSig` setting `KindError`; the validation logic itself is unsound.
  - **`TooManyError` falls back to `TooFewError`.** When a signature omits the too-many message, a caller passing extra arguments is told they passed too few.
  - **The arity ceiling is not enforced for variadic builtins**, which is correct, but it means `error`'s only arity guard is the minimum.
  - **Undefined and null always reach the implementation**, bypassing kind checks entirely. Every `Impl` must therefore handle them, and the `Fn` doc comment says so — but nothing enforces it, so a new builtin that assumes a well-kinded argument will panic on undefined input.
  - **`DeriveSafe` is a hand-set boolean.** Nothing verifies that a builtin marked safe actually is; the guarantee rests on review, which is precisely why the `Env` interface is kept narrow enough that widening it is noticeable.
  - **The static checker and the runtime dispatcher must agree.** [[index.builtin_check]] validates against `Sig` at index time and [[runtime.eval_call]] enforces it again at call time. A `Sig` change alters both simultaneously, which is the point — but a resolution-order difference between the two (see [[runtime.eval_call]]) means they can disagree about *which* callee they are checking.
