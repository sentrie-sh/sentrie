---
id: parser.shape
type: Function / Endpoint
language: Go
file_path: parser/shape.go
tags: declaration, shape, structural-types, composition
---

# Node: parser.parseShapeStatement (Shape Declaration)

## 1. Architectural Role & Intent
Parses `shape <ident>` in its two mutually exclusive forms: a **simple** alias to a type reference (`shape Email string.matches(...)`) and a **complex** record with named fields, optionally composed from a base shape (`shape User with base/Shape { id: string, email?: Email }`). Shapes are Sentrie's named structural types, so this production is the entry point for user-defined type vocabulary that [[ast.typeref]] later references by FQN.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.shape` | `CALLS` | [[parser.typeref]] | Parses the simple form's type and each field's type, including constraints. |
| `parser.shape` | `CALLS` | [[parser.fqn]] | Parses the `with <fqn>` base-shape reference. |
| `parser.shape` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect`, `canExpect`, `canExpectAnyOf`, `peek`, `errorf`. |
| `parser.shape` | `CALLS` | [[ast]] | Emits `ast.NewShapeStatement`, `ast.Cmplx`, and `ast.ShapeField` values. |
| [[parser.statement]] | `CALLS` | [[parser.shape]] | Registered for `tokens.KeywordShape` at top level. |
| [[parser.policy]] | `CALLS` | [[parser.shape]] | Registered for the same kind at policy scope — the same handler serves both. |
| [[index.shape]] | `DEPENDS_ON` | [[ast]] | Resolves shape names, composition, and field types. |
| [[runtime.typeref_shape]] | `DEPENDS_ON` | [[ast]] | Validates values against the resolved shape at runtime. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseShapeStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `shape` and the name, then branches: a head of `{` or `with` selects the complex form, anything else is parsed as a simple type reference. Exactly one of `Simple`/`Complex` is populated on the resulting node.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing keyword or name, or when both forms fail to produce a value.

- **Signature:** `parseComplexShape(ctx: context.Context, p: *Parser) -> *ast.Cmplx`
  - **Behavior:** Optionally consumes `with <fqn>` as the composition base, then a required `{ … }` field block. Fields are collected into a **map keyed by field name**, with trailing/line comments skipped between fields. The range is extended to the closing brace.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on a failed base FQN, a missing `{` or `}`, or a failed field.

- **Signature:** `parseShapeField(ctx: context.Context, p: *Parser) -> *ast.ShapeField`
  - **Behavior:** Parses `name['?'] : <type>`. The `?` marks presence-optionality, mirroring the fact syntax. Carries three targeted migration errors that reject the retired `!` forms (`name!: T`, `name?!: T`, `name!?: T`) with the modern replacement spelled out.
  - **Side Effects:** Consumes tokens; emits `parseShapeField_start` / `_end` debug logs.
  - **Exceptions:** Returns `nil` for a missing name, any legacy `!` form, a missing `:`, or a type-reference failure.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless functions; the `Cmplx` value accumulates fields during the parse.
- **Performance/Scale Notes:** Two debug logs per field. Nothing else notable.
- **Dependencies Risk:**
  - **Fields are a map, so source order is lost and duplicates are silently overwritten.** `shape S { a: string, a: number }` parses without complaint and keeps only the last `a`. Any duplicate-field diagnostic must be added in [[index.shape]] — it cannot be recovered from the AST.
  - **Unterminated field blocks can run away.** The field loop tests only for `}`, not EOF, so a missing closing brace produces an error from deep inside field parsing rather than at the shape header.
  - **Field type errors are detected via `p.err`, not a nil check.** `parseShapeField` assigns `field.Type` and then checks the parser error; if a future change let `parseTypeRef` return nil without recording an error, the subsequent `field.Type.Span()` would panic.
  - **Composition is unresolved here.** `with <fqn>` is stored raw; whether the base exists, is visible, and whether fields conflict is entirely [[index.shape]]'s responsibility.
  - **Simple and complex are exclusive but untyped as such.** Consumers must check which of `Simple`/`Complex` is non-nil; the AST does not model it as a sum type.
