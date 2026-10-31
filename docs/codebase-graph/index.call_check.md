---
id: index.call_check
type: Function / Endpoint
language: Go
file_path: index/call_check.go
tags: static-analysis, call-resolution, diagnostics, lambdas
---

# Node: index.checkCallTargets (Static Callee Validation)

## 1. Architectural Role & Intent
Checks the callee of every call whose callee is a bare identifier, resolving it in the order the runtime's `getTarget` uses ([[runtime.eval_call]]): a value binding, then a visible derive, then a builtin. It is what lets `sentrie validate` accept a call to a let-bound lambda while rejecting a callee that names nothing callable ([#161](https://github.com/sentrie-sh/sentrie/issues/161)). It walks the same expression sites as [[index.builtin_check]] through `forEachScopedExpr`, so every policy slot and derive body is covered.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.call_check` | `CALLS` | [[index.builtin_kind]] | `localCallee`, `lookupDerive`, `isShadowed` and `resolveIdentKind` on the walk's `kindCheckCtx`. |
| `index.call_check` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` supplies each expression with its lexical scope. |
| `index.call_check` | `DEPENDS_ON` | [[builtins]] | `builtins.Table` decides whether an unbound name is a builtin. |
| `index.call_check` | `DEPENDS_ON` | [[xerr]] | Emits span-anchored `CallTargetError`, which unwraps to `ErrIndex`. |
| `index.call_check` | `READS_FROM` | [[index.policy]] | `Policy.Uses` for the imported-name hint; `Policy.Rules` through `localCallee`. |
| [[index.validate]] | `CALLS` | [[index.call_check]] | Runs after derive purity, before builtin call checks. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkCallTargets(ctx) -> error`
  - **Behavior:** Visits every `CallExpression` with an `*ast.Identifier` callee. A value binding must not be a rule, must not be memoized with `!`, must not have a statically known kind other than callable, and when it is a let bound directly to a lambda literal must receive between the lambda's required and total parameter count. A visible derive or builtin is accepted (builtin calls are checked by [[index.builtin_check]]). Anything else is an unknown function; when the policy imports the name with `use`, the message names the module alias to call it through.
  - **Side Effects:** None.
  - **Exceptions:** `errors.Join` of every `CallTargetError`; `validation cancelled` from the walk.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; scopes are cloned by the walk.
- **Performance/Scale Notes:** One more pass over every expression site, on top of the builtin, nondeterminism and capabilities walks.
- **Dependencies Risk:**
  - **Kind checks are gradual.** A binding whose kind cannot be inferred (an unannotated parameter, a let initialised from a call) is accepted, and the runtime reports a non-callable value when the call runs.
  - **Arity is checked only for a let bound directly to a lambda literal.** A parameter or a let aliasing another callable is checked at runtime by `padAndValidateCallableArgs`; derive call arity is not checked statically at all.
  - **Rule calls are rejected outright**, although the runtime would call a rule's value if it were callable. Static is deliberately stricter here.
  - **Precedence is shared with the runtime.** `localCallee` and `isBuiltinCall` must stay in step with `calleeIsBinding` in [[runtime.eval_call]].
