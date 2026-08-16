---
id: runtime.trace
type: System / Package
language: Go
file_path: runtime/trace/
tags: observability, explainability, decision-tree, serialization
---

# Node: runtime/trace (Decision Trace Tree)

## 1. Architectural Role & Intent
A deliberately minimal tree structure recording every evaluation step: what kind of node ran, what operator, how long it took, what it produced, and what it failed with. This is the substrate for explainability - the reason an engine can answer "why was this denied?" rather than only "denied" - and it is JSON-serialisable so a trace can be returned over the API or written to a log.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.trace` | `LAYERED_ON` | [[ast]] | Holds the originating `ast.Node` and reads its `Kind()`. |
| `runtime.trace` | `LAYERED_ON` | [[box]] | `Result` carries the boxed value produced by the step. |
| [[runtime.eval]] | `CALLS` | `runtime.trace` | Every evaluator opens a node, attaches children, and records a result or error. |
| [[runtime.executor]] | `READS_FROM` | `runtime.trace` | `ExecutorOutput.RuleNode` is the root of a rule's trace subtree. |
| [[api]] | `READS_FROM` | `runtime.trace` | Serialised into decision responses when explanation is requested. |

## 3. Interface Contracts & Public Surface

- **Signature:** `New(ctx, n ast.Node, op string, meta map[string]any) -> (context.Context, *Node, DoneFn)`
  - **Behavior:** Creates a node stamped with the AST node's `Kind()`, the operator, and metadata, and starts a timer. The returned `DoneFn`, invoked via `defer`, records the elapsed duration. The context is returned **unchanged** - the tree is assembled by explicit `Attach` calls, not by context propagation.
  - **Side Effects:** Allocation; starts a timer.
  - **Exceptions:** None. **Panics if `n` is nil**, since `n.Kind()` is called immediately.

- **Signature:** `(*Node).Attach(children ...*Node) -> *Node` / `SetResult(v box.Value) -> *Node` / `SetErr(err error) -> *Node`
  - **Behavior:** Fluent mutators returning the receiver for chaining. `Attach` with no arguments is a no-op; `SetErr` with a nil error is a no-op and stores `err.Error()` rather than the error itself.
  - **Side Effects:** Mutates the node.
  - **Exceptions:** None.

- **Signature:** `IgnoredStmt(n ast.Node) -> *Node` / `UnsupportedExpression(n ast.Node) -> *Node`
  - **Behavior:** Marker nodes recording that the evaluator encountered something it does not handle, tagged with the Go type name. Used by [[runtime.eval_block]] and [[runtime.eval]] respectively.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Node` struct - JSON contract
  - **Behavior:** `kind`, `op`, `duration`, `meta`, `children`, `result`, `err`. The `Node ast.Node` field is tagged `json:"-"` and excluded, so a serialised trace carries no source positions.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Nodes are mutable and built up in place during evaluation. A tree is per-execution and discarded afterwards unless retained by the caller.
- **Performance/Scale Notes:** **A node is allocated for every evaluation step, unconditionally.** There is no sampling, no depth limit, and no way to disable tracing - a deeply recursive policy over a large collection produces a proportionally large tree, held entirely in memory. `time.Now()` is called twice per node. For hot paths this is the evaluator's main fixed overhead, and it is paid whether or not anyone will read the trace.
- **Dependencies Risk:**
  - **Traces embed evaluated values.** `Result` holds the actual boxed value at each step, so a trace of a policy over sensitive facts contains those facts in full. Anywhere a trace is logged or returned to a caller is a **data-exposure surface**, and nothing in this package redacts.
  - **Nil nodes flow freely.** Several evaluators return a nil `*trace.Node` on error paths - notably [[runtime.eval_cast]], where a recovered panic returns nil. `Attach` tolerates a nil child by appending it, so nil entries can end up **inside** `Children` and reach a consumer that dereferences them.
  - **Errors are flattened to strings** at capture time, so a consumer cannot use `errors.Is` or `errors.As` against a trace; the error chain is lost.
  - **The context returned by `New` is the input context.** The signature suggests context-based propagation, but the tree is assembled manually - so a forgotten `Attach` silently drops an entire subtree with no structural failure.
  - **`Duration` is only set if `DoneFn` runs.** An evaluator that returns without invoking its deferred done function leaves the duration at zero, which is indistinguishable from a genuinely instant step.
  - **Source positions are excluded from JSON**, so a serialised trace cannot be mapped back to policy source without the original AST - limiting how useful an exported trace is to an author debugging remotely.
