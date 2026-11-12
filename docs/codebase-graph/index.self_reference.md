---
id: index.self_reference
type: Function / Endpoint
language: Go
file_path: index/self_reference.go
tags: static-analysis, recursion, name-resolution, diagnostics
---

# Node: index.detectSelfReference (Scoped Self-Reference Check)

## 1. Architectural Role & Intent
Reports a rule or let whose own value reads itself, such as `rule r = r` or `let x = x + 1` ([#155](https://github.com/sentrie-sh/sentrie/issues/155)). The identifier graph in [[index.validate]] cannot: it ignores scope, so its self edges are dropped (`dag.ErrSelfLoop`), since a block let, lambda parameter or fact sharing the rule's name would read as a self-reference. This check resolves each identifier in the order the runtime's `evalIdent` uses ([[runtime.eval_ident]]) and reports only an identifier that resolves to the rule or let it is written in. Before it, such a pack passed `sentrie validate` and failed on its first evaluation with `'t/p/r' references itself`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.self_reference` | `CALLS` | [[index.expr_walk]] | `forEachExprChild` for every node kind the walk has no case of its own for. |
| `index.self_reference` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` feeds `selfRefBinders`, the per-policy block lets and lambda parameters. |
| `index.self_reference` | `CALLS` | [[index.is_resolve]] | `joinSiteErrors` orders the report. |
| `index.self_reference` | `READS_FROM` | [[index.policy]] | `Rules`, `Lets`, `Facts`, `Data`, `RuleExports` and each rule's contributions. |
| `index.self_reference` | `DEPENDS_ON` | [[xerr]] | Errors wrap `ErrInfiniteRecursion` and `ErrIndex`. |
| [[index.validate]] | `CALLS` | [[index.self_reference]] | Runs right after `detectReferenceCycle`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).detectSelfReference(ctx) -> error`
  - **Behavior:** Per policy, walks each rule's default, `when` and body with the rule as owner, each policy let's value with the let as owner (none when the value is a lambda literal), each export attachment with the exported rule as owner (`execRule` evaluates attachments with the rule still on the reference stack, so `export decision of r attach note as r` fails at exec), and fact defaults and contribution slots with no owner, so only their block lets can be reported. A block binds all its lets before walking their values (as `evalBlock` injects them), and each block let's value is walked with that let as owner. An identifier resolves to: a parameter of any enclosing lambda (the runtime reads locals first, across the whole chain), then pack data, then a fact, then the nearest let (block lets inside out, then the policy's), then a rule. It is reported when that is its owner, unless the owner is a let whose name is bound anywhere else in the policy (`boundElsewhere`: another block let or policy let, a lambda parameter, a rule, a fact or pack data). A pre-pass over `forEachScopedExpr` ([[index.builtin_check]]) (`selfRefBinders`) collects every block let and lambda parameter per policy, derive bodies excluded.
  - **Side Effects:** None.
  - **Exceptions:** Joined, in source order, one per self-reading identifier at its span: `rule '<fqn>' references itself at <span>: infinite recursion: <fqn> -> <fqn>: index error`, or `let '<name>' references itself at ...`. `validation cancelled` on context cancellation; the walker's unsupported-node error.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; scopes are frames built per walk.
- **Fails safe; it may miss, never invents.**
  - **A lambda body has no owner.** A let bound to a lambda may call itself (#161) and a lambda passed to a call may never run, so `rule r = filter(xs, (x) => { yield r })` is not reported and still fails at runtime.
  - **Callees are skipped.** `getTarget` resolves a callee without evaluating it; calling a rule is [[index.call_check]]'s.
  - **A let self-reference whose name is bound elsewhere in the policy is skipped.** `evalIdent` reads cached locals across every ancestor context before it looks a let up, an evaluated let or rule is cached as a local, and a same-policy rule or policy let is evaluated in the context that reads it (`execRule` gets the caller's `ec`; cross-policy `execRuleFrom` starts fresh). So the let's own read may return another binding's value: with `let x = 1`, `rule r = (x + { let x = x + 1 yield x }) == 2` is `true`, and `rule r2 = { let x = x + 1 yield x }` called from `rule r = { let x = 1 yield x + r2 } == 2` (or from inside `(x) => { ... }`) reads the caller's `x`. Whether that happens depends on evaluation order and call paths, which the check does not model, so any other binding of the name anywhere in the policy (another block or policy let, a lambda parameter, a rule, a fact, pack data) skips the report (PR review on #155). Known gap: packs that still fail at runtime are skipped too, such as `rule r = { let x = 1 yield { let x = x + 1 yield x } }`, and `let x = 1` with `rule r = count(filter([1], (v) => { let x = x + v yield x == 2 })) == 1`. `let x = x + 1` with no other `x` in the policy is still reported.
  - **A name declared twice in one block** resolves to neither declaration (the runtime fails the block with a let conflict), so it is never reported here. Nothing else reports that conflict statically yet.
- **Contributions** that read their own rule are [[index.contribute]]'s (`detectContributionSelfReference`), with its own wording.
- **Derive bodies are not walked**; they cannot read rules or lets ([[index.derive_purity]]).
