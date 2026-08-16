---
id: index.rule
type: Class
language: Go
file_path: index/rule.go
tags: rule-model, evaluation-unit, dependency-graph
---

# Node: index.Rule (Rule Record)

## 1. Architectural Role & Intent
`Rule` is the semantic record of one `rule name = [default …] [when …] body` declaration — the atomic unit of evaluation in Sentrie. It keeps the three expression slots independent so [[runtime]] can treat the guard, the body, and the fallback as separate decisions, and it doubles as a graph vertex: rules are the node type of the rule DAG built in [[index.validate]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.rule` | `DEPENDS_ON` | [[ast]] | Holds `ast.Expression` references for `Default`, `When`, and `Body` without copying or evaluating them; builds the FQN with `ast.CreateFQN`. |
| `index.rule` | `DEPENDS_ON` | [[tokens]] | `Span()` exposes the declaration range for diagnostics. |
| [[index.policy]] | `CALLS` | [[index.rule]] | `AddRule` calls `createRule` during policy ingestion. |
| [[index.validate]] | `READS_FROM` | [[index.rule]] | Uses rules as DAG vertices and walks all three expression slots for identifier-cycle analysis. |
| [[index.builtin_check]] | `READS_FROM` | [[index.rule]] | Kind-checks builtin calls in `Default`, `When`, and `Body`. |
| [[dag]] | `DEPENDS_ON` | [[index.rule]] | `*Rule` satisfies the graph's `String()` node constraint. |
| [[runtime]] | `READS_FROM` | [[index.rule]] | Evaluates guard, body, and default. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Rule` struct — `{ Node *ast.RuleStatement, Policy *Policy, Name string, FQN ast.FQN, Default, When, Body ast.Expression }`
  - **Behavior:** Note the AST field is called `Node`, unlike `Shape`, `Policy`, and `Derive`, which all call theirs `Statement`. `Default` and `When` are nil when the clauses are absent — nil is the encoding for "no fallback" and "no guard". `Body` may be a block, a plain expression, or an `ast.ImportClause` when the rule delegates to another policy's decision.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `createRule(p *Policy, stmt *ast.RuleStatement) -> (*Rule, error)`
  - **Behavior:** Pure field copy with the FQN derived from the policy. It returns an `error` for call-site symmetry but **can never fail**.
  - **Side Effects:** Allocation only.
  - **Exceptions:** None.

- **Signature:** `(*Rule).String() -> string`
  - **Behavior:** Renders the FQN. Used verbatim in cycle-detection messages, so rule cycles read as arrows between fully-qualified rule paths.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Rule).Span() -> tokens.Range`
  - **Behavior:** Delegates to `Node`.
  - **Side Effects:** None.
  - **Exceptions:** Panics if `Node` is nil — always populated in practice.

## 4. Operational Context & Gotchas
- **Statefulness:** Immutable after construction.
- **Performance/Scale Notes:** A pure value record; all cost lives in the passes that consume it.
- **Dependencies Risk:**
  - **An `ImportClause` body is a cross-policy edge.** When `Body` is an import clause the rule's real dependency lives in another policy, discovered only by `detectRuleCycle`, which resolves the target through [[index.resolve]]. A miss there surfaces as a validation error, not an ingestion error.
  - **Rule cycles are detected on two different graphs.** `detectReferenceCycle` builds a per-policy graph over **bare identifiers** (catching `a = b`, `b = a` inside one policy), while `detectRuleCycle` builds a global graph over **import edges** only. A cycle that alternates between the two mechanisms is not guaranteed to be caught by either.
  - **Nothing type-checks the slots.** No constraint requires `When` to be trinary or `Default` to match `Body`'s type; that is a runtime concern.
  - **`createRule`'s error return is vestigial**, so callers that carefully handle it are guarding against a case that cannot occur.
