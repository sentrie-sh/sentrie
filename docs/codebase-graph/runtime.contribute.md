---
id: runtime.contribute
type: Module / File
language: Go
file_path: runtime/contribute.go
tags: rule-evaluation, incremental-rule, determinism, type-validation
---

# Node: runtime/contribute.go (Rule Contributions and Declared Rule Types)

## 1. Architectural Role & Intent
Combines a rule's own outcome with its `contribute` statements ([#160](https://github.com/sentrie-sh/sentrie/issues/160)) and checks rule values against a declared rule type. Contributions are evaluated one after another in declaration order inside the rule's own evaluation, so the combined result never depends on scheduling ([#143](https://github.com/sentrie-sh/sentrie/issues/143)). Because the combination happens inside rule evaluation, every reader of the rule - another rule's reference, an export, an attachment, a cross-policy import - sees the combined value.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.contribute` | `READS_FROM` | [[index.contribute]] | `ContributionKindOf` and each contribution's guard and body. |
| `runtime.contribute` | `CALLS` | [[runtime.eval]] | Evaluates each guard and body. |
| `runtime.contribute` | `CALLS` | [[runtime.typeref]] | `validateValueAgainstTypeRef` for each contributed value, the combined value, and rule body or default values. |
| `runtime.contribute` | `CALLS` | [[runtime.decision]] | Wraps the combined value via `DecisionOf`; unwraps an imported decision with `reported` before validating. |
| `runtime.contribute` | `CALLS` | [[runtime.trace]] | Opens `rule-contribution`, `contribution-when`, and `contribution-body` nodes. |
| [[runtime.executor]] | `CALLS` | [[runtime.contribute]] | `evaluateRuleOutcome` calls `applyContributions`; `evaluateRuleBase` calls `validateRuleValue`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateRuleValue(ctx, ec, e, p, r, val, at) -> error`
  - **Behavior:** Checks a value the rule's body or default produced against `r.Node.Type`. A rule with no declared type, and a value that is undefined (after unwrapping an imported decision to the value it reports), are not checked. A guard that does not hold with no default produces no value and is never checked.
  - **Side Effects:** May evaluate constraint arguments.
  - **Exceptions:** Type-mismatch and constraint errors from [[runtime.typeref]].

- **Signature:** `applyContributions(ctx, ec, e, p, r, base *Decision, rn *trace.Node) -> (*Decision, error)`
  - **Behavior:** For a `list[T]` rule the result is the base list followed by each applied contribution's value; a base that is not a list starts the list empty: undefined, a guard that did not hold with no default, and **a `null` base (for a nullable `list[T]?` rule) all become `[]`**, so a contributed nullable list rule never reports `null`. For a `boolean` or `trinary` rule the result is the Kleene OR of the base state and each applied contribution. Every contribution is evaluated (no short-circuit), each applied value is checked against the yield type, and the combined value is checked against the full declared type, so list constraints apply to the combined list.
  - **Side Effects:** Trace nodes under `rn`.
  - **Exceptions:** The first guard, body, or type error, returned with a nil decision so `ExecRule` reports `Unknown`, never the base or a partly combined value. `evaluateRuleBase` does the same for a declared-type mismatch on the body or default.

- **Signature:** `evalContribution(ctx, ec, e, p, i, c, rn) -> (box.Value, bool, error)`
  - **Behavior:** Reports `applied == false` when the guard is not true (`False` or `Unknown`) or the body yields undefined, recording `skipped: "when"` or `skipped: "undefined"` in the `rule-contribution` node's meta beside its declaration `index`.
  - **Side Effects:** Trace nodes.
  - **Exceptions:** Guard or body evaluation errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; all state is in the execution context.
- **Performance/Scale Notes:** One evaluation per contribution per rule evaluation; the rule's result is not memoised beyond what the execution context already caches.
- **Dependencies Risk:**
  - **A boolean rule's skipped base is `Unknown`, not `False`.** Kleene OR keeps `Unknown OR False` as `Unknown`, so a boolean rule that should default to false says so (`rule allow: boolean = false`).
  - **An attachment cannot read its own exported rule.** Attachments run while the rule is still on the recursion stack, so `attach n as count(deny)` on `export decision of deny` is reported as self-recursion; this predates contributions.
