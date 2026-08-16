---
id: parser.transform
type: Function / Endpoint
language: Go
file_path: parser/transform.go
tags: transform, jq, document-manipulation, prefix-handler
---

# Node: parser.parseTransformExpression (JQ Transform)

## 1. Architectural Role & Intent
Parses `transform <expr> with "<jq>"`, applying a JQ-compatible transformation program to a value — the escape hatch for reshaping documents that would be tedious to navigate with field access alone. Keeping the transformer as an opaque string literal means the JQ grammar is never parsed by Sentrie; it is passed through to the runtime's transform engine intact.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.transform` | `CALLS` | [[parser.expression]] | Parses the transform argument at `LOWEST`. |
| `parser.transform` | `CALLS` | [[ast]] | Emits `ast.NewTransformExpression(argument, transformerString, span)`. |
| [[parser.lookups]] | `CALLS` | [[parser.transform]] | Registered as the prefix handler for `KeywordTransform`. |
| [[runtime.eval_transform]] | `DEPENDS_ON` | [[parser.transform]] | Compiles and applies the JQ program to the evaluated argument. |
| [[box.value]] | `READS_FROM` | [[parser.transform]] | Documents and dicts are the usual transform inputs. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseTransformExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Consumes `transform`, parses the argument expression at `LOWEST`, requires the `with` keyword, and requires a **string literal** transformer — not an expression, so the program cannot be computed at runtime. The span covers keyword through transformer string.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `transform`/`with` keyword, a failed argument expression, or a transformer that is not a string literal.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Parse-time cost is trivial; the runtime cost of compiling and applying the JQ program is not, and nothing here caches it.
- **Dependencies Risk:**
  - **The argument is parsed at `LOWEST`, so it swallows everything up to `with`.** `transform a + b with "..."` transforms the sum, and a missing `with` produces an error far from the `transform` keyword because the expression parser ran to the end of the construct first.
  - **The transformer is unvalidated at parse time.** A syntactically invalid JQ program parses cleanly and fails only when evaluated, so authoring errors surface late and without a source position inside the string.
  - **A heredoc works here.** `with <<<JQ … JQ` produces a `String` token like any other, which is the practical way to write multi-line transform programs — see [[lexer]].
  - **`with` is overloaded** — the transform separator here, the shape composition keyword in [[parser.shape]], and the import binding keyword in [[parser.import]].
