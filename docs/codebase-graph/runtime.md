---
id: runtime
type: System / Package
language: Go
file_path: runtime/
tags: evaluation, execution, sandboxing, javascript, concurrency, back-end
---

# Node: Runtime (Policy Evaluation Engine)

## 1. Architectural Role & Intent
`runtime` is Sentrie's back-end: it takes a validated [[index]] plus a caller-supplied fact map and produces decisions. It owns the tree-walking evaluator, the per-execution scope chain, the embedded JavaScript/TypeScript sandbox used by `use` modules, memoization, cross-policy decision imports, and the trace tree that explains how a decision was reached. This is the only package that executes untrusted-adjacent code and the only one that runs concurrently by design, which makes it the highest-risk area of the system.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime` | `LAYERED_ON` | [[index]] | Resolves policies, rules, facts, lets, derives, and export metadata. Requires a validated index. |
| `runtime` | `LAYERED_ON` | [[ast]] | Walks expression nodes directly; there is no separate IR or bytecode. |
| `runtime` | `LAYERED_ON` | [[box]] | `box.Value` is the universal runtime value representation and boundary codec. |
| `runtime` | `LAYERED_ON` | [[trinary]] | Decisions are Kleene three-valued; `Unknown` is the failure-safe default. |
| `runtime` | `LAYERED_ON` | [[builtins]] | Dispatches native builtin calls through the `builtins.Env` interface. |
| `runtime` | `LAYERED_ON` | [[pack]] | Reads `Permissions` to gate module and environment access, and `Location` to resolve module paths. |
| `runtime` | `LAYERED_ON` | [[xerr]] | Structured error sentinels for required facts, recursion, and type failures. |
| `runtime` | `LAYERED_ON` | [[runtime.js]] | Owns the goja VM registry, TypeScript compilation, and the `@sentra/*` standard library. |
| `runtime` | `LAYERED_ON` | [[runtime.trace]] | Builds the decision explanation tree alongside evaluation. |
| `runtime` | `IMPORTS` | `ext.dop251.goja` | Embedded ECMAScript interpreter for `use` modules. |
| `runtime` | `IMPORTS` | `ext.jackc.puddle` | Pools JS VM instances per module binding. |
| `runtime` | `IMPORTS` | `ext.binaek.perch` | Size-bounded caches for module bindings and memoized calls. |
| [[cmd]] | `CALLS` | [[runtime]] | The `exec` command builds an executor and runs a policy or rule. |
| [[api]] | `CALLS` | [[runtime]] | Serves evaluation requests against a long-lived executor. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewExecutor(idx *index.Index, opts ...NewExecutorOption) -> (Executor, error)`
  - **Behavior:** Builds the executor, registers all fourteen Go-backed `@sentra/*` builtin modules plus the TypeScript `js` shim, and reserves both caches. See [[runtime.executor]].
  - **Side Effects:** Allocates two `perch` caches (100 MB module bindings, 10 MB memoization by default).
  - **Exceptions:** **Panics** if `idx.Pack` is nil - it dereferences `idx.Pack.Location` before any nil check.

- **Signature:** `Executor` interface - `ExecPolicy(ctx, namespace, policy, facts) -> ([]*ExecutorOutput, error)`, `ExecRule(ctx, namespace, policy, rule, facts) -> (*ExecutorOutput, error)`, `Index() -> *index.Index`
  - **Behavior:** The public entrypoints. `ExecPolicy` fans out over every exported rule **concurrently**; `ExecRule` runs one.
  - **Side Effects:** Executes JavaScript, populates caches, builds trace trees.
  - **Exceptions:** Missing required facts, unexported rules, evaluation errors, recursion errors.

- **Signature:** `ExecutorOutput` - `{ PolicyName, Namespace, RuleName, Decision *Decision, Attachments DecisionAttachments, RuleNode *trace.Node }`
  - **Behavior:** The result envelope, JSON-serializable for CLI and API consumers.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `Decision` - `{ State trinary.Value, Value box.Value }` / `DecisionOf(val) -> *Decision`
  - **Behavior:** The decision envelope and its coercion rule. See [[runtime.decision]].
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `ExecutionContext` - the per-run scope chain
  - **Behavior:** Holds facts, lets, locals, module bindings, and the recursion stack. See [[runtime.exec_ctx]].
  - **Side Effects:** Mutated throughout evaluation.
  - **Exceptions:** `ErrIllegalFactInjection`.

- **Signature:** `Callable` interface - `Arity() -> int`, `Invoke(ctx, site, args) -> (box.Value, error)`
  - **Behavior:** Unifies lambdas and derives so builtins can accept either. See [[runtime.callable]].
  - **Side Effects:** Evaluation.
  - **Exceptions:** Arity and type-validation failures.

- **Signature:** `CallSite` - implements `builtins.Env`
  - **Behavior:** The frame handed to every builtin, enabling higher-order builtins to call back into evaluation. See [[runtime.builtin_call]].
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** Type-ref validation surface - `validateValueAgainstTypeRef` and the per-kind `typeref_*.go` checkers
  - **Behavior:** Enforces declared types and constraints on facts, parameters, and returns at evaluation time.
  - **Side Effects:** None.
  - **Exceptions:** `ErrTypeRef`, `ErrConstraintFailed`, `ErrUnknownConstraint` - see [[runtime.err_typedef]].

## 4. Operational Context & Gotchas
- **Statefulness:** The executor is a **long-lived singleton** holding caches and JS VM pools; each `ExecRule` creates an **ephemeral** `ExecutionContext` that is discarded afterwards. `Dispose()` exists but is currently a no-op.
- **Performance/Scale Notes:**
  - `ExecPolicy` runs every exported rule in parallel goroutines against a **shared executor**, so all cache and VM-pool concurrency issues are exercised on every policy run.
  - JS VM instances are pooled at **MaxSize 10** per module binding; a policy with heavy module use across many concurrent evaluations will block on acquisition.
  - Module bindings are cached globally by canonical module path, and memoized calls share a 10 MB budget by default - evictions are silent.
- **Dependencies Risk:**
  - **This package converts static guarantees into runtime behaviour.** Everything [[index.validate]] could not prove - type conformance, arity of dynamic callables, module export presence - fails here, at request time, in front of a user.
  - **Concurrency correctness is not fully established.** `ExecPolicy`'s panic recovery writes shared state without holding its mutex, and attached child contexts share the parent's `modules` map while locking their own mutex. See [[runtime.executor]] and [[runtime.exec_ctx]].
  - **Several nil-dereference paths are reachable from policy input**, notably the fact alias/name mismatch in [[runtime.executor]] and the unchecked `output.RuleNode` in [[runtime.imports]].
  - **The JS sandbox is the trust boundary.** Module permissions come from the pack manifest, but once a VM is bound, cache keying decides which export set a policy actually sees - see the canonical-key caveat in [[runtime.executor]].
  - **Failure is biased toward `Unknown`, not toward denial.** A rule that errors yields an `Unknown` decision envelope alongside the error; callers that ignore the error read a permissive-looking result.
