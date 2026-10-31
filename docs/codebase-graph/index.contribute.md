---
id: index.contribute
type: Class
language: Go
file_path: index/contribute.go
tags: rule-model, incremental-rule, static-check
---

# Node: index.Contribution (Rule Contribution Record)

## 1. Architectural Role & Intent
`Contribution` is the semantic record of one `contribute` statement, attached to the [[index.rule]] it adds to ([#160](https://github.com/sentrie-sh/sentrie/issues/160)). The file also owns the static rules for contributions: which rules may take them, how each kind combines, and the literal-yield check that catches obvious type errors at index time rather than at evaluation.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.contribute` | `DEPENDS_ON` | [[parser.contribute]] | Holds the parsed `ast.ContributeStatement` and its guard and body. |
| `index.contribute` | `MUTATES` | [[index.rule]] | Appends to `Rule.Contributions` in declaration order. |
| `index.contribute` | `DEPENDS_ON` | [[ast]] | Reads the target rule's declared `TypeRef` and literal node types. |
| `index.contribute` | `DEPENDS_ON` | [[xerr]] | Every rejection wraps `xerr.ErrIndex`. |
| [[index.policy]] | `CALLS` | [[index.contribute]] | `addContribution` runs for each `ast.ContributeStatement` during policy ingestion. |
| [[index.validate]] | `CALLS` | [[index.contribute]] | `detectContributionSelfReference` runs after the reference-cycle check. |
| [[runtime.contribute]] | `READS_FROM` | [[index.contribute]] | Uses `ContributionKindOf` and each contribution's guard and body. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Contribution` struct - `{ Node *ast.ContributeStatement, Rule *Rule, When, Body ast.Expression }`
  - **Behavior:** `When` is nil when the contribution has no guard. `Rule` points back at the target.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `ContributionKindOf(typeRef ast.TypeRef) -> (ContributionKind, ast.TypeRef)`
  - **Behavior:** After unwrapping a nullable type: `list[T]` is `ContributionList` with yield type `T`; `boolean` or `trinary` (both a `TrinaryTypeRef`) is `ContributionBool` with the type itself; nil or anything else is `ContributionNone`.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Policy).addContribution(stmt *ast.ContributeStatement) -> error`
  - **Behavior:** Attaches the statement to `p.Rules[stmt.RuleName]` after checking, in order: the rule is already declared in this policy (a contribution must follow its rule, as `export` must); the rule declares a type of kind list or boolean; the rule's body is not an import clause; and the body is not a literal, or a block yielding a literal, whose kind cannot satisfy the yield type. A `record[...]` yield type accepts a list literal, since records are fixed-length lists at runtime.
  - **Side Effects:** Appends to the rule's `Contributions`.
  - **Exceptions:** `cannot contribute to unknown rule`, `cannot contribute to rule ... the rule declares ...`, `... the rule imports its decision from another policy`, `contribution to rule ... yields a <kind>, but the rule takes <type>`; all wrap `xerr.ErrIndex`.

- **Signature:** `(*Index).detectContributionSelfReference(ctx) -> error`
  - **Behavior:** Rejects a contribution whose guard or body names its own rule, walking with the scope-aware walker so a lambda parameter or block `let` of the same name shadows the rule. The reference-cycle graph drops self edges, so this is the only static check for that loop. Longer loops (`contribute deny when seen ...` with `rule seen = count(deny)`) are caught by the reference-cycle graph.
  - **Side Effects:** None.
  - **Exceptions:** `contribution to rule '<name>' at <span> reads the rule itself at <span>`, wrapping `xerr.ErrIndex`; cancellation.

## 4. Operational Context & Gotchas
- **Statefulness:** Contributions are immutable once the policy is built.
- **Performance/Scale Notes:** Constant work per statement.
- **Dependencies Risk:**
  - **The literal check is deliberately partial.** Only literal yields are checked statically; any computed value is checked against the declared type at runtime by [[runtime.contribute]]. Document and shape element types are never checked here.
  - **Contributions feed the other index passes.** [[index.validate]] adds each contribution's guard and body to the target rule's reference-cycle edges, and [[index.builtin_check]] walks them, so a contribution is held to the same static checks as the rule body.
