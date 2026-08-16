---
id: parser.parse
type: Function / Endpoint
language: Go
file_path: parser/parse.go
tags: entrypoint, program-structure, namespace-enforcement, front-end
---

# Node: parser.ParseProgram (Top-Level Parse Driver)

## 1. Architectural Role & Intent
`parser/parse.go` holds `ParseProgram`, the sole exported parsing entrypoint and the only place where a whole-file structural rule is enforced: **the first non-comment statement must be a `namespace` declaration, and no second namespace may appear.** It exists to wrap the statement loop with that invariant plus the comment/semicolon housekeeping that individual productions deliberately do not handle, producing the `ast.Program` that every downstream stage consumes.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.parse` | `CALLS` | [[parser.statement]] | Delegates each statement to `parseStatement`, which dispatches on the top-level handler table. |
| `parser.parse` | `CALLS` | [[parser.parser]] | Uses `hasTokens`, `canExpect`, and `advance` to drive the token window. |
| `parser.parse` | `CALLS` | [[ast]] | Builds `ast.Program` and wraps stray trailing comments into `ast.NewCommentStatement`. |
| `parser.parse` | `DEPENDS_ON` | [[tokens]] | Recognises `PunctSemicolon` and `TrailingComment` for housekeeping. |
| [[loader]] | `CALLS` | [[parser.parse]] | `LoadPrograms` calls this once per discovered `.sentrie` file and collects the results. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Parser).ParseProgram(ctx: context.Context) -> (*ast.Program, error)`
  - **Behavior:** Executes in three phases.
    1. **Prologue** — loops parsing statements, appending any `CommentStatement` (and its immediately following trailing comment) to the program, until the first non-comment statement is found.
    2. **Namespace gate** — type-asserts that statement to `*ast.NamespaceStatement` and fails the whole parse if it is not; then consumes an optional `;` and an optional trailing comment.
    3. **Body loop** — parses statements to EOF, rejecting any further `NamespaceStatement`, appending each statement plus any trailing comment, and consuming optional semicolons between statements.

    Returns `(nil, nil)` for a file with no tokens, and a program containing only comment statements for a comment-only file — both are legal, non-error outcomes.
  - **Side Effects:** Consumes the entire token stream and therefore the underlying reader. Sets `p.err` on failure. Note the context is threaded into productions for logging but **is never checked for cancellation here** — a parse runs to completion regardless.
  - **Exceptions:**
    - Any accumulated `p.err` from a statement production, returned as the joined error set with a `nil` program.
    - `program must start with namespace, got %T at <span>` when the first non-comment statement is anything else.
    - `namespace cannot be declared after namespace declaration at <span>` for a second namespace.
    - `failed to parse statement at line L, column C` when a production returns `nil` without having recorded an error.

- **Signature:** `ast.Program` (produced) — `{ Statements []ast.Statement, Reference string }`
  - **Behavior:** `Reference` is the filename handed to `NewParser`. `Statements` preserves source order **including comments**, so consumers cannot assume every element is a declaration.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** Not reentrant. It drives the parser to EOF and can be called exactly once per `Parser`; a second call sees `hasTokens() == false` and returns an empty program.
- **Performance/Scale Notes:** Single linear pass; cost is dominated by the expression Pratt loop inside statement productions rather than anything here. Appending to `Statements` grows a slice that is never pre-sized, which is immaterial at policy-file scale.
- **Dependencies Risk:**
  - **Two distinct "empty" results.** An empty file returns `(nil, nil)` — a **nil program with no error**. [[loader]] special-cases this by skipping, but any other caller that dereferences the result without a nil check will panic on an empty `.sentrie` file. A comment-only file, by contrast, returns a non-nil program with zero declarations.
  - **Fail-fast on the first errored statement.** The loop checks `p.err` after every statement and returns immediately, so even though [[parser.parser]] joins errors, in practice a program-level parse surfaces only the diagnostics accumulated up to the first failing statement — not every error in the file.
  - **Comments occupy statement slots.** `CommentStatement` entries are interleaved with real declarations in `Statements`, and trailing comments are lifted into their own statements. [[index]] and any other walker must filter them explicitly.
  - **Semicolons are optional and silently swallowed** between statements; their absence is never an error, so a missing separator produces a confusing downstream diagnostic instead of a "missing `;`" message.
  - **Context is decorative.** Cancellation during a large parse is not honoured here; only [[loader]]'s per-file walk observes it.
