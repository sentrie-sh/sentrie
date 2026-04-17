---
name: Pipeline hole syntax
overview: Add `#` as a pipeline placeholder so the piped value can fill a non-first parameter (e.g. string replace). When `#` is absent, keep today’s behavior (prepend as first argument). Discourage nested inner-call holes in favor of chaining. Lexer + AST + parser substitution, tests, docs; reject stray `#` at eval.
todos:
  - id: lex-token
    content: Add TokenPipelineHole + lexer `case '#'` + lexer tests
    status: completed
  - id: ast-hole
    content: Add PipelineHoleExpression AST + interface test hookup
    status: completed
  - id: parse-prefix
    content: Register prefix parser for hole token
    status: completed
  - id: pipeline-subst
    content: Implement hole detection/substitution in parsePipelineExpression (CallExpression branch); recurse within each top-level argument subtree where needed; preserve memoization
    status: completed
  - id: eval-guard
    content: "eval.go: error on stray PipelineHoleExpression"
    status: completed
  - id: tests-cover
    content: Extend pipeline tests + run coverage on lexer/parser
    status: completed
  - id: docs-website
    content: Update function-chaining.md (and PR description files)
    status: completed
isProject: false
---

# Pipeline placeholder (`#`) for function chaining

## Motivation (primary)

Pipelines already read well when the piped value is the **first** argument. The hole exists for APIs where the “natural” pipeline value belongs **elsewhere**—for example the needle in a replace call:

```sentrie
let replaceChar = "..."
replaceChar |> str.replace(input, #, "$$")
```

That lowers to `str.replace(input, replaceChar, "$$")`. Without `#`, today’s desugaring would only allow prepending (`str.replace(replaceChar, input, ...)`), which is the wrong argument order.

**Style:** Prefer **chaining** when the data flow is “apply g, then f” (`x |> g |> f`). Patterns like `x |> f(g(#))` are **discouraged**: they are usually *less* readable than pushing the inner step into its own `|>` step. The implementation may still substitute `#` inside nested expression trees for correctness (or for rare cases), but docs and examples should **not** present nested inner calls as idiomatic.

## Semantics (agreed)

- **RHS is a call** on an identifier or module-qualified field access (unchanged from [parser/pipeline.go](parser/pipeline.go)).
- **No `#` in the argument list**: same as today — prepend the left-hand expression as the **first** argument (`x |> f()` → `f(x)`; `x |> f(1)` → `f(x, 1)`).
- **One or more `#` in the argument list**: **do not** prepend. Recursively replace every `#` with the piped value within the RHS call’s **argument** subtrees (so a hole can appear in any argument position, including inside a top-level ternary/list if it ever appears there). **Hero example:** `needle |> str.replace(haystack, #, repl)` → `str.replace(haystack, needle, repl)`.
- **Placeholder token**: **`#`** as a new dedicated token (avoids overloading `%` / modulo). Document clearly; modulo stays `a % b` unchanged.

## Implementation (sentrie repo)

### 1. Lexer and tokens

- Add `TokenPipelineHole` (name TBD; e.g. `PipelineHole`) in [tokens/token_kind.go](tokens/token_kind.go).
- In [lexer/lexer.go](lexer/lexer.go), add `case '#':` that emits the new token (single rune, same pattern as `@`).
- Tests in [lexer/lexer_test.go](lexer/lexer_test.go): `#` alone, `#` inside `f(a, #, b)`, commas, adjacent tokens.

### 2. AST

- New node `PipelineHoleExpression` (new file under `ast/`, e.g. `ast/pipeline_hole.go`) implementing `Expression`: `Span`, `String()` as `#`, `Kind()`, `expressionNode()`.
- Extend [ast/node_test.go](ast/node_test.go) (or equivalent) so the expression interface coverage includes the new type.

### 3. Parser

- Register **prefix** parser for `TokenPipelineHole` in [parser/lookups.go](parser/lookups.go) (e.g. `parsePipelineHoleExpression`).
- **Pipeline lowering** in [parser/pipeline.go](parser/pipeline.go) for `*ast.CallExpression`:
  - Detect whether any argument subtree contains `*ast.PipelineHoleExpression`.
  - If **yes**: build new arguments with `substitutePipelineHoles(expr, left)`; **do not** prepend `left`.
  - If **no**: keep current `append(left, rhs.Arguments...)`.
  - Re-apply `applyPipelineMemoizationSuffix` on the final call as today.
- Implement `substitutePipelineHoles` with an exhaustive `switch` over expression types that can embed sub-expressions (calls, lists, maps, ternary, unary, infix, field/index access, cast, etc.). Do **not** substitute into the **callee** of the pipeline RHS call. Nested `f(g(#))` remains *technically* substitutable but is **not** a recommended pattern in documentation.

### 4. Stray `#` outside pipelines

- Pipelines eliminate holes during lowering; if a hole survives (e.g. `let x = #`), evaluation should fail clearly.
- Add `case *ast.PipelineHoleExpression` in [runtime/eval.go](runtime/eval.go) returning a descriptive error.

### 5. Tests

- [parser/pipeline_test.go](parser/pipeline_test.go): focus table on **non-first** placement and real-world-shaped callees, e.g.:
  - `x |> f(1, #)` → `f(1, x)`; `x |> f(#, 2)` → `f(x, 2)`.
  - **Motivating:** `needle |> str.replace(haystack, #, "$$")` (or equivalent module-qualified names used in tests) → `str.replace(haystack, needle, "$$")`.
  - `x |> f(#)` ≡ `f(x)` (explicit first slot; same as prepend but with `#`).
  - Mixed chain: `x |> f(1, #) |> g(#)` → `g(f(1, x))`.
  - Multiple holes in **flat** args: `x |> f(#, #)` → `f(x, x)` if we support it; document both refer to the piped value.
  - Memoization suffix still applies after lowering.
- **Do not** center tests on `x |> f(g(#))`; at most one edge-case test if recursion is required for coverage, with a comment that style prefers `x |> g |> f`.
- Negative: `#` in invalid pipeline RHS still rejected by existing pipeline target rules.
- Optional runtime test: expression containing only `#` errors in eval.

### 6. Coverage

- Keep **≥90%** coverage on new/changed lexer/parser paths (project norm from issue #76 work); run `go test` with cover for `lexer` and `parser` packages.

### 7. Documentation (website repo)

- Update [website/src/content/docs/reference/function-chaining.md](website/src/content/docs/reference/function-chaining.md):
  - Lead with **non-first argument** motivation and a `str.replace`-style example.
  - Explain “omit `#` → value becomes first argument” (current behavior).
  - Add a short **Multiple holes** note with one example (for completeness): `x |> f(#, #)` → `f(x, x)`, and explicitly say all `#` placeholders bind to the same piped value.
  - Add a dedicated **Style** subsection (heading-level, not a single bullet) that states clearly:
    - Prefer **straight chaining** when the data flow is “apply one function, then another”: e.g. `x |> g |> f` rather than folding the same logic into `x |> f(g(#))`.
    - Use **`#`** when the piped value must appear in a **non-first** parameter (the motivating replace-style case); that is what holes are for, not nesting inner calls for readability.
    - Nested inner-call holes may be supported by the parser for edge cases, but treat **`x |> outer(inner(#))`** as a **discouraged** pattern when **`x |> inner |> outer`** expresses the same flow more clearly.
  - Note modulo `%` is unrelated to `#`.
- [website/src/content/docs/reference/precedence.md](website/src/content/docs/reference/precedence.md): optional one-line mention only if useful.

### 8. PR housekeeping

- **sentrie**: update branch `PR_DESCRIPTION_<branch>.md` per [AGENTS.md](AGENTS.md) (issue number / scope).
- **website**: update local `PR_DESCRIPTION.md` (even if gitignored, per website AGENTS.md).

## Non-goals (keep scope tight)

- No `%` spelling in v1 unless explicitly expanded later.
- No changes to ordinary call parsing beyond what’s required for `#` as a primary expression.
- No runtime semantic for `#` other than parser lowering + eval error if it leaks.
- Documentation should not promote nested inner-call holes as idiomatic, even if the parser allows them.
