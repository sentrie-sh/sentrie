---
id: parser.is
type: Function / Endpoint
language: Go
file_path: parser/is.go
tags: predicates, definedness, emptiness, type-check, infix-handler, undefined-semantics
---

# Node: parser.parseIsExpression (Definedness, Emptiness and Type Predicates)

## 1. Architectural Role & Intent
Parses the `is` family: `x is defined`, `x is empty`, the type check `x is T` (#150), their negated forms with an optional `not`, and the legacy `a is b` equality. The definedness predicate is architecturally significant - it is the syntax through which a policy can *ask* whether a fact was supplied rather than having its absence silently propagate as `Unknown` through [[trinary]] logic, which is the difference between an explicit "I cannot decide" and an accidental one.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.is` | `CALLS` | [[parser.expression]] | Parses the right operand when it does not start with a type keyword. |
| `parser.is` | `CALLS` | [[parser.typeref]] | `parseBaseTypeRef` and the ternary-aware `parseTypeRefSuffixes` for a type right-hand side. |
| `parser.is` | `CALLS` | [[ast]] | Emits `IsDefinedExpression`, `IsEmptyExpression`, `IsExpression`, or `InfixExpression`, optionally wrapped in `UnaryExpression`. |
| `parser.is` | `CALLS` | [[parser.parser]] | Uses `expect(KeywordIs)`, `head`, `advance`, `canExpect`. |
| [[parser.lookups]] | `CALLS` | [[parser.is]] | Registered as the infix handler for `KeywordIs` at `EQUALITY` precedence. |
| [[box.value]] | `READS_FROM` | [[parser.is]] | `IsDefined` resolves against the boxed undefined sentinel rather than null. |
| [[runtime.eval]] | `DEPENDS_ON` | [[parser.is]] | Evaluates the produced forms. |
| [[index.is_resolve]] | `DEPENDS_ON` | [[parser.is]] | Settles the name right-hand side the parser leaves ambiguous. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseIsExpression(ctx, p, left: ast.Expression, precedence: Precedence) -> ast.Expression`
  - **Behavior:** Consumes `is`, optionally consumes `not`, then dispatches on the follower:
    - `defined` → `IsDefinedExpression(left)`
    - `empty` → `IsEmptyExpression(left)`
    - a type keyword (`string`, `number`, `boolean`, `trinary`, `document`, `list`, `dict`, `record`), or the identifier `set` followed by `[` (#162) → a full type reference, built as `IsExpression(left, Type)`. So `x is set[0]` with a value named `set` is no longer an equality: it fails to parse as a type
    - otherwise parses a right operand at `EQUALITY` precedence:
      - a bare or slash-qualified identifier followed by `@` or a nullable `?` → that name as a shape type with its suffixes, `IsExpression(left, Type)`
      - a bare or slash-qualified identifier with no suffix → `IsExpression(left, Right, Type)` carrying both readings for [[index.is_resolve]]
      - anything else → `InfixExpression(left, "is", right)`, the legacy equality

    A captured `not` wraps whichever node was produced in a `UnaryExpression`, so `is not defined` desugars to `not (is defined)`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` if `is` is missing, or if the right operand or type fails to parse (the underlying error having been recorded by the operand's own handler).

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **The general-comparison branch guards against a nil right operand** before calling `right.Span()`, matching [[parser.not]]; incomplete input such as `x is` records a parse error instead of panicking.
  - **`?` after a type is read ternary-first.** A `?` followed by a token that can start an expression is the ternary operator (`x is string ? a : b`); otherwise it is the nullable suffix (`x is string?`). After a type keyword, a `?` before `{` is also nullable, so `when x is string? { ... }` works (`Parser.questionIsNullable(true)`). After a bare or qualified name, which may be a value, `?` before `{` stays the ternary that `x is y ? {...} : {...}` has always been, so a nullable shape before a rule body needs parentheses, `when (x is User?) { ... }`, or no `?` (a non-null value that conforms to `User?` conforms to `User`).
  - **A name on the right is not classified here.** Whether `x is U` is a shape check or equality depends on the shapes and bindings in scope, which only [[index.is_resolve]] knows; an unvalidated tree still has both readings.
  - **The predicate spans are truncated.** `IsDefinedExpression` and `IsEmptyExpression` are built with the range of the `is` token captured *before* the follower is consumed, so their spans cover neither the left operand nor the `defined`/`empty` keyword.
  - **`defined` vs `null` are different questions.** `is defined` tests presence (the undefined sentinel in [[box.value]]), not nullness - a fact explicitly set to null **is** defined. Conflating them is the most common semantic mistake with this operator.
  - **`is not X` is desugared, not a distinct node**, so tooling cannot distinguish `x is not defined` from `not (x is defined)`.
