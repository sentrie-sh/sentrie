---
id: index.policy
type: Class
language: Go
file_path: index/policy.go
tags: policy-model, scoping, metadata, decision-export, phase-ordering
---

# Node: index.Policy (Policy Scope and Declaration Ordering)

## 1. Architectural Role & Intent
`Policy` is the semantic model of one `policy { … }` block: its metadata, facts (keyed by **alias**), module `use` bindings, lets, rules, shapes, derives, and its exported decisions. `createPolicy` is both constructor and validator, and its defining job is enforcing a **four-phase declaration order** — metadata, then facts, then uses, then body — that the grammar does not express. It also enforces the pack's central contract: a policy with no exported rule is rejected outright.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.policy` | `DEPENDS_ON` | [[ast]] | Type-switches over every policy-body statement; retains all nodes by reference. |
| `index.policy` | `DEPENDS_ON` | `ext.masterminds.semver` | `version` metadata is parsed into a concrete `*semver.Version`. |
| `index.policy` | `DEPENDS_ON` | [[xerr]] | Uses `ErrConflict`, `ErrPolicyMetadataContiguous`, `ErrPolicyEmptyTitle`, `ErrPolicyInvalidVersion`, `ErrPolicyEmptyTagKey`, `ErrPolicyFactAfterUse`, `ErrInvalidInvocation`, `ErrIndex`. |
| `index.policy` | `CALLS` | [[index.policy_stmt]] | `policyStmtKindOf` classifies statements (to skip comments); `buildTagsByKey` derives the tag index. |
| `index.policy` | `CALLS` | [[index.rule]] | `AddRule` builds `Rule` records via `createRule`. |
| `index.policy` | `CALLS` | [[index.shape]] | `AddShape` builds policy-local shapes via `createShape`. |
| `index.policy` | `CALLS` | [[index.derive]] | `addDerive` with the namespace snapshot overlaid by the policy's derives-so-far. |
| `index.policy` | `CALLS` | [[index.namespace]] | Registered via `addPolicy`, which enforces cross-kind name uniqueness. |
| [[index.index]] | `CALLS` | [[index.policy]] | `AddProgram` calls `createPolicy` with a clone of the namespace's visible derives. |
| [[index.resolve]] | `READS_FROM` | [[index.policy]] | `VerifyRuleExported` is declared on this type. |
| [[index.builtin_kind]] | `READS_FROM` | [[index.policy]] | Seeds the static kind-check scope from `Facts`, `Lets`, and `Derives`. |
| [[runtime]] | `READS_FROM` | [[index.policy]] | Resolves facts by alias, lets, rules, uses, and exported decisions. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Policy` struct — `{ Statement, Namespace, Name, FQN, FilePath, Statements, Title *string, Description *string, VersionLiteral string, Version *semver.Version, TagPairs []PolicyTagPair, TagsByKey map[string][]string, Lets, Facts, Rules, Derives, RuleExports map[string]*ExportedRule, Uses, Shapes }` plus unexported `seenIdentifiers`
  - **Behavior:** `Facts` is keyed by **alias** (defaulting to the declared name), `Uses` by import alias, `RuleExports` by rule name. `Title` and `Description` are pointers so "absent" is distinguishable from "empty". `TagPairs` preserves source order; `TagsByKey` is the derived multi-map for queries.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `ExportedRule` — `{ RuleName string, Attachments []*RuleExportAttachment }` / `RuleExportAttachment` — `{ Name string, Value ast.Expression }` / `PolicyTagPair` — `{ Key, Value string }`
  - **Behavior:** Attachment values stay unevaluated expressions, computed by [[runtime]] at decision time.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `createPolicy(ns *Namespace, policy *ast.PolicyStatement, program *ast.Program, idx *Index, nsDerives map[string]*Derive) -> (*Policy, error)`
  - **Behavior:** Walks the body once, skipping comments, advancing a `policyPhase` state machine (`meta → facts → uses → body`) and rejecting any statement that appears out of phase. Builds `TagsByKey` at the end and requires at least one rule export.
  - **Side Effects:** Registers policy-scoped derives into `idx.DerivesByFQN`.
  - **Exceptions:** `ErrPolicyMetadataContiguous` (metadata after facts/uses but before body); `latePolicyHeaderErr` (metadata, fact, use, or derive after the body started); `ErrPolicyFactAfterUse`; duplicate metadata conflicts; `ErrPolicyEmptyTitle`; `ErrPolicyInvalidVersion`; `ErrPolicyEmptyTagKey`; `cannot export unknown rule`; `cannot rebind to existing alias`; `unsupported statement in policy`; `policy '%s' does not export any rules`.

- **Signature:** `(*Policy).AddFact(fact) -> error`
  - **Behavior:** Keys by `fact.Alias` and registers it in `seenIdentifiers`. Enforces that a **required (non-optional) fact cannot carry a default**.
  - **Side Effects:** Mutates `Facts`.
  - **Exceptions:** `ErrConflict("fact declaration", …)`; `ErrInvalidInvocation` for a required fact with a default.

- **Signature:** `(*Policy).AddLet(let) -> error` / `(*Policy).AddRule(rule) -> error` / `(*Policy).AddShape(shape) -> error`
  - **Behavior:** Register-with-conflict-check. Lets and rules share `seenIdentifiers`, so a let and a rule cannot share a name. **Shapes do not** — `AddShape` checks only the `Shapes` map.
  - **Side Effects:** Mutate the corresponding maps.
  - **Exceptions:** `ErrConflict` for each kind; `failed to create shape` wrapping construction errors.

- **Signature:** `(*Policy).String() -> string`
  - **Behavior:** Renders the FQN; used in DAG cycle messages.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Mutable during `createPolicy`, then read-only except for shape hydration in [[index.commit]].
- **Performance/Scale Notes:** One linear pass over the body; `seenIdentifiers` keeps duplicate detection O(1) per declaration.
- **Dependencies Risk:**
  - **Phase ordering is strict and one-way.** Once any body statement (let, rule, export, shape, derive) appears, the phase latches to `body` and every later metadata, fact, or `use` is rejected. Reordering a file's declarations can therefore turn a valid policy invalid with no grammar change.
  - **Metadata must be contiguous *and* first.** Interleaving `title` after a `fact` triggers `ErrPolicyMetadataContiguous`, a different error from the "late header" case — two distinct diagnostics for what a user perceives as the same mistake.
  - **Facts are keyed by alias.** Lookups by declared name miss every aliased fact; the alias is the runtime-facing identifier.
  - **Shapes escape `seenIdentifiers`.** A policy-local shape may share a name with a rule or let, unlike every other declaration pair — an asymmetry worth knowing before assuming names are globally unique inside a policy.
  - **The duplicate rule-export conflict reports the same span twice.** `ErrConflict("rule export", stmt.Span(), stmt.Span())` points at the offending statement for both halves, so the diagnostic cannot show the original export's location.
  - **`VersionLiteral` and `Version` can disagree in whitespace.** SemVer is validated against the trimmed literal while `VersionLiteral` stays verbatim for display.
  - **No helper-only policies.** The zero-export check is a hard error, so a policy that exists purely to be imported from must still export a decision.
