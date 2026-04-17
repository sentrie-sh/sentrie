---
name: Issue 76 Pipeline Operator Plan
overview: Implement `|>` as parser-level desugaring into ordinary call expressions, with exhaustive lexer/parser coverage and semantic regression checks so `use` behavior remains unchanged.
todos:
  - id: token-lexer
    content: Add TokenPipeForward and lexer handling for `|>` plus bare `|` errors
    status: completed
  - id: parser-plumbing
    content: Add pipeline precedence and infix handler registration
    status: completed
  - id: pipeline-parser
    content: Implement parsePipelineExpression with desugaring, validation, metadata preservation, and spans
    status: completed
  - id: pipeline-tests
    content: Add exhaustive lexer/parser positive and negative tests, including precedence and memoization cases
    status: completed
  - id: use-semantics-regression
    content: Add semantic regression test proving `use` behavior is unchanged for bare imported names in pipelines
    status: completed
  - id: coverage-gate
    content: Run package coverage and verify >=90% coverage for all new code
    status: completed
  - id: docs-exhaustive
    content: Update website documentation exhaustively for pipeline syntax, semantics, boundaries, and migration guidance
    status: completed
  - id: pr-description
    content: Update PR_DESCRIPTION_<branch_name>.md with issue-76-only summary and testing notes
    status: completed
  - id: pipeline-memoization-extension
    content: Extend pipeline parsing/desugaring to support memoization on identifier/field-access RHS targets and add matching tests
    status: completed
  - id: memoization-suffix-refactor
    content: Add shared parseMemoizationSuffix helper and wire it into parser/call.go and parser/pipeline.go
    status: completed
isProject: false
---

# Issue 76: Pipeline Operator (`|>`) Implementation Plan

## Scope and constraints
- Add `|>` as syntax sugar only; runtime evaluation remains unchanged.
- Keep `use` alias semantics unchanged (no local symbol injection).
- Reject bare `|` at lexing time.
- Do not modify ordinary call parsing in [parser/call.go](parser/call.go).
- Support memoization for pipeline targets consistently, including identifier/field-access targets (not only pre-existing call-expression RHS forms).
- Require at least 90% coverage for all newly added lexer/parser code paths.

## Implementation steps
### 1) Token and lexer support
- Add `TokenPipeForward` to [tokens/token_kind.go](tokens/token_kind.go).
- Update [lexer/lexer.go](lexer/lexer.go) to:
  - emit `TokenPipeForward` for `|>`
  - emit a lexer `Error` token for bare `|` (including EOF/trailing forms).

### 2) Pratt precedence and infix registration
- Add a dedicated pipeline precedence in [parser/precedence.go](parser/precedence.go) just above `LOWEST` and below ternary/logical/arithmetic levels.
- Map `TokenPipeForward` to that precedence.
- Register a new infix handler in [parser/lookups.go](parser/lookups.go): `TokenPipeForward -> parsePipelineExpression`.

### 3) Pipeline desugaring parser
- Add [parser/pipeline.go](parser/pipeline.go) with `parsePipelineExpression` that:
  - consumes `|>`
  - parses RHS using pipeline precedence
  - validates RHS forms:
    - `*ast.Identifier`
    - `*ast.FieldAccessExpression`
    - `*ast.CallExpression` whose callee is identifier or field-access
  - lowers to `ast.NewCallExpression` by prepending LHS as argument 0
  - preserves `Memoized` and `MemoizeTTL` when RHS is already a call
  - supports memoization syntax on non-call RHS targets as well (e.g., `lhs |> ident!` and `lhs |> alias.fn!30`) by constructing memoized `ast.CallExpression` nodes with matching TTL semantics
  - sets a stable combined span from LHS start through RHS end
  - emits a clear parser error for invalid RHS shapes.
- Do not modify [parser/call.go](parser/call.go) call parsing behavior.

### 3.1) Shared memoization suffix parser refactor
- Add a shared parser helper (for example in a dedicated parser utility file) named `parseMemoizationSuffix`.
- Helper contract:
  - returns `nil` when no `!` suffix is present
  - returns a struct payload when `!` is present, containing optional `*time.Duration` TTL (`nil` TTL means default memoization behavior)
- Use this helper in both:
  - [parser/call.go](parser/call.go) (replace inline `!`/TTL parsing)
  - [parser/pipeline.go](parser/pipeline.go) (replace duplicated pipeline memoization suffix parsing)
- Keep parsing semantics unchanged (`!` and optional integer TTL in seconds), while reducing duplication and keeping span updates correct in each caller.

### 4) Exhaustive tests (from test-writer + gap fixes)
- **Lexer tests** in [lexer/lexer_test.go](lexer/lexer_test.go):
  - `|>` tokenization with/without spaces (`a|>b`)
  - multiline pipeline tokenization
  - bare `|`, trailing `|`, and malformed variants (`||`, `|>>`) produce errors.
- **Parser lowering and precedence tests** across [parser/precedence_test.go](parser/precedence_test.go), [parser/expression_test.go](parser/expression_test.go), and [parser/error_test.go](parser/error_test.go):
  - `value |> len` -> `len(value)`
  - `value |> string.trim` -> `string.trim(value)`
  - `value |> string.replaceAll(" ", "-")` -> `string.replaceAll(value, " ", "-")`
  - chaining: `value |> len |> math.abs` -> `math.abs(len(value))`
  - mixed chaining: `value |> string.trim |> len` -> `len(string.trim(value))`
  - low precedence: `a + b |> len`, `a ? b : c |> len`
  - invalid RHS forms rejected with pipeline-specific errors where the RHS callable root is not an identifier or module-qualified field access (e.g., grouped, infix, ternary, list, map, index, or field-access expressions whose root is not an identifier).
- **Memoization metadata preservation** tests:
  - first confirm memoization call syntax against existing parser/runtime tests before adding new pipeline cases (so examples match current grammar exactly).
  - `x |> f!`
  - `x |> f!10`
  - `x |> mod.f!30`
  - `x |> f()!`
  - `x |> f()!10`
  - field-qualified call variants preserving TTL and memoized flags.
- **Semantic regression for `use` behavior**:
  - parser accepts `value |> trim` syntactically because `Identifier` is an allowed RHS form; this parse-time acceptance must not imply resolution-time success.
  - add policy/runtime-level regression test (existing suite in parser/index/runtime as appropriate) proving `use { trim } from @sentrie/string` does not make bare `trim(...)` resolvable unless already resolvable under current rules.

### 5) Coverage and validation gate
- Run targeted coverage for changed packages:
  - default path: `go test ./lexer ./parser -coverpkg=./lexer,./parser,./tokens -coverprofile=pipeline.cover.out`
  - if semantic regression lands in runtime tests, expand to include runtime package in the same run.
  - `go tool cover -func=pipeline.cover.out`
- Ensure newly introduced code paths (lexer pipe branch + `parsePipelineExpression` branches) are >=90% covered.

### 6) Exhaustive documentation update
- Update user-facing docs in the website repository comprehensively, including:
  - language-level syntax and desugaring rules (`lhs |> ident`, `lhs |> alias.fn`, call variants with prepended first arg)
  - operator precedence and associativity expectations, including mixed chain examples (`value |> string.trim |> len`)
  - validity boundaries for RHS callable targets (allowed and rejected forms with concrete examples)
  - explicit parse-time vs resolution-time behavior for identifiers (`value |> trim` may parse but still fail resolution if unresolved)
  - `use` interaction constraints (no implicit local symbol injection)
  - multiline formatting and no-whitespace examples (`a|>b`) for readability and lexer behavior expectations
  - a short troubleshooting section for common parser errors around invalid RHS shapes or bare `|`.
- Ensure all relevant docs pages are aligned (getting started, language/reference, and CLI docs where expression examples appear) so the feature is discoverable and consistent.

### 7) Branch PR description update
- Update branch PR description file in repo root per repo convention (`PR_DESCRIPTION_<branch_name>.md`) to summarize only issue #76 changes, testing, reviewer focus areas, and docs updates.

## Reviewer focus points
- Pipeline precedence placement and associativity correctness.
- RHS shape validation boundaries (no accidental broadening).
- Memoization metadata preservation during desugaring.
- No semantic drift in `use` name resolution.