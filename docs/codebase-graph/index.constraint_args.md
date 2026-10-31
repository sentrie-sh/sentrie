---
id: index.constraint_args
type: Function / Endpoint
language: Go
file_path: index/constraint_args.go
tags: static-analysis, constraints, type-refinement, diagnostics
---

# Node: index.checkConstraintArgs (Static Constraint Argument Check)

## 1. Architectural Role & Intent
Rejects, at index time, a type constraint whose own arguments are broken: an invalid regexp, a wrong argument kind, a zero divisor, a range whose minimum is above its maximum (#273). Before this check such a constraint was only reported when a value of the constraint's own kind reached it, so `x is not string@regexp("[")` was `true` for a number and `sentrie validate` passed. The check is each definition's own `CheckArgs` from [[constraints]], the same one the runtime makes before it looks at a value ([[runtime.typeref_constraints]]), so the two cannot disagree.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.constraint_args` | `DEPENDS_ON` | [[constraints]] | `ForTypeRef` selects the table for a type ref; `ConstraintDefinition.CheckArgs` checks the folded arguments. |
| `index.constraint_args` | `READS_FROM` | [[index.policy]] | Facts and policy lets carry type refs. |
| `index.constraint_args` | `READS_FROM` | [[index.shape]] | Each `ShapeStatement`'s alias type and field types, once per statement. |
| `index.constraint_args` | `READS_FROM` | [[index.derive]] | A derive lambda's parameter and return types. |
| `index.constraint_args` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` (`builtin_calls.go`) visits every expression: block lets, lambda types, casts and `is` types. |
| `index.constraint_args` | `DEPENDS_ON` | [[xerr]] | Emits `ConstraintArgError` (wraps `ErrIndex` and the argument failure). |
| [[index.validate]] | `CALLS` | [[index.constraint_args]] | Runs last, after `checkBuiltinCalls`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkConstraintArgs(ctx) -> error`
  - **Behavior:** Visits every type ref position (facts, policy lets, shape aliases and fields, derive and lambda parameter and return types, block lets, casts, `x is T`), unwraps a nullable, and recurses into list element, dict value and record field types. For each constraint whose name is in the table and whose arguments are all literals, folds the arguments to `box` values as the runtime evaluates them and runs `CheckArgs`. Collects every failure, sorted by message, and returns them joined.
  - **Side Effects:** None.
  - **Exceptions:** Joined `xerr.ConstraintArgError`, each at the constraint's span with the message `invalid argument to constraint '<name>': <argument failure>`; it matches `xerr.ErrIndex` and `constraints.ErrInvalidArgument` under `errors.Is`. `validation cancelled` on context cancellation.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One pass over the policy set plus one expression walk; a regexp argument is compiled once per constraint occurrence.
- **Dependencies Risk:**
  - **Unknown names are skipped, not reported.** The parser already rejects an unknown name or a wrong count through [[ast]]'s generated tables; whether the runtime implements a name is #113.
  - **Only literal arguments are checked.** The parser accepts only literals (`parseConstraintLiteral` in [[parser.literal]]), so a non-literal cannot reach here today; if one ever does, it is skipped rather than guessed at.
  - **A new type ref kind needs a case in two places.** `constraints.ForTypeRef` for its table, and the recursion here for any child type refs.
