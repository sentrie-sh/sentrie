---
id: runtime.exec_ctx
type: Class
language: Go
file_path: runtime/exec_ctx.go
tags: scoping, closures, concurrency, recursion-detection, ephemeral-state
---

# Node: runtime.ExecutionContext (Scope Chain)

## 1. Architectural Role & Intent
The per-execution scope chain: facts, policy lets, evaluated locals, module bindings, and the recursion stack for one rule run. It supports two very different child forms - an **attached** child that inherits the parent chain (used for blocks and lambda closures) and a **detached** child that deliberately severs facts, lets, and modules (used for derive bodies, where purity must hold at runtime as well as at index time).

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.exec_ctx` | `DEPENDS_ON` | [[index]] | Holds the `*index.Policy` and consults `policy.Rules` during local assignment. |
| `runtime.exec_ctx` | `DEPENDS_ON` | [[box]] | Locals and fact values are `box.Value`. |
| `runtime.exec_ctx` | `DEPENDS_ON` | [[ast]] | Lets are stored as unevaluated `ast.VarDeclaration` nodes and evaluated lazily. |
| `runtime.exec_ctx` | `DEPENDS_ON` | [[xerr]] | `ErrConflict` for duplicate lets, `ErrInfiniteRecursion` for cycles. |
| [[runtime.executor]] | `CALLS` | [[runtime.exec_ctx]] | Creates the root context and injects facts, lets, and modules. |
| [[runtime.callable]] | `CALLS` | [[runtime.exec_ctx]] | Lambda invocation builds an attached child from the **capture** context. |
| [[runtime.derive_invoke]] | `CALLS` | [[runtime.exec_ctx]] | Derive invocation builds a detached child and sets `evalDerive`. |
| [[runtime.eval]] | `READS_FROM` | [[runtime.exec_ctx]] | Identifier resolution walks facts, lets, locals, and modules through this chain. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ExecutionContext` - `{ rwmu, policy, createdAt, parent, refStack, facts, lets, locals, modules, executor, evalDerive }`
  - **Behavior:** `facts` is populated only on the root; `lets` hold unevaluated declarations; `locals` hold evaluated values. `evalDerive` marks derive-body evaluation and is read by identifier resolution to block facts, rules, module calls, and impure builtins.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `NewExecutionContext(policy, executor) -> *ExecutionContext`
  - **Behavior:** Root context with all maps allocated, no parent, and `createdAt` stamped - the value every `now()`-style builtin reads, so time is stable across a whole execution.
  - **Side Effects:** Allocation.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).AttachedChildContext() -> *ExecutionContext`
  - **Behavior:** Inherits the parent pointer, policy, executor, `createdAt`, `evalDerive`, and - **by reference** - the `modules` map. Clones the refStack. Deliberately sets `facts: nil` so children cannot hold facts, forcing lookups to bubble to the root.
  - **Side Effects:** Allocation.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).DetachedChildContext() -> *ExecutionContext`
  - **Behavior:** The purity boundary: **no parent**, no facts, no lets, no modules, fresh locals, cloned refStack (so derive→derive recursion is still caught across the boundary), and inherited `createdAt` so `now()` stays policy-stable inside derives.
  - **Side Effects:** Allocation.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).InjectFact(ctx, name, v, isDefault, typeRef) -> error`
  - **Behavior:** Root-only fact injection.
  - **Side Effects:** Mutates `facts`.
  - **Exceptions:** `ErrIllegalFactInjection` when called on any child.

- **Signature:** `(*ExecutionContext).InjectLet(name, v) -> error`
  - **Behavior:** Registers a let in the **current** context, never the parent.
  - **Side Effects:** Mutates `lets`.
  - **Exceptions:** `xerr.ErrConflict("let declaration", …)`.

- **Signature:** `(*ExecutionContext).SetLocal(name, value, force bool)` / `GetLocal(name) -> (box.Value, bool)`
  - **Behavior:** `GetLocal` walks up the chain. `SetLocal` with `force` writes locally; without it, it writes locally **only if this context declares that name** as a fact, let, or rule, otherwise it forwards the write to the parent - the mechanism that makes memoized rule and let results land at the right scope level.
  - **Side Effects:** Mutates `locals`, possibly on an ancestor.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).GetFact(name)` / `GetLet(name)` / `IsFactInjected` / `IsLetInjected`
  - **Behavior:** `GetFact` delegates to the parent **first** and only reads its own map at the root, so facts are always root-resolved. `GetLet` checks locally then bubbles.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).BindModule(alias, m)` / `Module(alias) -> (*ModuleBinding, bool)`
  - **Behavior:** Alias-to-binding registry for `use` statements.
  - **Side Effects:** Mutates `modules`.
  - **Exceptions:** None.

- **Signature:** `(*ExecutionContext).PushRefStack(uniqueID) -> error` / `PopRefStack()` / `GetRefStack() -> []string`
  - **Behavior:** Runtime cycle detection by FQN. Push fails if the ID is already on the stack.
  - **Side Effects:** Mutates `refStack`.
  - **Exceptions:** `'%s' references itself: xerr.ErrInfiniteRecursion(...)`.

- **Signature:** `(*ExecutionContext).CreatedAt() -> time.Time` / `Dispose()`
  - **Behavior:** `CreatedAt` recurses to the root. `Dispose` is a **no-op** despite its doc comment describing arena release.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Ephemeral per rule execution, mutated throughout, and disposed (nominally) at the end. Children are cheap allocations.
- **Performance/Scale Notes:** `refStack` is cloned on every child creation, so deeply nested lambda or derive calls copy the stack repeatedly. Lookups walk the parent chain linearly, making deep block nesting progressively slower.
- **Dependencies Risk:**
  - **Attached children share the parent's `modules` map by reference while locking their own `rwmu`.** A `BindModule` on a child and a `Module` read on the parent take *different* mutexes over the *same* map. Binding currently happens before evaluation begins, so this is latent rather than active - but it is a real data race if module binding ever becomes dynamic.
  - **`SetLocal` releases the read lock before taking the write lock.** Between the two, another goroutine can change the fact/let/rule membership the decision was based on. Given `ExecPolicy` runs rules concurrently, contexts must not be shared across those goroutines - and they are not, which is the only thing making this safe.
  - **`evalDerive` is intentionally unguarded**, documented as safe because derive evaluation is single-goroutine per context. Any future parallel derive evaluation breaks that assumption silently.
  - **`Dispose()` does nothing.** The comment promises arena release and warns against reuse; neither is enforced, so leaks would be invisible and reuse-after-dispose would appear to work.
  - **Attached children inherit `evalDerive`**, which is what keeps purity enforcement intact across block evaluation inside a derive body. Creating a child by any other route would silently escape the derive sandbox.
  - **`GetFact` always bubbles to the root before reading.** A child that somehow held facts could never serve them, which is why `AttachedChildContext` sets the map to nil rather than empty.
