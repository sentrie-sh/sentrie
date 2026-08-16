---
id: parser.let
type: Function / Endpoint
language: Go
file_path: parser/let.go
tags: declaration, binding, local-variable, type-annotation
---

# Node: parser.parseLetsStatement (Variable Declaration)

## 1. Architectural Role & Intent
Parses `let <ident> [: <type>] = <expr>`, the local binding form used inside policies and blocks to name intermediate values. The optional type annotation makes it the one declaration where a type reference is genuinely optional, and the initialiser is mandatory — Sentrie has no uninitialised bindings, which is what lets the evaluator treat every `let` as a pure definition rather than a mutable slot.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.let` | `CALLS` | [[parser.typeref]] | Parses the optional `: T` annotation. |
| `parser.let` | `CALLS` | [[parser.expression]] | Parses the mandatory initialiser at `LOWEST`. |
| `parser.let` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect(TokenAssign)`, `canExpect`. |
| `parser.let` | `CALLS` | [[ast]] | Emits `ast.NewVarDeclaration(name, typeRef, value, span)`. |
| [[parser.policy]] | `CALLS` | [[parser.let]] | Registered for `tokens.KeywordLet` in the policy-scope table. |
| [[parser.block]] | `CALLS` | [[parser.let]] | Block bodies accept `let` statements before their `yield`. |
| [[runtime.eval]] | `DEPENDS_ON` | [[parser.let]] | Binds the evaluated value into the current scope. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseLetsStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `let` **unconditionally** (via bare `advance()`, without verifying the keyword — the handler table guarantees it), then the name, then an optional `: <type>`, then a required `=`, then the initialiser. The span runs from `let` to the end of the initialiser.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing name, a failed type reference, a missing `=`, or a failed initialiser expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Leftover debug scaffolding.** The production contains a dead `if nameIdent.Value == "tri"` branch assigning to an unused local — a breakpoint hook that should be removed. It has no behavioural effect but will confuse anyone reading the file for logic.
  - **Span end is set from the colon, not the type.** When an annotation is present, `rnge.To` is assigned the **colon's** end position before the initialiser overwrites it, so intermediate spans are wrong; the final assignment from the value expression corrects it. Any change that makes the initialiser optional would expose the bug.
  - **The keyword is consumed unchecked.** Calling this function directly (outside table dispatch) with a different head token silently swallows that token. It is only safe because registration guarantees the head.
  - **The annotation is not enforced here.** Whether the initialiser satisfies the declared type is a later concern — this production only records both.
