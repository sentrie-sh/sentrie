---
id: parser.literal
type: Module / File
language: Go
file_path: parser/literal.go
tags: constraints, compile-time-constants, literal-only, type-system
---

# Node: parser.literal (Constraint-Argument Literal Parser)

## 1. Architectural Role & Intent
`parser/literal.go` implements a **restricted, literal-only** expression parser used exclusively for type-constraint arguments (`string.between(1, 10)`, `number.oneOf([1, 2, 3])`). It exists to enforce a hard language rule: constraint arguments must be compile-time constants, never runtime expressions. Rather than parsing an expression and rejecting non-constants afterwards, it defines a parallel grammar that can only produce literals - making the restriction structural instead of a post-hoc check.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.literal` | `CALLS` | [[parser.primary]] | Delegates each scalar form to the shared leaf handlers. |
| `parser.literal` | `CALLS` | [[ast]] | Emits `ListLiteral` and `MapLiteral` for aggregate constraint arguments. |
| `parser.literal` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect`, `canExpect`, `hasTokens`, `errorf`. |
| [[parser.typeref]] | `CALLS` | [[parser.literal]] | The sole caller: constraint suffix arguments are parsed here. |
| [[ast.typeref]] | `DEPENDS_ON` | [[parser.literal]] | Stores the parsed literals as `TypeRefConstraint.Args`. |
| [[constraints]] | `READS_FROM` | [[parser.literal]] | Evaluates these argument nodes when applying a checker to a value. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseConstraintLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Switches on the head token and accepts exactly eight forms: string, int, float (including exponent forms such as `1e-3`, which the lexer emits as `Float`), a unary minus followed by an int or float (`parseConstraintNegativeLiteral`, #146), the three trinary keywords, null, `[` (literal list), and `{` (literal map). A lexer `Error` token (for example a malformed number literal `1e`) is reported with the lexer's own message. Everything else is rejected.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `constraint arguments must be literals, got %s` for any other token - the diagnostic that distinguishes "you wrote an expression" from a generic syntax error. The position (`file:line:col-col`) comes once, from the `parsing error at` prefix `errorf` adds; the message used to repeat it, formatting `Range.From` with `%s`, which printed `%!s(int=…)` (#146).

- **Signature:** `parseConstraintNegativeLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Consumes `-` and the number token after it and returns one negative `IntegerLiteral` or `FloatLiteral` spanning both, so the index's argument folding ([[index.constraint_args]]), the runtime and [[format]] see an ordinary literal. An integer below the int64 range becomes a float, as for a positive one (#166); `-9223372036854775808` is an exact `MinInt64`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `a minus in a constraint argument must be followed by a number literal, got %s` for `-x`, `-"a"`, `-(1)` or `-[1]`.

- **Signature:** `parseConstraintListLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `[ lit, lit, … ]` where every element recurses through `parseConstraintLiteral`, so nesting stays literal-only. Requires a comma between elements and permits a closing bracket immediately after any element.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing bracket, a non-literal element, or a missing comma.

- **Signature:** `parseConstraintMapLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `{ "key": lit, … }`. Keys must be **string literals** - unlike the general map literal in [[parser.collection]], computed `[expr]` keys are not accepted here. Values recurse through `parseConstraintLiteral`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `map keys must be string literals, got %s`; returns `nil` on a missing brace, colon, comma, or a non-literal value.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable; constraint arguments are small by construction.
- **Dependencies Risk:**
  - **The colon check does not consume.** In the map parser, a missing `:` is detected with `canExpect` and returns `nil` **without calling `errorf`**, so the parse aborts with no diagnostic recorded - the caller then reports something generic and misleading. This is the one silent-failure path in the file.
  - **A parallel grammar that can drift.** Two literal syntaxes now exist: this one and the general one in [[parser.collection]]. Adding a literal form to the language means updating both, and forgetting this file makes the new form unusable in constraints for no obvious reason.
  - **Negative numbers are folded here, not in general expressions.** Elsewhere `-1` is a `UnaryExpression` ([[parser.unary]]); in a constraint argument it is a negative literal, and only directly before a number. The map-key path still rejects anything but a string key.
  - **Trailing commas are accepted in lists** (the loop breaks on the closing bracket after consuming a comma) but the behaviour differs subtly between the list and map paths - do not assume symmetry.
