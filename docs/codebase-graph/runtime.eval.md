---
id: runtime.eval
type: Function / Endpoint
language: Go
file_path: runtime/eval.go
tags: evaluation, dispatch, tree-walking, tracing
---

# Node: runtime.eval (Expression Dispatch)

## 1. Architectural Role & Intent
The central dispatch of the tree-walking evaluator. Every expression in Sentrie flows through this one type switch, which either evaluates literals inline or delegates to a specialised `eval*` function. Its uniform return shape - value, trace node, error - is what makes the decision explanation tree grow automatically alongside evaluation rather than as a separate pass.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval` | `DEPENDS_ON` | [[ast]] | Type-switches over every concrete expression node. |
| `runtime.eval` | `DEPENDS_ON` | [[box]] | Constructs `box.Null`, `Trinary`, `Number`, `String`, `List`, `Dict` for literals. |
| `runtime.eval` | `CALLS` | [[runtime.trace]] | Opens a node per expression and attaches children, results, and errors. |
| `runtime.eval` | `CALLS` | [[runtime.eval_call]] | Call expressions. |
| `runtime.eval` | `CALLS` | [[runtime.eval_ident]] | Identifier resolution. |
| `runtime.eval` | `CALLS` | [[runtime.eval_infix]] | Binary operators. |
| `runtime.eval` | `CALLS` | [[runtime.eval_unary]] | Prefix operators. |
| `runtime.eval` | `CALLS` | [[runtime.eval_access]] | Field and index access. |
| `runtime.eval` | `CALLS` | [[runtime.eval_block]] | Block expressions with `let`/`yield`. |
| `runtime.eval` | `CALLS` | [[runtime.eval_ternary]] | Conditional and Elvis forms. |
| `runtime.eval` | `CALLS` | [[runtime.eval_cast]] | `as` casts. |
| `runtime.eval` | `CALLS` | [[runtime.eval_lambda]] | Lambda closure creation. |
| `runtime.eval` | `CALLS` | [[runtime.eval_transform]] | `transform … with` expressions. |
| `runtime.eval` | `CALLS` | [[runtime.imports]] | `ImportClause` nodes become cross-policy decision imports. |
| [[runtime.executor]] | `CALLS` | [[runtime.eval]] | Fact defaults, `when`, `default`, body, and attachments all enter here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `eval(ctx, ec *ExecutionContext, exec *executorImpl, p *index.Policy, e ast.Expression) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Unwraps comment wrappers transparently; evaluates the six literal forms inline; delegates everything else. On error it returns `box.Undefined()` with the error recorded on the trace node, so callers get a consistent envelope regardless of failure point.
  - **Side Effects:** Trace tree construction; anything the delegated evaluators do, including JS execution.
  - **Exceptions:** `map key is not a string` for non-string map keys; `pipeline placeholder '#' must be used inside a pipeline call target`; `unsupported expression node: %T`.

- **Signature:** Literal handling - `NullLiteral`, `TrinaryLiteral`, `IntegerLiteral`, `FloatLiteral`, `StringLiteral`, `ListLiteral`, `MapLiteral`
  - **Behavior:** Scalars box directly. Lists evaluate elements in order, aborting on the first error. Maps evaluate **key then value** per entry, requiring the key to box as a string.
  - **Side Effects:** None beyond tracing.
  - **Exceptions:** As above.

- **Signature:** Comment wrappers - `PrecedingCommentExpression`, `TrailingCommentExpression`
  - **Behavior:** Fully transparent: the wrapper is unwrapped and **no trace node is created for it**, so comments never appear in the explanation tree.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless dispatch; all state lives in the execution context.
- **Performance/Scale Notes:** A `trace.Node` is allocated for **every** literal and every expression, so trace cost scales with total AST nodes evaluated, not with interesting decisions. Deep recursion mirrors AST depth and there is no depth limit here - runaway nesting relies on the parser and on `PushRefStack` rather than on this walker.
- **Dependencies Risk:**
  - **Integer and float both box to `box.Number`.** The distinction the parser preserved is erased at the first evaluation step, so no downstream code can recover it.
  - **Duplicate map keys silently collapse.** The parser retains duplicates in `MapLiteral.Entries`, but this loop writes into a Go map, so the last write wins with no diagnostic.
  - **The `#` pipeline hole is a runtime error, not a parse error.** A hole that escapes the parser's rewrite (see [[parser.pipeline]]) surfaces only when the containing branch is actually evaluated, so it can lurk in an untaken path.
  - **The `default` branch is the AST-coverage backstop.** Any expression type the parser can produce but this switch does not list fails at request time with `unsupported expression node`, not at index time.
  - **Trace nodes are built even when nobody consumes the trace**, since there is no conditional gating here - the explanation tree is always paid for.
