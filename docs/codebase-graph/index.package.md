---
id: index.package
type: System / Package
language: Go
file_path: index/
tags: semantic-analysis, symbol-table, validation, dependency-graph, middle-end
---

# Node: Index (Semantic Model and Validator)

## 1. Architectural Role & Intent
`index` is Sentrie's middle-end: it consumes the `ast.Program` values produced by [[parser]] and builds a queryable semantic model — namespaces, policies, rules, shapes, derives, and their exports — then validates that model before anything is allowed to execute. It is where every rule that a grammar cannot express is enforced: declaration ordering inside a policy, name uniqueness across kinds, shape composition, derive purity, builtin call arity and argument kinds, and cycle freedom across four separate dependency graphs. [[runtime]] executes only against a validated, committed index.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.package` | `LAYERED_ON` | [[ast]] | Consumes `ast.Program` and every statement type; retains AST nodes by reference throughout the model. |
| `index.package` | `LAYERED_ON` | [[dag]] | Builds four graphs — rule, shape, derive, and a per-policy identifier graph — for cycle detection and topological ordering. |
| `index.package` | `LAYERED_ON` | [[pack]] | Holds the pack manifest alongside the semantic model. |
| `index.package` | `LAYERED_ON` | [[xerr]] | Every diagnostic is an `xerr` sentinel or `ErrConflict`, giving structured two-span conflict errors. |
| `index.package` | `LAYERED_ON` | [[builtins]] | Reads builtin declarations and signatures to validate call sites; consults `IsDeriveSafe` for purity. |
| `index.package` | `LAYERED_ON` | [[box]] | Uses `box.ValueKind` as the vocabulary for static argument-kind checking. |
| `index.package` | `IMPORTS` | `ext.masterminds.semver` | Parses and validates per-policy `version` metadata. |
| [[loader]] | `LAYERED_ON` | [[index.package]] | Supplies the parsed programs that populate the index. |
| [[runtime]] | `LAYERED_ON` | [[index.package]] | Resolves namespaces, policies, rules, and derives against the committed index during evaluation. |
| [[cmd]] | `CALLS` | [[index.package]] | `validate` and `exec` build an index, call `Validate`, and report or proceed. |
| [[api]] | `CALLS` | [[index.package]] | Serves evaluation requests against a pre-built index. |

## 3. Interface Contracts & Public Surface

- **Signature:** `CreateIndex() -> *Index`
  - **Behavior:** Allocates an empty index with its lock, maps, and the two `sync.Once` guards for validation and commit. See [[index.index]].
  - **Side Effects:** Allocation only.
  - **Exceptions:** None.

- **Signature:** `(*Index).SetPack(ctx, p *pack.PackFile) -> error` / `(*Index).AddProgram(ctx, astProgram *ast.Program) -> error`
  - **Behavior:** The two population entrypoints. `AddProgram` is where namespaces are created, top-level statements are dispatched, and most structural conflicts are detected. See [[index.index]].
  - **Side Effects:** Mutates the index under a write lock.
  - **Exceptions:** Name conflicts, unknown derive exports, unsupported top-level statements.

- **Signature:** `(*Index).Validate(ctx) -> error` / `(*Index).IsValid(ctx) -> error`
  - **Behavior:** Runs the full validation pipeline exactly once and then commits. See [[index.validate]].
  - **Side Effects:** Populates the rule and shape DAGs; triggers [[index.commit]].
  - **Exceptions:** Any validation failure, wrapped as `validation error: …`.

- **Signature:** `(*Index).Commit(ctx) -> error`
  - **Behavior:** Hydrates shape composition in topological order. See [[index.commit]].
  - **Side Effects:** Mutates shape models in place.
  - **Exceptions:** Composition conflicts and unresolvable base shapes.

- **Signature:** Resolution surface — `ResolveNamespace`, `ResolvePolicy`, `ResolveShape`, `ResolveDerive`, `ResolveSegments`, `VerifyRuleExported`, `VerifyShapeExported`, `VerifyDeriveExported`, plus the `RuleFQN`/`ShapeFQN`/`DeriveFQN` builders
  - **Behavior:** The read API used by [[runtime]] and [[cmd]]. See [[index.resolve]] and [[index.segments]].
  - **Side Effects:** None.
  - **Exceptions:** `xerr` not-found and not-exported sentinels.

- **Signature:** Model types — `Index`, `Namespace`, `Policy`, `Rule`, `Shape`, `Derive`, `Program`, `ExportedRule`, `ExportedShape`, `ExportedDerive`, `ShapeModel`, `ShapeModelField`, `PolicyTagPair`, `RuleExportAttachment`
  - **Behavior:** The exported semantic vocabulary. All fields are public and all retain AST node references.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** **Stateful, mutable, and phase-ordered.** An index is populated (`SetPack`, `AddProgram`), then validated, then committed, then read. `Validate` and `Commit` are each `sync.Once`-guarded, so the first result is cached permanently — including a failure.
- **Performance/Scale Notes:** Validation walks every rule, let, shape, and derive several times over (reference cycles, rule cycles, shape cycles, derive cycles, purity, builtin kinds), so cost is roughly linear in policy count times checks. All of it happens once per index, not per evaluation.
- **Dependencies Risk:**
  - **Order-sensitive population.** Derive visibility is captured as a **bind-time snapshot** when each derive is registered, so a program added *later* does not retroactively become visible to derives registered earlier. Load dependent programs after their helpers, or keep helpers in the same file — see [[index.derive]].
  - **`AddProgram` skips the first statement.** Its loop starts at index 1 on the assumption that statement zero is the namespace. A file whose leading statements are comments (legal per [[parser.parse]]) has its second statement silently ignored.
  - **Locking is partial.** The write lock guards `SetPack` and `AddProgram`, but `Validate`, `Commit`, and every `Resolve*` read the model without acquiring `theLock`. Concurrent population and evaluation is unsafe.
  - **Every policy must export a decision.** `createPolicy` fails a policy with zero rule exports, so an index cannot contain a "library" policy that only defines helpers.
  - **Failure is sticky.** Because validation is `sync.Once`-guarded, an index that failed validation can never be repaired and re-validated; build a fresh one.
