---
id: runtime.eval_call
type: Function / Endpoint
language: Go
file_path: runtime/eval_call.go
tags: dispatch, memoization, derives, builtins, javascript, purity
---

# Node: runtime.evalCall (Call Dispatch and Memoization)

## 1. Architectural Role & Intent
The busiest node in the evaluator. It evaluates arguments, resolves the callee through a four-way precedence ladder — slash-qualified derive, local/short-name derive, native builtin, JavaScript module function — wraps the resolved target in the memoization layer when `!` was used, and invokes it. It is also where derive purity is enforced at dispatch: inside a derive body, impure builtins and all module calls are refused.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_call` | `DEPENDS_ON` | [[builtins]] | `builtins.Table` lookup, then `Decl.Precheck` followed by `Decl.Impl`. |
| `runtime.eval_call` | `DEPENDS_ON` | [[index.package]] | Resolves derives via `Policy.Derives`, `Namespace.Derives`, and `Index.DerivesByFQN`. |
| `runtime.eval_call` | `DEPENDS_ON` | `ext.hashstructure` | Hashes boundary-marshalled arguments for the memoization key. |
| `runtime.eval_call` | `CALLS` | [[runtime.derive_invoke]] | Every derive dispatch routes through `invokeDerive`. |
| `runtime.eval_call` | `CALLS` | [[runtime.modules]] | Module calls marshal arguments and invoke `ModuleBinding.Call`. |
| `runtime.eval_call` | `CALLS` | [[runtime.builtin_call]] | Constructs a `CallSite` per builtin invocation. |
| `runtime.eval_call` | `MUTATES` | `ext.perch` | Reads and populates the executor's call-memoization cache. |
| `runtime.eval_call` | `READS_FROM` | [[runtime.exec_ctx]] | `ec.Module(alias)` and the `evalDerive` purity marker. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_call]] | All `ast.CallExpression` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalCall(ctx, ec, exec, p, t *ast.CallExpression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Evaluates arguments left to right; rejects callable arguments on memoized calls; resolves the target; wraps it for caching; invokes. An `xerr.InjectedError` is passed through verbatim so user-raised errors keep their message, while everything else is wrapped as `failed to call function '%s'`.
  - **Side Effects:** Arbitrary — builtins, derives, and JavaScript all run from here; the memoization cache is populated.
  - **Exceptions:** `memoized call cannot take callable arguments`; target resolution failures; anything the callee raises.

- **Signature:** `getTarget(_, ec, exec, p, c) -> (func(context.Context, ...box.Value) (box.Value, error), error)`
  - **Behavior:** The precedence ladder. A slash-qualified callee of **three or more segments** is tried as a derive FQN first (the segment floor exists to avoid mistaking a two-part division chain for an FQN). Then an identifier callee is tried as a short-name derive, then as a builtin, then as `alias.fn` against bound modules.
  - **Side Effects:** None; returns a closure.
  - **Exceptions:** `builtin %q is not permitted inside a derive`; `TypeScript module calls like %q are not permitted inside a derive`; `xerr.ErrImportResolution`; `xerr.ErrModuleInvocation`; derive visibility and export errors.

- **Signature:** `lookupDeriveByIdentifier(ec, p, name) -> *index.Derive`
  - **Behavior:** Inside a derive, resolves **only** through the caller derive's `DefineShort` snapshot; outside, checks the policy's derives then the namespace's. The two paths are deliberately disjoint.
  - **Side Effects:** None.
  - **Exceptions:** None; returns nil on miss.

- **Signature:** `lookupDeriveBySlashFQ(ec, exec, p, fqn) -> (*index.Derive, error)` / `enforceDeriveExportForCaller` / `enforceDerivePolicyScopeForCaller` / `callerNamespaceFQNForDeriveExport` / `callerPolicyForDeriveScope`
  - **Behavior:** FQN resolution plus the two-stage visibility gate: policy scoping (`VisibleFromPolicy`) always, and cross-namespace export verification (`VerifyDeriveExported`) only when the caller's namespace differs from the derive's. The "caller" is the enclosing derive when inside one, otherwise the policy.
  - **Side Effects:** None.
  - **Exceptions:** `unknown derive %q`; visibility and not-exported errors.

- **Signature:** `calculateHashKey(node *ast.CallExpression, args []box.Value) -> string`
  - **Behavior:** Marshals arguments to boundary values, hashes them, and prefixes the **AST node pointer** so the key is per-call-site. Returns `""` on any failure.
  - **Side Effects:** None.
  - **Exceptions:** None — failures are encoded as the empty string.

- **Signature:** `splitAliasFn(s string) -> (string, string)`
  - **Behavior:** Splits `alias.fn` on the first dot; returns `(s, "")` when there is no dot.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per call, but reads and writes the executor's shared memoization cache.
- **Performance/Scale Notes:** Memoized calls default to a **5 minute TTL** and share a 10 MB budget, so eviction is silent and a hot policy can thrash. Every module call marshals all arguments across the boundary. The `Precheck`/`Impl` split lets builtins short-circuit before doing work.
- **Dependencies Risk:**
  - **An unhashable memoized argument collapses onto the key `""`.** `calculateHashKey` returns the empty string on failure and the caller passes it straight to the cache, so unrelated calls become cache siblings and can receive each other's results. Tracked as a filed issue.
  - **The precedence ladder must stay in sync with [[index.builtin_kind]]'s `isBuiltinCall`.** Static checking assumes local binding > derive > builtin; this function implements derive > builtin with locals handled earlier in [[runtime.eval_ident]]. Drift means checks are applied to a different callee than the one invoked.
  - **The three-segment floor for slash FQNs is a parser-ambiguity workaround.** `a/b` is division; `a/b/c` may be a derive. A genuinely two-segment derive FQN is therefore unreachable by slash syntax.
  - **Module resolution errors are misleading when the callee has no dot.** A bare unknown identifier that is not a derive or builtin falls through to `splitAliasFn`, yields an empty module name, and reports `ErrImportResolution` — an import error for what is really an unknown-function error.
  - **Purity errors at dispatch are the last line, not the first.** [[index.derive_purity]] should have rejected these statically; reaching the runtime message means a static gap was found.
  - **Memoization keys embed a pointer**, which is stable only for the lifetime of the loaded index. The cache lives on the executor, which holds the index, so this is safe today by construction rather than by design.
