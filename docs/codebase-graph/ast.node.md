---
id: ast.node
type: Class
language: Go
file_path: ast/node.go
tags: interface-contract, sealed-hierarchy, source-positions, type-switch
---

# Node: ast.Node (Sealed Node Hierarchy Root)

## 1. Architectural Role & Intent
`ast/node.go` declares the four interfaces and one embedded base struct that every Sentrie syntax node conforms to: `Positionable`, `Node`, `Statement`, `Expression`, and `baseNode`. It exists to guarantee two invariants across the entire tree — that *every* node can report the exact source range it came from, and that the statement/expression partition is **sealed** to the `ast` package via unexported marker methods, making downstream type switches safe to treat as exhaustive.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `ast.node` | `DEPENDS_ON` | [[tokens]] | `baseNode.Rnge` is a `tokens.Range`; `Span()` is the universal accessor for it. |
| [[ast]] | `INHERITS_FROM` | [[ast.node]] | Every concrete node type embeds `*baseNode` and thereby satisfies `Node`. |
| [[ast.typeref]] | `INHERITS_FROM` | [[ast.node]] | `baseTypeRef` embeds `*baseNode`, so `TypeRef` is a strict extension of `Node`. |
| [[parser]] | `CALLS` | [[ast.node]] | Sets each node's span from the token range consumed by the production. |
| [[index.package]] | `CALLS` | [[ast.node]] | Reads `Span()` to attach file/line/column to validation diagnostics. |
| [[runtime]] | `CALLS` | [[ast.node]] | Reads `Span()` to position runtime errors and trace entries; type-switches on `Statement`/`Expression`. |
| [[xerr]] | `DEPENDS_ON` | [[tokens]] | Diagnostics carry the span obtained from `Span()` into the rendered error. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Positionable` interface — `Span() -> tokens.Range`
  - **Behavior:** Minimal positional contract. Kept separate from `Node` so that helper types which are not full nodes can still be positioned.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Node` interface — `Positionable` + `String() -> string` + `Kind() -> string`
  - **Behavior:** The universal node contract. `Kind()` returns the snake_case discriminator string set by the constructor (`"rule_statement"`, `"call"`, `"block"`, …) and is the cheap, allocation-free way to identify a node without reflection or a type assertion. `String()` renders an approximate source form for diagnostics.
  - **Side Effects:** `String()` allocates and recurses over children.
  - **Exceptions:** None.

- **Signature:** `Statement` interface — `Node` + `statementNode()` (unexported)
  - **Behavior:** Marks declaration-level nodes. Because the marker is unexported, **no package outside `ast` can define a `Statement`**; a type switch over the known statement types is therefore closed and a `default` branch signals a missing case in this package rather than a foreign implementation.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Expression` interface — `Node` + `expressionNode()` (unexported)
  - **Behavior:** Marks evaluable nodes, sealed identically to `Statement`. Note `BlockExpression` and `LambdaExpression` are expressions, not statements — Sentrie blocks yield values.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `baseNode` struct (unexported) — `{ Rnge tokens.Range, Kind_ string }` with `Span()` and `Kind()`
  - **Behavior:** The shared implementation embedded as a **pointer** (`*baseNode`) by every concrete node. Its fields are exported-by-name within the package so constructors can set them positionally.
  - **Side Effects:** `Rnge.To` is mutated in place by `TypeRef.AddConstraint` to extend a type reference's span over its trailing constraints.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Effectively immutable after construction. The one exception is the span-extension mutation performed by [[ast.typeref]] during parsing.
- **Performance/Scale Notes:** `*baseNode` is embedded by pointer, so every node costs an extra allocation and an extra pointer hop on `Span()`/`Kind()`. `Kind()` is a string comparison — cheap, but a concrete type switch is still faster and is what [[runtime]] uses on hot paths; reserve `Kind()` for logging, serialization, and generic tooling.
- **Dependencies Risk:** No external failure domain. The hazards are structural:
  - **Nil `baseNode` panics.** A node constructed as a bare struct literal (bypassing its `New*` constructor) has a nil `*baseNode`, and the first `Span()` or `Kind()` call panics. Always construct through the package constructors.
  - **Mixed value/pointer receivers.** Several node types declare `String()`/`statementNode()` on the *value* receiver while `Span()`/`Kind()` come from the embedded pointer, so only the `*T` form satisfies `Node`. Always hold nodes as pointers; the `var _ Node = &X{}` assertions in each file confirm which form is the interface-satisfying one.
  - **`Kind()` strings are an untyped contract.** They are plain strings with no constant declarations, so a typo in a consumer's comparison fails silently at runtime rather than at compile time. Prefer type switches where correctness matters.
