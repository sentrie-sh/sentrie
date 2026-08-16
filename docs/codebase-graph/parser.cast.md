---
id: parser.cast
type: Function / Endpoint
language: Go
file_path: parser/cast.go
tags: cast, type-conversion, infix-handler
---

# Node: parser.parseAsCastExpression (Type Cast)

## 1. Architectural Role & Intent
Parses the postfix cast `expr as <type>` as an infix handler on `as`. It is the only syntactic bridge between the untyped expression world and the type vocabulary of [[ast.typeref]], letting a policy assert or convert a value's type mid-expression — typically to satisfy a shape or to coerce a document field into a scalar before comparison.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.cast` | `CALLS` | [[parser.typeref]] | Parses the target type, including constraints and nullability. |
| `parser.cast` | `CALLS` | [[ast]] | Emits `ast.NewCastExpression(left, targetType, span)`. |
| [[parser.lookups]] | `CALLS` | [[parser.cast]] | Registered as the infix handler for `KeywordAs`. |
| [[parser.precedence]] | `DEPENDS_ON` | [[parser.cast]] | `as` sits at `UNARY`, above all arithmetic and comparison operators. |
| [[runtime.eval_cast]] | `DEPENDS_ON` | [[parser.cast]] | Performs the actual conversion and constraint check against [[box.value]]. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseAsCastExpression(ctx, p, left: ast.Expression, _ Precedence) -> ast.Expression`
  - **Behavior:** Consumes `as`, parses a type reference, and builds a `CastExpression` spanning from the left operand's start to the type's end — a correctly computed range, unlike most of the expression productions. The precedence argument is explicitly ignored, since the target is a type reference rather than an expression and needs no binding power.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `as` or a failed type reference.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **`as` is heavily overloaded across the language.** It is a cast here, the fact alias in [[parser.fact]], the module alias in [[parser.use]], the attachment binder in [[parser.export_rule]], and the `with … as` binder in [[parser.import]]. Only this use is an expression operator.
  - **High precedence makes casts bind tighter than they read.** Because `as` sits at `UNARY`, `a + b as string` casts only `b`. Parenthesise when the intent is to cast a whole arithmetic result.
  - **Casts may carry constraints.** `x as string@minLength(3)` is legal and the constraint is checked at cast time, which makes a cast a validation site and not merely a conversion.
  - **No parse-time type checking.** Whether the cast is possible is entirely a runtime question.
