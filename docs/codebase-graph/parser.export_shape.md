---
id: parser.export_shape
type: Function / Endpoint
language: Go
file_path: parser/export_shape.go
tags: declaration, visibility, namespace-export, shape, derive
---

# Node: parser.parseExportStatement (Namespace-Level Export)

## 1. Architectural Role & Intent
Parses the **top-level** `export` form, dispatching on the following keyword into `export shape <ident>` or `export derive <ident>`. Sentrie's default visibility is namespace-private, so this production is the sole mechanism by which a shape or derive becomes visible to other namespaces — making it the front-end half of the visibility model that [[index.resolve]] enforces.

The `export` keyword is context-sensitive: at top level it reaches this handler, while inside a policy block it dispatches to [[parser.export_rule]] instead. Reading either node alone gives half the picture.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.export_shape` | `CALLS` | [[parser.parser]] | Uses `head`, `advance`, `expect`, `advanceExpected`, `errorf`. |
| `parser.export_shape` | `CALLS` | [[ast]] | Emits `ast.NewShapeExportStatement` or `ast.NewExportDeriveStatement`. |
| [[parser.statement]] | `CALLS` | [[parser.export_shape]] | Registered for `tokens.KeywordExport` in the **top-level** table. |
| [[index.resolve]] | `DEPENDS_ON` | [[parser.export_shape]] | Uses these statements to decide cross-namespace visibility. |
| [[parser.shape]] | `DEPENDS_ON` | [[parser.export_shape]] | Declares the shape that this statement exports by name. |
| [[parser.derive]] | `DEPENDS_ON` | [[parser.export_shape]] | Declares the derive that this statement exports by name. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseExportStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `export`, then switches on the next token: `shape` and `derive` delegate to the helpers below; anything else is an error. This two-token dispatch is why `export` can be a single table entry serving two declaration forms.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` with `expected 'shape' or 'derive' after export` for any other continuation.

- **Signature:** `parseShapeExportAfterExport(ctx, p, exportHead: tokens.Instance) -> ast.Statement`
  - **Behavior:** Consumes `shape` and the identifier, building a span from the original `export` token to the name so the whole phrase is covered.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` for a missing keyword or identifier.

- **Signature:** `parseExportDeriveAfterExport(ctx, p, exportHead: tokens.Instance) -> ast.Statement`
  - **Behavior:** Same shape, emitting `ExportDeriveStatement`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` for a missing keyword or identifier.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Export is a separate statement, not a modifier.** `export shape Foo` does not declare `Foo` — it references a shape declared elsewhere in the namespace. Exporting a name that was never declared is not detected here; [[index.package]] must catch the dangling reference.
  - **Only namespace-level derives are exportable.** The policy-scope counterpart in [[parser.export_rule]] rejects `export derive` explicitly. If you are tracing why an export was refused, check which scope the statement was written in.
  - **Nothing prevents duplicate exports** of the same name, and nothing here associates the export with its declaration — the linkage is by string name, resolved later.
  - **The file name understates its scope.** `export_shape.go` also owns the derive-export path; searching for "export derive" by filename will miss it.
