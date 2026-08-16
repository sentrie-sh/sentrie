---
id: parser.literal
type: Module / File
language: Go
file_path: parser/literal.go
tags: constraints, compile-time-constants, literal-only, type-system
---

# Node: parser.literal (Constraint-Argument Literal Parser)

## 1. Architectural Role & Intent
`parser/literal.go` implements a **restricted, literal-only** expression parser used exclusively for type-constraint arguments (`string.between(1, 10)`, `number.oneOf([1, 2, 3])`). It exists to enforce a hard language rule: constraint arguments must be compile-time constants, never runtime expressions. Rather than parsing an expression and rejecting non-constants afterwards, it defines a parallel grammar that can only produce literals — making the restriction structural instead of a post-hoc check.

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
  - **Behavior:** Switches on the head token and accepts exactly seven forms: string, int, float, the three trinary keywords, null, `[` (literal list), and `{` (literal map). Everything else is rejected.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `constraint arguments must be literals, got %s at %s` for any other token — the diagnostic that distinguishes "you wrote an expression" from a generic syntax error.

- **Signature:** `parseConstraintListLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `[ lit, lit, … ]` where every element recurses through `parseConstraintLiteral`, so nesting stays literal-only. Requires a comma between elements and permits a closing bracket immediately after any element.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing bracket, a non-literal element, or a missing comma.

- **Signature:** `parseConstraintMapLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `{ "key": lit, … }`. Keys must be **string literals** — unlike the general map literal in [[parser.collection]], computed `[expr]` keys are not accepted here. Values recurse through `parseConstraintLiteral`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `map keys must be string literals, got %s at %s`; returns `nil` on a missing brace, colon, comma, or a non-literal value.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable; constraint arguments are small by construction.
- **Dependencies Risk:**
  - **The colon check does not consume.** In the map parser, a missing `:` is detected with `canExpect` and returns `nil` **without calling `errorf`**, so the parse aborts with no diagnostic recorded — the caller then reports something generic and misleading. This is the one silent-failure path in the file.
  - **A parallel grammar that can drift.** Two literal syntaxes now exist: this one and the general one in [[parser.collection]]. Adding a literal form to the language means updating both, and forgetting this file makes the new form unusable in constraints for no obvious reason.
  - **Negative numbers are not literals here.** Unary minus is an *expression*, not a literal token, so `number.min(-1)` is rejected as "constraint arguments must be literals". Any negative bound must be expressed another way.
  - **Trailing commas are accepted in lists** (the loop breaks on the closing bracket after consuming a comma) but the behaviour differs subtly between the list and map paths — do not assume symmetry.
