---
id: parser.contribute
type: Function / Endpoint
language: Go
file_path: parser/contribute.go
tags: declaration, rule, incremental-rule, policy-semantics
---

# Node: parser.parseContributeStatement (Rule Contribution)

## 1. Architectural Role & Intent
Parses `contribute <ident> ['when' expr] expr`, one incremental addition to a rule declared with `rule` in the same policy ([#160](https://github.com/sentrie-sh/sentrie/issues/160)). It lets a result be built up from independent conditions instead of one rule enumerating them all. The parser only records the target name, the optional guard, and the body; whether the target exists and can take contributions is decided by [[index.contribute]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.contribute` | `CALLS` | [[parser.expression]] | Parses the `when` guard and the body at `LOWEST` precedence. |
| `parser.contribute` | `CALLS` | [[parser.parser]] | Uses `advanceExpected` and `canExpect`. |
| `parser.contribute` | `CALLS` | [[ast]] | Emits `ast.NewContributeStatement(name, when, body, span)`. |
| [[parser.policy]] | `CALLS` | [[parser.contribute]] | Registered for `tokens.KeywordContribute` in the policy-scope table only. |
| [[index.contribute]] | `DEPENDS_ON` | [[parser.contribute]] | Attaches the parsed statement to its rule. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseContributeStatement(ctx: context.Context, parser: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `contribute` and the target name, then an optional `when <expr>`, then the body. There is no `=`, no `default`, and no import clause: a contribution that does not apply is skipped rather than falling back. `when` is optional, so `contribute deny { yield "x" }` contributes unconditionally (still skipped when the body yields undefined). The span runs from `contribute` to the end of the body.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on a missing name or when the guard or body fails to parse.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Dominated by the nested expression parses.
- **Dependencies Risk:**
  - **`contribute` is a reserved word.** Like every keyword it is lexed as a keyword token everywhere, so it can no longer be an identifier, a field name after `.`, or a namespace path segment.
  - **The body is any expression**, as for `rule`; the grammar documents a block. A guard followed by a list literal (`when x [ ... ]`) parses as an index access, exactly as it does for `rule`, so write the body as a block.
  - **Policy scope only.** At namespace scope `contribute` is an unexpected token, and blocks only accept `let` and comments, so it cannot appear inside a block either.
