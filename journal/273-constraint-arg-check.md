## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #273
- A broken constraint argument, such as an invalid regexp, is only reported for values of the constraint's own kind, so `x is not string@regexp("[")` is `true` for a number.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- Decision agreed with the maintainer before pickup: check constraint arguments when the pack is indexed (count, kinds, and that a regexp compiles), reported by `sentrie validate`, and at runtime check the arguments before the value's kind so a broken argument errors on every input. Constraint names stay with #113.
- _Model: claude-opus-5-5_

### 14:48 - Argument metadata in constraints
- `index` cannot import `constraints` while `constraints` imports `index` for the `*index.Policy` checker parameter, and index tests are internal, so a registration hook could not be tested there with the real tables. The parameter was unused by every checker, so it is dropped from `ConstraintChecker`; that inverts the edge and lets `Index.Validate` read the definitions directly. Ruled out: a hook set from `constraints` init (untestable in `index`), and a generated table like `ast`'s (cannot carry the regexp and divisor rules).
- `ConstraintDefinition` gains `Arg` (checks each argument: number, string, boolean, regexp that compiles, non-zero divisor, number or number list) and `Args` (rules across arguments: range minimum <= maximum), and `CheckArgs` checks the count against `NumArgs` then those. Every table entry is built with `define`, whose `Checker` runs `CheckArgs` first, so the static check, the runtime and a direct checker call share one argument check and cannot disagree. Checker bodies no longer re-check their arguments.
- Messages are unchanged, with one new one: a constraint that takes no arguments and is given some says `<name> constraint takes no arguments` (unreachable from source today, since the parser's generated tables already reject the count).
- _Model: claude-opus-5-5_

### 14:52 - Runtime checks arguments before the value's kind
- The eight typeref validators repeated one evaluate, look up, check loop after the kind check. It is now `bindConstraints` (evaluate the arguments, look up the definition, `CheckArgs`) called first in each validator, and `checkConstraints` after the kind check. `validateValueAgainstTypeRef` also binds before a nullable type's null short-circuit, so `null is not string?@regexp("[")` errors.
- Table selection moved to `constraints.ForTypeRef`, shared with the index check to come.
- An unknown constraint name is now reported at the same point (before the kind check). The parser already rejects unknown names through `ast`'s generated tables, so this is unreachable from source.
- A broken argument inside a list item type is still only reached per item, so an empty list or a non-list does not report it at runtime; the static check covers every position.
- _Model: claude-opus-5-5_

### 20:20 - Static check in Index.Validate
- `checkConstraintArgs` runs last in `validate`. It walks facts, policy lets, shape aliases and fields (each `ShapeStatement` once, so composed fields are not reported twice), derive parameter and return types, and through `forEachScopedExpr` block lets, lambda types, casts and `is` types, recursing into list, dict and record element types. Each constraint's literal arguments are folded to values and checked with the definition's `CheckArgs`, so the index and the runtime share one check. Unknown names are skipped (#113); a non-literal argument is skipped, though the parser only accepts literals (`parseConstraintLiteral`).
- Errors are `xerr.ConstraintArgError` at the constraint's span, matching `ErrIndex` and `constraints.ErrInvalidArgument`; all are reported, sorted, since the walk is over maps.
- Two runtime tests had broken arguments in validated packs (`TestEvalIsTypeCheckErrorsPropagate`, `TestEvalIsTypeDenyRuleDoesNotFailOpen`); those cases now fail validation, so they moved to the index suite, with the runtime path covered on an unvalidated index. Committing without validating panics (`commit` needs the shape DAG), so a "commit only" test helper was ruled out.
- Found while writing tests: `number@in(...)` cannot be written, because `in` lexes as a keyword and the parser wants an identifier after `@`. Out of scope; noted for the maintainer.
- Checked the issue's reproduction with a built binary: `sentrie validate` and `sentrie exec` now both fail with ``a.sentrie:6:28-40: invalid argument to constraint 'regexp': invalid regexp pattern: error parsing regexp: missing closing ]: `[`: index error``.
- _Model: claude-opus-5-5_

### 20:22 - Graph
- New nodes `index.constraint_args` and `runtime.typeref_constraints`. `constraints` loses the `LAYERED_ON index` edge and the "not checked statically" and "`NumArgs` not enforced" gotchas, and records `CheckArgs`, `ForTypeRef` and the unwritable `in` constraint. `index`, `index.validate` and every `runtime.typeref_*` node now say where arguments are checked.
- The `constraints` gotchas about unknown names and empty tables are #113's and were left as they are, though the parser already rejects unknown names.
- _Model: claude-opus-5-5_

### 20:35 - Test and CI pass
- Added `TestCheckArgsCoversEveryConstraint` (constraints): walks all eight tables, requires every constraint to be classified (so a new constraint without `Arg` metadata or a class fails), and checks a valid list, a wrong-kind argument, the wrong count, and that `Checker` reports the same failure as `CheckArgs` for a null value.
- Added runtime tests (`runtime/typeref_constraint_args_test.go`): `is` and `is not` against broken arguments for document, list, number, string and null values (including a nullable type), a broken field nested two shapes deep reached through a derive parameter, and a validated pack of valid arguments (fractional, equal range bounds, list sets, nullable null, list item type) that still evaluates as before.
- Bite checks (mutating the branch, restored from HEAD): `CheckArgs` returning nil fails the table and runtime tests; binding after the kind check in the string, number and trinary validators fails the document/list and nested-shape tests; dropping the bind on the nullable null path fails the nullable cases; dropping `Arg` from `starts_with` fails the table test.
- Ruled out: negative constraint arguments. The parser accepts only literals (`parseConstraintLiteral`), so `range(-5, -1)` is a parse error today (not introduced here); the negative case is covered at the `CheckArgs` level with a box value only. A wrong argument count cannot reach the index check from source, because the parser's generated tables reject it first (`sentrie validate` reports a parsing error), so the count path is covered only on `CheckArgs`.
- _Model: claude-sonnet-5-5_

### 20:44 - Cache compiled regexp patterns
- Review found a `regexp` constraint compiled three times per value (`bindConstraints`' `CheckArgs`, the `define` wrapper's `CheckArgs`, then the match), against once on main, and per element inside a list element type.
- Fixed with a package-level `sync.Map` of compiled patterns keyed by pattern string, used by `regexpArg` and the regexp checker body. Patterns are source literals, so the cache is bounded by the pack; a pattern that fails to compile is not cached. A pattern is now compiled once per process, below main's once per value.
- Did not add a way to skip the wrapper's second `CheckArgs` for bound constraints: with the cache it is a few map lookups, and skipping it would need a second, unchecked entry point on `ConstraintDefinition`.
- Test: `TestRegexpPatternIsCompiledOnce` (same `*regexp.Regexp` before and after checking values, through `export_test.go`), and an invalid pattern errors every time.
- _Model: claude-opus-5-5_

### 03:06 - Graph edges after review
- Dropped the eight `READS_FROM` rows in `constraints` that pointed the typeref validators at the per-kind tables, and the `CALLS runtime.eval` row (constraint argument evaluation) from the seven typeref nodes that had one. Both moved to `bindConstraints` in `runtime.typeref_constraints`, whose rows already say so.
- Folded a duplicated "Test and CI pass" entry out of the journal commit; it was appended twice while commits waited on signing.
- _Model: claude-opus-5-5_
