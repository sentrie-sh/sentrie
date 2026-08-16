---
id: ast
type: System / Package
language: Go
file_path: ast/
tags: syntax-tree, front-end, data-model, code-generation, ir
---

# Node: AST (Abstract Syntax Tree Node Family)

## 1. Architectural Role & Intent
`ast` defines the complete node vocabulary of the Sentrie language - 50+ statement, expression, and type-reference types spread one-per-file - and is the sole data contract between the front-end ([[parser]]) and everything downstream ([[index]], [[runtime]]). It exists to keep syntax representation free of evaluation concerns: nodes are plain immutable-by-convention structs carrying source spans and child references, with no evaluation, no symbol resolution, and no environment. Its second responsibility is *parse-time* type-constraint arity validation, driven by a table generated from the canonical definitions in [[constraints]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `ast` | `LAYERED_ON` | [[tokens]] | Every node embeds a `tokens.Range` span via `baseNode`; `FQN`, literals, and type refs all carry positions for diagnostics. |
| `ast` | `LAYERED_ON` | [[trinary]] | `TrinaryLiteral` stores a resolved `trinary.Value` rather than raw text. |
| `ast` | `LAYERED_ON` | [[xerr]] | `validateConstraint` returns `xerr.NotFoundError{}` for unknown constraint names on a type ref. |
| `ast` | `LAYERED_ON` | [[constraints]] | **Build-time only.** `ast/gen.go` (guarded by the `generate` build tag) reads the constraint checker tables and emits `typeref_constraint_args_gen.go`. The compiled package has no runtime edge to [[constraints]]. |
| [[parser]] | `CALLS` | [[ast]] | Every production calls a `New*` constructor; this is the parser's only output type. |
| [[index]] | `LAYERED_ON` | [[ast]] | Walks `Program.Statements` to build namespaces, policies, rules, derives, shapes, and the dependency graph. |
| [[runtime]] | `LAYERED_ON` | [[ast]] | The evaluator switches on concrete node types to evaluate expressions and execute statements. |
| [[loader]] | `LAYERED_ON` | [[ast]] | `LoadPrograms` returns `[]*ast.Program`, one per discovered policy file. |
| [[pack]] | `LAYERED_ON` | [[ast]] | `pack.Pack` pairs a `PackFile` manifest with the parsed `[]*ast.Program`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Program` struct - `{ Statements []Statement, Reference string }`
  - **Behavior:** Root container for one parsed source file. `Reference` is the originating file path. Note it is a bare struct - it is **not** a `Node` and carries no span of its own.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Node` interface - `Span() tokens.Range`, `String() string`, `Kind() string`
  - **Behavior:** Universal node contract. `Kind()` returns a stable snake_case discriminator string (`"rule_statement"`, `"lambda"`, `"list_typeref"`, …) supplied by each constructor; `String()` renders an approximate source form used in diagnostics and debug output, **not** a faithful pretty-printer.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Statement` interface (`Node` + private `statementNode()`) / `Expression` interface (`Node` + private `expressionNode()`)
  - **Behavior:** Sealed marker interfaces partitioning the node set. The unexported marker methods make the hierarchy closed to this package, so downstream type switches can be treated as exhaustive.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** Declaration statements - `NamespaceStatement`, `PolicyStatement`, `FactStatement`, `UseStatement`, `VarDeclaration`, `RuleStatement`, `ShapeStatement`, `DeriveStatement`, `ExportDeriveStatement`, `RuleExportStatement`, `ShapeExportStatement`, `ImportClause`, `WithClause`, `AttachmentClause`, `CommentStatement`, and the policy-metadata trio `TitleStatement`/`DescriptionStatement`/`VersionStatement`/`TagStatement`
  - **Behavior:** Structural surface of a policy pack. `RuleStatement` carries `Default`, `When`, and `Body` as independent optional expressions; `FactStatement` carries type, alias, default expression, and an `Optional` flag; `UseStatement` distinguishes a relative path (`RelativeFrom`) from a library reference (`LibFrom`).
  - **Side Effects:** None - all are inert data.
  - **Exceptions:** None; constructors cannot fail.

- **Signature:** Expressions - `BlockExpression`, `LambdaExpression`, `CallExpression`, `InfixExpression`, `UnaryExpression`, `TernaryExpression`, `CastExpression`, `TransformExpression`, `FieldAccessExpression`, `IndexAccessExpression`, `PipelineHoleExpression`, `IsDefinedExpression`, `IsEmptyExpression`, `Identifier`, `TrailingCommentExpression`, `PrecedingCommentExpression`, and the literal family (`Integer`, `Float`, `String`, `Trinary`, `Null`, `List`, `Map`)
  - **Behavior:** The evaluable surface. `BlockExpression` pairs statements with a mandatory `Yield` expression - blocks are expressions, not statements. `CallExpression` carries `Memoized`/`MemoizeTTL` so memoization is a *syntactic* property resolved by [[runtime]]. `PipelineHoleExpression` is the `#` placeholder substituted by `|>`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `FQN` struct + `NewFQN(parts, span)`, `CreateFQN(base, lastSegment)`, `String()`, `LastSegment()`, `Parent()`, `IsChildOf(other)`, `IsParentOf(other)`, `IsEmpty()`, `Ptr()`
  - **Behavior:** Slash-separated fully-qualified namespace path with hierarchy predicates used by [[index]] for visibility and resolution.
  - **Side Effects:** `CreateFQN` clones the base parts, so derived FQNs never alias the parent's slice.
  - **Exceptions:** None.

- **Signature:** `TypeRef` interface - `Node` + `typeref()`, `GetConstraints() []*TypeRefConstraint`, `AddConstraint(*TypeRefConstraint) error`
  - **Behavior:** The type-annotation family (`String`, `Number`, `Trinary`, `List`, `Dict`, `Record`, `Shape`, `Document`, `Nullable`). See [[ast.typeref]] for the full contract.
  - **Side Effects:** `AddConstraint` mutates the receiver and extends its span.
  - **Exceptions:** `xerr.NotFoundError` for an unknown constraint name; a plain error for wrong argument count.

- **Signature:** `RequiredLambdaArity(lam: *LambdaExpression) -> int`
  - **Behavior:** Counts non-optional parameters. This is the arity [[runtime]] pre-checks before invoking a callable, so optional (`?`) parameters do not make a call under-applied.
  - **Side Effects:** None.
  - **Exceptions:** None; returns `0` for a nil lambda rather than panicking.

## 4. Operational Context & Gotchas
- **Statefulness:** Nodes are heap-allocated structs constructed once by the parser and thereafter treated as read-only. There is no interning, no parent pointers, and no visitor framework - traversal is open-coded type switching in [[index]] and [[runtime]]. `TypeRef.AddConstraint` is the one legitimate post-construction mutation and happens during parsing only.
- **Performance/Scale Notes:** Every node embeds `*baseNode` (a pointer), so each node costs at least two allocations. `String()` implementations recurse over the whole subtree and allocate - never call them in evaluation hot paths, only in diagnostics.
- **Dependencies Risk:** No external failure domain, but several structural hazards:
  - **Generated table drift.** `typeref_constraint_args_gen.go` is produced from [[constraints]] under the `generate` build tag and the generator **panics unless `GIT_USER_NAME` and `GIT_USER_EMAIL` are set**. Adding a constraint to [[constraints]] without re-running `go generate` means the parser rejects it as unknown while the runtime would have accepted it.
  - **No structural validation.** Constructors accept anything - nil bodies, empty parameter lists, mismatched parallel slices. `LambdaExpression.ParamTypes` and `ParamOpts` are *parallel slices* that may be nil or shorter than `Params`; every reader must bounds-check rather than index directly.
  - **Interface-typed children.** `Statement`/`Expression`/`TypeRef` fields are frequently nil for optional syntax (`RuleStatement.When`, `FactStatement.Default`, `LambdaExpression.ReturnType`). A nil check is required before every recursive descent.
  - **Comments in the tree.** Comments survive as both statements and expression wrappers (`TrailingCommentExpression`, `PrecedingCommentExpression`), so consumers must unwrap them before matching on expression shape.
