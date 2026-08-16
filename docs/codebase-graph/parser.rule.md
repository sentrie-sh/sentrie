---
id: parser.rule
type: Function / Endpoint
language: Go
file_path: parser/rule.go
tags: declaration, rule, policy-semantics, decision
---

# Node: parser.parseRuleStatement (Rule Declaration)

## 1. Architectural Role & Intent
Parses `rule <ident> = ['default' expr] ['when' expr] (block | importClause)` — the central declaration of the language, since a rule is what ultimately produces a [[trinary]] decision. The production's job is to keep the three expression slots independent and optional, so the runtime can later distinguish "the guard said this rule does not apply" (yield the default) from "the body evaluated to Unknown".

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.rule` | `CALLS` | [[parser.expression]] | Parses `default`, `when`, and the body at `LOWEST` precedence. |
| `parser.rule` | `CALLS` | [[parser.import]] | Delegates to `parseImportExpression` when the body begins with `import`. |
| `parser.rule` | `CALLS` | [[parser.parser]] | Uses `advanceExpected`, `expect(TokenAssign)`, `canExpect`. |
| `parser.rule` | `CALLS` | [[ast]] | Emits `ast.NewRuleStatement(name, default, when, body, span)`. |
| [[parser.policy]] | `CALLS` | [[parser.rule]] | Registered for `tokens.KeywordRule` in the policy-scope table only. |
| [[index.rule]] | `DEPENDS_ON` | [[ast]] | Indexes the rule and its dependencies. |
| [[runtime.decision]] | `DEPENDS_ON` | [[ast]] | Evaluates guard, body, and default into the exported decision. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseRuleStatement(ctx: context.Context, parser: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `rule`, the name, and a **required** `=`. Then optionally consumes `default <expr>` and `when <expr>` — **in that fixed order**. The body is either an import clause (when the head is `import`) or any expression, which in practice is a `{ … yield … }` block but is not required to be. The span runs from `rule` to the end of the body.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` on a missing keyword, name, or `=`, or when any of the three expressions fails to parse.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Dominated by the nested expression parses. Nothing specific here.
- **Dependencies Risk:**
  - **Clause order is rigid and silently enforced.** `when` before `default` does not error clearly — the `default` check runs first, fails to match, then the `when` check consumes the guard, and the subsequent `default` keyword lands in body position where it produces a confusing "no prefix parse function found for 'default'" message.
  - **The body is any expression.** Writing `rule x = true` parses fine; there is no requirement of a block or a trinary-typed body at parse time. Type and shape enforcement is entirely [[index.validate]]'s and the runtime's responsibility.
  - **`default` and `when` are unconstrained expressions.** Nothing prevents a `default` that calls a function or references a fact; whether that is *legal* (purity, ordering) is decided in [[index.package]].
  - **All three slots may be nil in the AST.** `RuleStatement.Default`, `.When`, and `.Body` are interface fields; only the body is guaranteed non-nil by a successful parse.
