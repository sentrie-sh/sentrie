---
id: runtime.executor
type: Class
language: Go
file_path: runtime/executor.go
tags: orchestration, concurrency, fact-binding, module-binding, entrypoint
---

# Node: runtime.Executor (Execution Orchestrator)

## 1. Architectural Role & Intent
The top-level driver: it owns the index, the JS registry, and the two caches, and it sequences a full evaluation - resolve the policy, verify the rule is exported, bind facts (injected or defaulted), bind lets, bind `use` modules, validate fact types, evaluate the rule outcome, then compute export attachments. `ExecPolicy` fans this out across every exported rule concurrently, making the executor both the entrypoint and the concurrency boundary of the system.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.executor` | `DEPENDS_ON` | [[index]] | `ResolvePolicy`, `VerifyRuleExported`, `Policy.Facts/Lets/Uses/Rules/RuleExports`. |
| `runtime.executor` | `CALLS` | [[runtime.exec_ctx]] | Creates the per-run context and injects facts, lets, and module bindings. |
| `runtime.executor` | `CALLS` | [[runtime.eval]] | Evaluates fact defaults, the `when` gate, the body, the `default` clause, and attachments. |
| `runtime.executor` | `CALLS` | [[runtime.decision]] | Wraps the final value via `DecisionOf`. |
| `runtime.executor` | `CALLS` | [[runtime.modules]] | Loads one `jsModule` (pooled `JSInstance` VMs) per module path and builds a per-`use` `ModuleBinding` over it. |
| `runtime.executor` | `CALLS` | [[runtime.js]] | `NewRegistry`, `RegisterGoBuiltin`, `PrepareUse`, `NewAliasRuntime`, `SetupStdLib`, `Require`. |
| `runtime.executor` | `MUTATES` | `ext.binaek.perch` | Populates the module-binding and call-memoization caches. |
| `runtime.executor` | `IMPORTS` | `ext.jackc.puddle` | One VM pool per module binding, `MaxSize: 10`. |
| `runtime.executor` | `CALLS` | [[runtime.trace]] | Opens `rule-outcome`, `rule`, `rule-when`, `rule-default`, `rule-body`, and `attachment` nodes. |
| [[runtime.imports]] | `CALLS` | [[runtime.executor]] | Cross-policy decision imports re-enter through `ExecRule`. |
| [[cmd]] | `CALLS` | [[runtime.executor]] | Builds an executor and runs a policy or rule. |
| [[api]] | `CALLS` | [[runtime.executor]] | Reuses one executor across requests. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewExecutor(idx *index.Index, opts ...NewExecutorOption) -> (Executor, error)`
  - **Behavior:** Constructs the executor, registers fourteen Go builtin modules (`uuid`, `crypto`, `time`, `encoding`, `collection`, `jwt`, `regex`, `net`, `hash`, `url`, `string`, `json`, `semver`, `math`) and the TypeScript `js` shim, creates both caches, and applies options.
  - **Side Effects:** Creates a module-binding cache of 1,024 entries and a memoization cache of 10,000 entries. `perch` allocates entries lazily, so construction costs a few kilobytes regardless of either bound.
  - **Exceptions:** Returns `error` in its signature but **never returns a non-nil error**. It **panics** when `idx.Pack` is nil.

- **Signature:** `WithCallMemoizeCacheSize(entries int) -> NewExecutorOption`
  - **Behavior:** Replaces the memoization cache with one bounded at `entries` entries. The default cache it replaces holds nothing yet, so the swap is cheap.
  - **Exceptions:** Panics (inside `perch.New`) when `entries` is <= 0.
  - **Side Effects:** Allocation.
  - **Exceptions:** None.

- **Signature:** `(*executorImpl).ExecPolicy(ctx, namespace, policy, facts) -> ([]*ExecutorOutput, error)`
  - **Behavior:** Resolves the policy and launches one goroutine per exported rule via `wg.Go`, appending results under a mutex and joining errors.
  - **Side Effects:** Full concurrent evaluation; JS execution; cache population.
  - **Exceptions:** Policy resolution failure; joined per-rule errors; a synthesized `panic in ExecRule "<rule>": …` error joined with the others.

- **Signature:** `(*executorImpl).ExecRule(ctx, namespace, policy, rule, injectedFacts) -> (*ExecutorOutput, error)`
  - **Behavior:** Verifies the rule is exported, creates the context, then for each declared fact: uses the injected value if present, errors if a required fact is missing, or evaluates the declared default. Binds lets, binds `use` modules, and delegates to `execRule`. On error with no decision, substitutes an `Unknown` decision so the envelope is always populated.
  - **Side Effects:** Mutates the execution context; executes JavaScript.
  - **Exceptions:** `xerr.ErrRequiredFact`; `fact '%s' cannot be null`; `fact '%s' cannot have null default value`; `xerr.ErrUnresolvableFact`; anything from evaluation.

- **Signature:** `(*executorImpl).execRule(ctx, ec, namespace, policy, rule) -> (*Decision, DecisionAttachments, *trace.Node, error)`
  - **Behavior:** Looks up the rule, pushes it onto the recursion stack, opens the trace node, validates every injected fact against its declared type ref, evaluates the outcome, then evaluates each export attachment in declaration order.
  - **Side Effects:** Trace tree construction; recursion-stack mutation.
  - **Exceptions:** `xerr.ErrRuleNotFound`; recursion errors; type-validation failures; attachment evaluation errors.

- **Signature:** `evaluateRuleOutcome(ctx, ec, e, p, r) -> (*Decision, *trace.Node, error)`
  - **Behavior:** The rule semantics in one place: `when` defaults to `True` when absent; if the gate is **not true** the rule yields the `default` expression (or `Unknown` when there is none); otherwise it yields the body. Note the gate uses `IsTrue()`, so both `False` **and** `Unknown` take the default branch.
  - **Side Effects:** Trace nodes.
  - **Exceptions:** Propagated from the gate, default, or body.

- **Signature:** `(*executorImpl).bindUses(ctx, ec, p) -> error` / `getModuleBinding(ctx, use, ms) -> (*ModuleBinding, error)` / `getModule(ctx, ms) -> (*jsModule, error)` / `loadJSInstance(ctx, ms) -> (*JSInstance, error)`
  - **Behavior:** Resolves each `use` to a module spec relative to the policy file's directory and binds it under the policy's alias. `getModule` loads the module into a warmed `puddle` pool keyed by **module path alone**; `loadJSInstance` is the pool constructor and builds a VM carrying the module's whole exports object. `getModuleBinding` then checks this `use`'s named exports against the module and returns a per-`use` `ModuleBinding` holding the alias and the requested names. The export projection is applied per binding at call time, so two `use`s of one module never see each other's restriction ([#102](https://github.com/sentrie-sh/sentrie/issues/102)).
  - **Side Effects:** Compiles and executes JavaScript at bind time; populates the module cache.
  - **Exceptions:** Module resolution failures; `module %s missing required export %q` (per `use`, at bind time); pool warm-up failures.

## 4. Operational Context & Gotchas
- **Statefulness:** Long-lived and shared. Caches and VM pools outlive individual executions and are reused across concurrent calls.
- **Performance/Scale Notes:** Every exported rule of a policy runs in parallel against shared caches. VM pools cap at 10 instances per module, so contention appears as latency, not errors. Module binding happens per `ExecRule`, and the binding is **not** retained: see the gotcha below.
- **Dependencies Risk:**
  - **Panic recovery in `ExecPolicy` holds the mutex and joins errors**, including the panicking rule's name in the message, so concurrent rule failures are not lost or raced.
  - **`ExecPolicy` output order is nondeterministic.** It ranges over the `RuleExports` map and appends as goroutines finish, so callers must not rely on ordering.
  - **Module bindings are never actually cached** ([sentrie-sh/sentrie#191](https://github.com/sentrie-sh/sentrie/issues/191)). `getModuleBinding` passes `ttl = -1`, which `perch` treats as "do not cache": concurrent loads of the same key are deduplicated, then the entry is dropped. Every `ExecRule` therefore builds a new `puddle` pool and warms a VM per distinct module path it uses, and the old pool is never closed.
  - **`defer done()` inside the attachment loop** accumulates trace closures until `execRule` returns rather than closing each span promptly; the same pattern appears in `evaluateRuleOutcome`.
  - **The `when` gate treats `Unknown` as "not true".** A guard that evaluates to `Unknown` silently takes the default branch rather than propagating uncertainty into the body.
  - **`NewExecutor`'s error return is vestigial**, so a caller's error handling gives false assurance while the real failure mode is a nil-pack panic.
