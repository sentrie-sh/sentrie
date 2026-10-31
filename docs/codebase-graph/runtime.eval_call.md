---
id: runtime.eval_call
type: Function / Endpoint
language: Go
file_path: runtime/eval_call.go
tags: dispatch, memoization, derives, builtins, javascript, purity
---

# Node: runtime.evalCall (Call Dispatch and Memoization)

## 1. Architectural Role & Intent
The busiest node in the evaluator. It evaluates arguments, resolves the callee through a five-way precedence ladder - slash-qualified derive, value binding (a let-bound or parameter-bound callable, #161), local/short-name derive, native builtin, JavaScript module function - wraps the resolved target in the memoization layer when `!` was used, and invokes it. It is also where derive purity is enforced at dispatch: inside a derive body, impure builtins, all module calls, and calls to lambdas defined outside any derive are refused.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_call` | `DEPENDS_ON` | [[builtins]] | `builtins.Table` lookup, the `Decl.PermittedInDerive` guard inside derives, then `Decl.Precheck` followed by `Decl.Impl`. |
| `runtime.eval_call` | `DEPENDS_ON` | [[index]] | Resolves derives via `Policy.Derives`, `Namespace.Derives`, and `Index.DerivesByFQN`. |
| `runtime.eval_call` | `IMPORTS` | `ext.mitchellh.hashstructure` | Hashes boundary-marshalled arguments for the memoization key. |
| `runtime.eval_call` | `CALLS` | [[runtime.derive_invoke]] | Every derive dispatch routes through `invokeDerive`. |
| `runtime.eval_call` | `CALLS` | [[runtime.modules]] | Module calls marshal arguments and invoke `ModuleBinding.Call`. |
| `runtime.eval_call` | `CALLS` | [[runtime.builtin_call]] | Constructs a `CallSite` per builtin invocation and per call through a bound callable. |
| `runtime.eval_call` | `CALLS` | [[runtime.eval_ident]] | `bindingCallTarget` evaluates a binding callee exactly as an identifier reference. |
| `runtime.eval_call` | `CALLS` | [[runtime.callable]] | Invokes a bound callable through `Callable.Invoke`; inspects `lambdaCallable.capture` for the derive gate. |
| `runtime.eval_call` | `MUTATES` | `ext.binaek.perch` | Reads and populates the executor's call-memoization cache. |
| `runtime.eval_call` | `READS_FROM` | [[capabilities]] | The executor's grant gates builtin dispatch. |
| `runtime.eval_call` | `READS_FROM` | [[runtime.exec_ctx]] | `ec.Module(alias)` and the `evalDerive` purity marker. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_call]] | All `ast.CallExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalCall(ctx, ec, exec, p, t *ast.CallExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Evaluates arguments left to right; rejects callable arguments on memoized calls; resolves the target; wraps it for caching; invokes. An `xerr.InjectedError` is passed through verbatim so user-raised errors keep their message, and so is an error wrapping `ErrCallDepthExceeded` or `ErrEvalDepthExceeded`, so a recursion stopped at the depth limit is not prefixed once per level, while everything else is wrapped as `failed to call function '%s'`.
  - **Side Effects:** Arbitrary - builtins, derives, and JavaScript all run from here; the memoization cache is populated.
  - **Exceptions:** `memoized call cannot take callable arguments`; target resolution failures; anything the callee raises.

- **Signature:** `getTarget(_, ec, exec, p, c) -> (func(context.Context, ...box.Value) (box.Value, error), error)`
  - **Behavior:** The precedence ladder, lexical and innermost first. A slash-qualified callee of **three or more segments** is tried as a derive FQN first (the segment floor exists to avoid mistaking a two-part division chain for an FQN). Then an identifier callee that names a value binding (`calleeIsBinding`) is called through it (`bindingCallTarget`), so a let or parameter shadows a derive or builtin of the same name; otherwise it is tried as a short-name derive, then as a builtin, then as `alias.fn` against bound modules. A callee that resolves to none of these is an unknown function (`unresolvedCalleeError`).
  - **Side Effects:** None; returns a closure.
  - **Exceptions:** `builtin %q: not granted by the capabilities file` (wrapping `xerr.ErrNotGranted`) when the executor's [[capabilities]] grant omits the builtin, the runtime half of `index.CheckCapabilities`; `builtin %q is nondeterministic and is not permitted inside a derive` (for `Decl.Nondeterministic`, paired with [[index.derive_purity]]); `builtin %q is not permitted inside a derive`; `TypeScript module calls like %q are not permitted inside a derive`; `unknown function %q`; `xerr.ErrImportResolution` only for a name the policy imports with `use` and calls bare; `xerr.ErrModuleInvocation`; derive visibility and export errors; and the binding errors below.

- **Signature:** `calleeIsBinding(ec, p, name) -> bool`
  - **Behavior:** True when the name is a local or let in the context chain, or outside a derive body an injected or declared fact or a rule of the policy - the lookups `evalIdent` makes before derives. Inside a derive body facts and rules are explicitly not bindings.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `bindingCallTarget(ctx, ec, exec, p, c, id) -> (target, error)`
  - **Behavior:** Refuses a memoized call (the memo key would ignore what the lambda captured), evaluates the binding with `evalIdent`, requires a callable, and inside a derive body allows only a `deriveCallable` or a `lambdaCallable` whose capture has `evalDerive` set (created while evaluating some derive body). The target invokes the callable with a fresh `CallSite`; arity errors come from `padAndValidateCallableArgs`.
  - **Side Effects:** Evaluating the binding may evaluate and cache a let or execute a rule.
  - **Exceptions:** `%q is a local value and cannot be memoized with '!'`; `cannot call %q: it is a %s value, not a function`; `calling %q is not permitted inside a derive - the lambda was defined outside the derive ...`; errors from `evalIdent`.

- **Signature:** `lookupDeriveByIdentifier(ec, p, name) -> *index.Derive`
  - **Behavior:** Inside a derive, resolves **only** through the caller derive's `DefineShort` snapshot; outside, checks the policy's derives then the namespace's. The two paths are deliberately disjoint.
  - **Side Effects:** None.
  - **Exceptions:** None; returns nil on miss.

- **Signature:** `lookupDeriveBySlashFQ(ec, exec, p, fqn) -> (*index.Derive, error)` / `enforceDeriveExportForCaller` / `enforceDerivePolicyScopeForCaller` / `callerNamespaceFQNForDeriveExport` / `callerPolicyForDeriveScope`
  - **Behavior:** FQN resolution plus the two-stage visibility gate: policy scoping (`VisibleFromPolicy`) always, and cross-namespace export verification (`VerifyDeriveExported`) only when the caller's namespace differs from the derive's. The "caller" is the enclosing derive when inside one, otherwise the policy.
  - **Side Effects:** None.
  - **Exceptions:** `unknown derive %q`; visibility and not-exported errors.

- **Signature:** `calculateHashKey(node *ast.CallExpression, args []box.Value) -> (string, error)`
  - **Behavior:** Marshals arguments to boundary values, hashes them, and prefixes the **AST node pointer** so the key is per-call-site. Returns an error when marshalling or hashing fails. For a clock-dependent callee (`isClockDependentCall`: the builtins `now` and `now_ms`, `@sentrie/time.now`, `@sentrie/jwt.decode`/`verify`) `evalCall` appends the execution start to the key, so a memoized answer never crosses executions in the executor-wide cache.
  - **Side Effects:** None.
  - **Exceptions:** Propagates `box.TryToBoundaryAny` and `hashstructure.Hash` failures.

- **Signature:** `splitAliasFn(s string) -> (string, string)`
  - **Behavior:** Splits `alias.fn` on the first dot; returns `(s, "")` when there is no dot.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per call, but reads and writes the executor's shared memoization cache.
- **Performance/Scale Notes:** Memoized calls default to a **5 minute TTL** and share a 10,000-entry LRU (configurable with `WithCallMemoizeCacheSize`), so eviction is silent and a hot policy can thrash. Every module call marshals all arguments across the boundary. The `Precheck`/`Impl` split lets builtins short-circuit before doing work.
- **Dependencies Risk:**
  - **Unhashable memoized arguments bypass the cache.** When `calculateHashKey` fails, the call is invoked directly rather than cached under an empty key.
  - **The precedence ladder must stay in sync with [[index.builtin_kind]]'s `localCallee` and `isBuiltinCall`.** Both sides implement value binding > derive > builtin (#161). Drift means checks are applied to a different callee than the one invoked.
  - **Rules are bindings because `evalIdent` caches a referenced rule's value as a local.** Excluding rules would make call resolution depend on whether the rule had been read yet. So a rule or fact named like a builtin or derive shadows it in a call, and calling it fails as not callable.
  - **A let that calls the builtin it is named after calls itself.** `let first = first(xs, ...)` is a self-reference reported by the let recursion guard.
  - **The derive gate is by provenance, the static gate by syntax.** The runtime accepts any lambda created inside a derive body; [[index.derive_purity]] accepts only a let bound directly to a lambda literal, a sound subset. A callable argument reaching a builtin inside a derive (`filter(xs, fn)` with `fn` a parameter) is not gated here, because builtins invoke it through `CallSite.Call`.
  - **The three-segment floor for slash FQNs is a parser-ambiguity workaround.** `a/b` is division; `a/b/c` may be a derive. A genuinely two-segment derive FQN is therefore unreachable by slash syntax.
  - **Purity errors at dispatch are the last line, not the first.** [[index.derive_purity]] should have rejected these statically; reaching the runtime message means a static gap was found.
  - **Memoization keys embed a pointer**, which is stable only for the lifetime of the loaded index. The cache lives on the executor, which holds the index, so this is safe today by construction rather than by design.
