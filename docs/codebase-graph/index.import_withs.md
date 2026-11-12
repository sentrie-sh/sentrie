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
| `index.import_withs` | `READS_FROM` | [[index.policy]] | Matches each `with` name against the target's declared fact names (`FactByName`); `FactAliasesByName()` supplies the name to alias map for the message. |
| `index.import_withs` | `CALLS` | [[index.resolve]] | `ResolvePolicy` finds the imported policy (already known to exist after the rule-cycle pass). |
| `index.import_withs` | `DEPENDS_ON` | [[xerr]] | Emits `ImportUnknownFactError` (wraps `ErrIndex`); `ImportUnknownFactMessage` is shared with the runtime fallback. |
| [[index.validate]] | `CALLS` | [[index.import_withs]] | Runs right after `detectRuleCycle`. |
| [[runtime.imports]] | `DEPENDS_ON` | [[index.import_withs]] | Uses the same key space (declared fact names) and the same message text. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkImportWiths(ctx) -> error`
  - **Behavior:** For every rule body that is an `ast.ImportClause` with `with` clauses, resolves the target policy through `importTarget` (`validate.go`, shared with `detectRuleCycle`; a single segment means the importer's namespace) and reports each `with` name that is not a declared fact name of the target. Collects all misses, sorted by file, line, column, message, and returns them joined.
  - **Side Effects:** None.
  - **Exceptions:** Joined `xerr.ImportUnknownFactError`; `error resolving policy` (cannot occur after `detectRuleCycle`); `validation cancelled`.

- **Signature:** `(*Policy).FactAliasesByName() -> map[string]string` (in `index/policy.go`)
  - **Behavior:** Each fact's declared name (the key a `with` clause uses) mapped to its alias.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Key space is the declared fact names** ([#147](https://github.com/sentrie-sh/sentrie/issues/147)), not the `Policy.Facts` keys (aliases). `ImportDecision` and `ExecRule` use the same names, so the static and runtime checks agree by construction. The message lists the accepted names and, when the `with` name is a fact's alias, points to that fact's name.
- **Imports are only parsed as a rule body** (`parser/rule.go`), the same assumption `detectRuleCycle` makes. If imports ever become legal in nested expressions, this check and `detectRuleCycle` must both learn to walk them.
- **The reported span ends at the end of the `with` expression**, and for some expressions (a map literal) that end position is reported as column 0 by the parser; the start is right.
