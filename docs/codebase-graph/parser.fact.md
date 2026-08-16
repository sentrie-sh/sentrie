---
id: parser.fact
type: Function / Endpoint
language: Go
file_path: parser/fact.go
tags: declaration, fact, input-contract, optionality, nullability
---

# Node: parser.parseFactStatement (Fact Declaration)

## 1. Architectural Role & Intent
Parses `fact <ident>['?'] : <type> ['as' <ident>] ['default' <expr>]`. Facts are a policy's **input contract** - the typed data a caller must supply - so this production is where Sentrie's deliberate split between *presence* optionality (`name?`, the fact may be absent) and *value* nullability (`: T?`, the value may be null) is established. It also carries explicit rejection messages for the retired `!` syntax, so migrating policies get a directive error rather than a parse failure.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.fact` | `CALLS` | [[parser.typeref]] | Parses the declared type, including constraint suffixes. |
| `parser.fact` | `CALLS` | [[parser.expression]] | Parses the `default` expression at `LOWEST`. |
| `parser.fact` | `CALLS` | [[parser.parser]] | Uses `expect`, `advanceExpected`, `canExpect`, `canExpectAnyOf`, `errorf`. |
| `parser.fact` | `CALLS` | [[ast]] | Emits `ast.NewFactStatement(name, type, alias, default, optional, span)`. |
| [[parser.policy]] | `CALLS` | [[parser.fact]] | Registered for `tokens.KeywordFact` in the policy-scope table only - facts are policy-scoped. |
| [[runtime.exec_ctx]] | `DEPENDS_ON` | [[parser.fact]] | Binds supplied input values against the declared facts at execution time. |
| [[constraints]] | `CALLS` | [[parser.fact]] | The declared type's constraints are checked against the supplied value at runtime. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseFactStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `fact` and the name. Requires the next token to be `:` or `?`; a `?` sets `Optional = true` and is then followed by the required `:`. Parses the type reference, then optionally `as <ident>` (which overrides the alias, defaulted to the fact's own name) and `default <expr>`. The span is extended after each clause so it always covers the full declaration.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing keyword/name/colon, a failed type reference, a failed default expression, or either of the two targeted migration errors: `legacy '!' fact syntax is no longer supported; use 'fact X: T' or 'fact X?: T?'` and `expected ':' or '?' after fact name at <range>`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Optionality and nullability are different axes and both are expressible.** `fact x?: T` means the caller may omit `x`; `fact x: T?` means `x` must be supplied but may be null; `fact x?: T?` means both. Conflating them is the most likely source of runtime "missing fact" versus "null value" confusion - and only the first is a presence failure.
  - **Alias defaults to the name, then is overwritten.** `Alias` is initialised to the fact's name and replaced only if `as` appears, so downstream code should read `Alias` (never `Name`) when resolving an external input key.
  - **The optional-marker span update is off.** After consuming `?`, the range end is set from `p.head()` - the token *after* the marker - so a fact declaration's span can extend slightly past the `?`. Harmless for messages, but do not treat spans here as byte-exact.
  - **`default` is parsed but not evaluated or type-checked here.** Whether the default satisfies the declared type (and its constraints) is decided later.
