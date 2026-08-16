---
id: index.builtin_check
type: Function / Endpoint
language: Go
file_path: index/builtin_check.go
tags: static-analysis, arity-checking, type-checking, diagnostics
---

# Node: index.checkBuiltinCalls (Static Builtin Call Validation)

## 1. Architectural Role & Intent
Walks every rule expression and every derive body looking for calls to native builtins, and validates each against its declared signature: argument count against the required/optional/variadic shape, argument value kinds against the allowed set, and callback arity for higher-order builtins. It moves a whole class of errors - `length()` with no argument, `filter(xs, 3)` - from evaluation time to index time, with a source span attached.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.builtin_check` | `DEPENDS_ON` | [[builtins]] | Reads `Decl.Sig` - params, variadic tail, optionality, allowed kinds, callback arities, and the pre-authored error strings. |
| `index.builtin_check` | `DEPENDS_ON` | [[box]] | Compares against `box.ValueKind`, including the `ValueCallable` special case. |
| `index.builtin_check` | `DEPENDS_ON` | [[xerr]] | Emits `ErrBuiltinCallArity`, `ErrBuiltinArgKind`, `ErrBuiltinCallableArity`, each carrying a span. |
| `index.builtin_check` | `CALLS` | [[index.builtin_kind]] | Uses `kindCheckCtx` for scope construction, `isBuiltinCall`, `resolveKind`, and `resolveCallableArity`. |
| `index.builtin_check` | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` is the default traversal branch. |
| `index.builtin_check` | `READS_FROM` | [[index.rule]] | Walks `Default`, `When`, and `Body` of every rule. |
| `index.builtin_check` | `READS_FROM` | [[index.derive]] | Walks every derive's lambda body with a derive-flavoured scope. |
| [[index.validate]] | `CALLS` | [[index.builtin_check]] | Runs as the final check in the validation pipeline. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkBuiltinCalls(ctx) -> error`
  - **Behavior:** Two loops - one over policies building a rule-flavoured `kindCheckCtx` per policy, one over `DerivesByFQN` building a derive-flavoured one. Collects **all** errors and returns them joined, so a single run reports every bad call rather than stopping at the first.
  - **Side Effects:** None.
  - **Exceptions:** `errors.Join` of every diagnostic; `validation cancelled` on context cancellation.

- **Signature:** `(*kindCheckCtx).walkExpr(e ast.Expression) -> []error`
  - **Behavior:** Scope-aware traversal. A call is checked then recursed into; a **block** clones the scope and binds each `let` *after* checking its initializer, so bindings shadow correctly in sequence; a **lambda** pushes a child scope with its parameters. Everything else delegates to the generic child walker.
  - **Side Effects:** None; scopes are cloned, never mutated in place across branches.
  - **Exceptions:** Returns errors rather than raising.

- **Signature:** `(*kindCheckCtx).checkBuiltinCall(c *ast.CallExpression) -> []error`
  - **Behavior:** Skips non-builtin callees. Computes `min` from the count of non-optional params and `max` from total params; reports too-many only when the signature is **non-variadic** and a `TooManyError` message exists, pointing the span at the first excess argument. Reports too-few pointing at the callee identifier. Then per argument: checks the value kind when statically known and the mismatch policy is not `MismatchUndefined`, synthesising an "expected callable" message when none was authored; and checks callback arity when the signature declares allowed arities.
  - **Side Effects:** None.
  - **Exceptions:** Returns the three `xerr` builtin diagnostics.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per invocation; scopes are values copied down the walk.
- **Performance/Scale Notes:** One walk per rule slot and per derive body, on top of the walks already performed by purity and cycle checking - the same ASTs are traversed several times per validation. `cloneBindingScope` copies on every block and lambda.
- **Dependencies Risk:**
  - **Checks are opt-in per signature.** A kind mismatch is only reported when `Kinds` is populated **and** `OnMismatch != MismatchUndefined` **and** a message exists (or the callable special case applies). A builtin declared without these fields is effectively unchecked, so absence of a diagnostic does not mean the call is correct.
  - **Unknown kinds pass silently.** `resolveKind` returns "not known" for most expressions, and unknown always means accept. This is deliberately biased toward **no false positives**, which means real errors routinely slip through to runtime.
  - **Shadowing precedence must match the runtime.** `isBuiltinCall` skips names bound locally or by a derive, mirroring the runtime's `getTarget` order. If the runtime's precedence changes, this check would validate against the wrong callee.
  - **Non-`VarDeclaration` statements in a block are skipped, not reported**, unlike the purity walk which rejects them - the two walkers disagree about how strict a block is.
  - **Diagnostics are joined, so output can be long.** A pack with a systematic mistake reports every occurrence at once.
