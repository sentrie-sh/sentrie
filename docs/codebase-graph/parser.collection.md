---
id: parser.collection
type: Module / File
language: Go
file_path: parser/collection.go
tags: literals, list, map, aggregate-expressions
---

# Node: parser.collection (List and Map Literals)

## 1. Architectural Role & Intent
`parser/collection.go` parses the two aggregate literal forms - `[a, b, c]` and `{ "k": v, [expr]: v }` - with **arbitrary expressions** as elements, values, and (in bracket form) keys. It is the general-purpose counterpart to the literal-only constraint grammar in [[parser.literal]], and the computed-key syntax is what lets policies build maps whose keys derive from facts rather than being fixed at authoring time.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.collection` | `CALLS` | [[parser.expression]] | Elements, values, and computed keys are parsed at `LOWEST`. |
| `parser.collection` | `CALLS` | [[ast]] | Emits `ast.NewListLiteral` and `ast.NewMapLiteral` with `ast.MapEntry` pairs. |
| `parser.collection` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect`, `canExpect`, `hasTokens`, `errorf`. |
| [[parser.lookups]] | `CALLS` | [[parser.collection]] | `parseListLiteral` is the prefix handler for `[`. |
| [[parser.left_curly]] | `CALLS` | [[parser.collection]] | `parseMapLiteral` is reached only through the `{` disambiguation, never registered directly. |
| [[box.value]] | `DEPENDS_ON` | [[parser.collection]] | List and dict values are materialised from these nodes at evaluation. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseListLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `[ expr, expr, … ]`, permitting an empty list and a trailing comma. Each element is a full expression, so lists may contain calls, pipelines, and nested collections. The span covers both brackets.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `[` or `]` or a failed element expression.

- **Signature:** `parseMapLiteral(ctx, p) -> ast.Expression`
  - **Behavior:** Parses `{ key : value, … }` where a key is **either** a string literal **or** a bracketed computed expression `[expr]`. Values are full expressions. Consumes the opening `{` unconditionally, since it is only reached after [[parser.left_curly]] has already decided this is a map.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** `expected string or [expression] as map key, got %s at %s`; returns `nil` on a missing `:`, `}`, or a failed value expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One nested `parseExpression` per element/value; cost is proportional to literal size.
- **Dependencies Risk:**
  - **Commas are optional between entries.** Both loops consume a comma only *if present* and then continue, so `[1 2 3]` and `{"a": 1 "b": 2}` parse without complaint as three-element and two-entry literals. Missing separators are silently accepted rather than diagnosed.
  - **Duplicate map keys are retained.** Entries are a slice, not a map, so `{"a": 1, "a": 2}` produces two entries and the collision-resolution rule is whatever [[runtime.eval]] does when materialising the dict - it is not decided here.
  - **`[` is doubly overloaded.** As a prefix it starts a list literal; as an infix it is index access ([[parser.access]]); inside a map literal it delimits a computed key. All three are the same token kind, disambiguated purely by position.
  - **Unterminated literals rely on `hasTokens()`.** Reaching EOF exits the loop and the subsequent `advanceExpected` reports `expected ], got EOF`, positioned at end-of-file rather than at the opening bracket.
