---
id: index.builtin_check
type: Function / Endpoint
language: Go
file_path: index/builtin_check.go
tags: static-analysis, arity-checking, type-checking, diagnostics
---

# Node: index.checkBuiltinCalls (Static Builtin Call Validation)

## 1. Architectural Role & Intent
Walks every expression a policy holds (fact defaults, policy lets, rule `default`/`when`/body slots, rule contribution guards and bodies, `import decision ... with` expressions, export attachments) and every derive body, looking for calls to native builtins, and validates each against its declared signature: argument count against the required/optional/variadic shape, argument value kinds against the allowed set, and callback arity for higher-order builtins. It moves a whole class of errors - `length()` with no argument, `filter(xs, 3)` - from evaluation time to index time, with a source span attached. The walk is `forEachBuiltinCall` (`index/builtin_calls.go`), shared with the nondeterminism warnings and the capabilities check in [[index]], so a new expression site is covered by all three at once (#218). The runtime makes the same checks from the same signature when the call is evaluated.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.builtin_check` | `DEPENDS_ON` | [[builtins]] | Reads `Decl.Sig` - params, variadic tail, optionality, allowed kinds, callback arities, and the pre-authored error strings. |
| `index.builtin_check` | `DEPENDS_ON` | [[box]] | Compares against `box.ValueKind`, including the `ValueCallable` special case. |
| `index.builtin_check` | `DEPENDS_ON` | [[xerr]] | Emits `ErrBuiltinCallArity`, `ErrBuiltinArgKind`, `ErrBuiltinCallableArity`, each carrying a span. |
| `index.builtin_check` | `CALLS` | [[index.builtin_kind]] | Uses `kindCheckCtx` for scope construction, `isBuiltinCall`, `resolveKind`, and `resolveCallableArity`. |
| `index.builtin_check` | `CALLS` | [[index.derive_expr_walk]] | `forEachDeriveExprChild` is the default traversal branch. |
| `index.builtin_check` | `READS_FROM` | [[index.rule]] | Walks `Default`, `When`, and `Body` of every rule, including the `with` expressions when the body is an `ast.ImportClause`, and the `When` and `Body` of each of its contributions. |
| `index.builtin_check` | `READS_FROM` | [[index.policy]] | Walks every fact default, policy `let` value and export attachment value. |
| `index.builtin_check` | `READS_FROM` | [[index.derive]] | Walks every derive's lambda body with a derive-flavoured scope. |
| [[index.validate]] | `CALLS` | [[index.builtin_check]] | Runs as the final check in the validation pipeline. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkBuiltinCalls(ctx) -> error`
  - **Behavior:** A visitor over `forEachBuiltinCall`: each call that resolves to a builtin is checked by `checkBuiltinCall` in the scope it resolved in. Collects **all** errors and returns them joined, so a single run reports every bad call rather than stopping at the first.
  - **Side Effects:** None.
  - **Exceptions:** `errors.Join` of every diagnostic; `validation cancelled` on context cancellation (checked per namespace, per policy and per derive).

- **Signature:** `(*Index).forEachBuiltinCall(ctx, visit func(k *kindCheckCtx, decl *builtins.Decl, call *ast.CallExpression)) -> error`
  - **Behavior:** A filter over `forEachScopedExpr` that passes on each call the runtime sends to a builtin. Used by `checkBuiltinCalls`, `NondeterminismWarnings` and `CheckCapabilities`.
  - **Side Effects:** None.
  - **Exceptions:** `validation cancelled`.

- **Signature:** `(*Index).forEachScopedExpr(ctx, visit func(k *kindCheckCtx, e ast.Expression)) -> error`
  - **Behavior:** The one walk over every expression site: per policy, a rule-flavoured `kindCheckCtx` (facts, lets and derives in scope) over fact defaults, lets, rule slots and export attachments; per derive, a derive-flavoured one over the lambda body. `walkScopedExprs` does the scope-aware pre-order traversal: every node is visited before its children, so a visitor that rewrites a node ([[index.is_resolve]]) is walked as rewritten; a **block** clones the scope and binds each `let` *after* walking its initializer; a **lambda** pushes a child scope with its parameters; an **`ast.ImportClause`** walks each `with` expression in the importing rule's scope (the runtime evaluates them there); everything else delegates to the generic child walker. Used by `forEachBuiltinCall` and [[index.is_resolve]].
  - **Side Effects:** None; scopes are cloned, never mutated in place across branches.
  - **Exceptions:** `validation cancelled`.

- **Signature:** `(*kindCheckCtx).checkBuiltinCall(decl *builtins.Decl, c *ast.CallExpression) -> []error`
  - **Behavior:** Takes the declaration the walk already resolved. Computes `min` from the count of non-optional params and `max` from total params; reports too-many only when the signature is **non-variadic** and a `TooManyError` message exists, pointing the span at the first excess argument. Reports too-few pointing at the callee identifier. Then per argument: checks the value kind when statically known and the mismatch policy is not `MismatchUndefined`, synthesising an "expected callable" message when none was authored; and checks callback arity when the signature declares allowed arities.
  - **Side Effects:** None.
  - **Exceptions:** Returns the three `xerr` builtin diagnostics.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per invocation; scopes are values copied down the walk.
- **Performance/Scale Notes:** One walk per expression site and per derive body for each of the three visitors (this check, nondeterminism warnings, capabilities), on top of the walks performed by purity and cycle checking - the same ASTs are traversed several times per validation. `cloneBindingScope` copies on every block and lambda.
- **Dependencies Risk:**
  - **Checks are opt-in per signature.** A kind mismatch is only reported when `Kinds` is populated **and** `OnMismatch != MismatchUndefined` **and** a message exists (or the callable special case applies). A builtin declared without these fields is effectively unchecked, so absence of a diagnostic does not mean the call is correct.
  - **Unknown kinds pass silently.** `resolveKind` returns "not known" for most expressions, and unknown always means accept. This is deliberately biased toward **no false positives**, which means real errors routinely slip through to runtime.
  - **Shadowing precedence must match the runtime.** `isBuiltinCall` skips names bound locally or by a derive, mirroring the runtime's `getTarget` order. If the runtime's precedence changes, this check would validate against the wrong callee.
  - **Non-`VarDeclaration` statements in a block are skipped, not reported**, unlike the purity walk which rejects them - the two walkers disagree about how strict a block is.
  - **Type-constraint arguments are not walked**, including those of an `x is T` type or an `as T` cast. The parser accepts only literals there (`parseConstraintLiteral`), so no call can appear; if that ever loosens, the walk must learn about type refs.
  - **`forEachDeriveExprChild` has no `ast.ImportClause` case**, because a derive cannot contain one; `walkScopedExprs` handles it explicitly instead. Any other expression type that the child walker rejects is silently skipped by this walk.
  - **Diagnostics are joined, so output can be long.** A pack with a systematic mistake reports every occurrence at once.
