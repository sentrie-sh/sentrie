---
id: index.import_withs
type: Function / Endpoint
language: Go
file_path: index/import_withs.go
tags: static-analysis, imports, facts, diagnostics
---

# Node: index.checkImportWiths (Unknown `with` Fact Check)

## 1. Architectural Role & Intent
Rejects, at index time, an `import decision ... with <name> as <expr>` clause whose `<name>` is not a fact of the imported policy (#151). Before this check a typo silently left the target fact unbound or at its default, and `sentrie validate` passed. The error carries the span of the `with` clause and the facts the target declares, sorted. [[runtime.imports]] repeats the check as a fallback so a pack built without the index still fails closed.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.import_withs` | `READS_FROM` | [[index.rule]] | Looks at each rule whose `Body` is an `ast.ImportClause`, the only place the parser produces one. |
| `index.import_withs` | `READS_FROM` | [[index.policy]] | Matches each `with` name against the target `Policy.Facts` keys; `FactNames()` lists them sorted. |
| `index.import_withs` | `CALLS` | [[index.resolve]] | `ResolvePolicy` finds the imported policy (already known to exist after the rule-cycle pass). |
| `index.import_withs` | `DEPENDS_ON` | [[xerr]] | Emits `ImportUnknownFactError` (wraps `ErrIndex`); `ImportUnknownFactMessage` is shared with the runtime fallback. |
| [[index.validate]] | `CALLS` | [[index.import_withs]] | Runs right after `detectRuleCycle`. |
| [[runtime.imports]] | `DEPENDS_ON` | [[index.import_withs]] | Uses the same key space (`Policy.Facts` keys) and the same message text. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkImportWiths(ctx) -> error`
  - **Behavior:** For every rule body that is an `ast.ImportClause` with `with` clauses, resolves the target policy (namespace from the path's parent, or the importer's namespace for a single segment) and reports each `with` name absent from `target.Facts`. Collects all misses, sorted by file, line, column, message, and returns them joined.
  - **Side Effects:** None.
  - **Exceptions:** Joined `xerr.ImportUnknownFactError`; `error resolving policy` (cannot occur after `detectRuleCycle`); `validation cancelled`.

- **Signature:** `(*Policy).FactNames() -> []string`
  - **Behavior:** The keys of `Policy.Facts`, sorted. These are the names a `with` clause must use.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Key space is `Policy.Facts` keys, which are fact aliases** (`AddFact` stores `p.Facts[fact.Alias]`). `ImportDecision` and `ExecRule` use the same keys, so the static and runtime checks agree by construction. Issue #147 may change what the key is; this check follows whatever `Policy.Facts` is keyed by, so it needs no edit then.
- **Imports are only parsed as a rule body** (`parser/rule.go`), the same assumption `detectRuleCycle` makes. If imports ever become legal in nested expressions, this check and `detectRuleCycle` must both learn to walk them.
- **The reported span ends at the end of the `with` expression**, and for some expressions (a map literal) that end position is reported as column 0 by the parser; the start is right.
