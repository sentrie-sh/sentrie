---
id: parser.typed_lambda
type: Function / Endpoint
language: Go
file_path: parser/typed_lambda.go
tags: lambda, type-annotations, optional-parameters, arity
---

# Node: parser.parseTypedLambdaExpression (Typed Lambda)

## 1. Architectural Role & Intent
Parses the full lambda form `(a: string, b?: number): trinary => { … }` — per-parameter type annotations, optional markers, and an optional return type. It is the fallback path taken when the fast untyped probe in [[parser.lambda]] declines, and it is where two real language rules are enforced: parameter names must be unique, and **required parameters may not follow optional ones**. That ordering rule is what makes `ast.RequiredLambdaArity` a meaningful prefix count at call time.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.typed_lambda` | `CALLS` | [[parser.typeref]] | Parses each parameter's type and the return type. |
| `parser.typed_lambda` | `CALLS` | [[parser.block]] | Requires the body to be a block expression. |
| `parser.typed_lambda` | `CALLS` | [[ast]] | Emits `ast.NewLambdaExpressionFull(names, types, opts, returnType, body, span)`. |
| [[parser.block]] | `CALLS` | [[parser.typed_lambda]] | Invoked from `parseGroupedExpression` after the fast probe fails. |
| [[parser.derive]] | `DEPENDS_ON` | [[parser.typed_lambda]] | Derives require a lambda; typed derives come through this path. |
| [[runtime.callable]] | `DEPENDS_ON` | [[parser.typed_lambda]] | Uses `RequiredLambdaArity` and the parallel type slices to bind and check arguments. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseTypedLambdaExpression(ctx, p, lparenFrom: tokens.Pos) -> *ast.LambdaExpression`
  - **Behavior:** Assumes `(` is already consumed and `p.current` is the first token inside it. Loops over parameters, each `ident ['?'] [': ' type]`, collecting three **parallel slices** (`names`, `types`, `opts`) where an untyped parameter contributes a `nil` type. Then an optional `: returnType`, a required `=>`, and a block body. The span runs from the caller-supplied `(` position to the body's end.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `required lambda parameter %q cannot follow an optional parameter`; `duplicate lambda parameter %q`; `expected ',' or ')' in lambda parameter list, got %s`; `lambda body must be a block expression { ... yield ... }`; `nil` on a missing `)`, `=>`, or a failed type reference.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Parallel slices are always fully populated here, but not on the fast path.** This production appends to `types`/`opts` for every parameter (using `nil`/`false` placeholders), whereas [[parser.lambda]] produces a lambda with **nil** `ParamTypes` and `ParamOpts`. Consumers must bounds-check both slices rather than assuming they parallel `Params` — see [[ast]].
  - **The optional-ordering rule is checked *after* the type is parsed**, so `(a?: string, b: number)` reports the error at `b` with the type already consumed. Correct, but the position points past the actual mistake.
  - **The `file` for the span comes from `p.current` at entry**, not from the `(` token, so a lambda whose parameter list somehow spanned files would mislabel — harmless today, fragile if includes are ever added.
  - **Entry conditions are implicit.** The function documents its precondition in a comment rather than checking it; calling it with `(` unconsumed silently misparses.
  - **A lambda body must be a block.** Expression-bodied lambdas (`(x) => x + 1`) are not supported anywhere in the language.
