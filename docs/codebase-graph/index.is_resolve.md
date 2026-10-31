---
id: index.is_resolve
type: Function / Endpoint
language: Go
file_path: index/is_resolve.go
tags: type-check, name-resolution, shapes, validation
---

# Node: index.resolveIsExpressions (Shape or Value on the Right of `is`)

## 1. Architectural Role & Intent
Settles what a bare or slash-qualified name means on the right of `x is Name`. The parser ([[parser.is]]) cannot tell a shape from a value, so it records both readings on the `ast.IsExpression`; this pass picks one per site from the shapes and bindings in scope, and turns a name that is neither into an index error instead of a runtime "identifier not found". It is the only pass that rewrites the AST, and it runs first in [[index.validate]] so that no later walk reads a shape name as a value reference.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.is_resolve` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` supplies every expression site with its `kindCheckCtx` scope. |
| `index.is_resolve` | `MUTATES` | [[ast]] | Clears `IsExpression.Right` (type reading) or `IsExpression.Type` (value reading). |
| `index.is_resolve` | `READS_FROM` | [[index.shape]] | Policy and namespace shapes; exported shapes of the namespace a qualified name points at. |
| `index.is_resolve` | `DEPENDS_ON` | [[xerr]] | Errors wrap `xerr.ErrIndex`; `ErrShapeNotFound` and `ErrNotExported` for qualified names. |
| [[index.validate]] | `CALLS` | [[index.is_resolve]] | First step of `validate`. |
| [[runtime.eval_is_type]] | `CALLS` | [[index.is_resolve]] | `LookupShape` for nodes an unvalidated index left ambiguous. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).resolveIsExpressions(ctx) -> error`
  - **Behavior:** For every ambiguous `IsExpression` (both `Right` and `Type` set): if the name is a shape visible from the site and, for a single identifier, not also a value in scope, clears `Right`; if a single identifier is both, reports an error; if it is not a shape, clears `Type` when every identifier of the name is a value in scope and reports an error naming the first that is not. Then, for every `IsExpression` whose `Type` is set, `checkIsTypeShapes` looks up each shape reference inside it (through `?`, `list[...]`, `dict[...]`, `record[...]`) and reports any not visible from the site, so `x is list[Usr]` is an index error too.
  - **Side Effects:** Mutates AST nodes. Idempotent: a resolved node is no longer ambiguous and is skipped.
  - **Exceptions:** All site errors joined: `` `is N`: "N" names both a shape and a value in scope ``, `` `is N`: "x" is neither a shape nor a value in scope ``, `` `is N`: shape "N" is not visible here ``, and the qualified-name lookup errors; each prefixed with its span, sorted into source order (file, line, column) so the output does not depend on map iteration.

- **Signature:** `(*Index).LookupShape(policy *Policy, ns *Namespace, ref *ast.FQN) -> (*Shape, error)`
  - **Behavior:** Mirrors runtime `validateAgainstShapeTypeRef`: `policy.Shapes[ref]`, then `ns.Shapes[ref]` (by the full name), then for three or more segments the last segment in the namespace the prefix names, which must export it. `(nil, nil)` when nothing of that name is visible.
  - **Exceptions:** `ErrShapeNotFound` when the prefix names a namespace without that shape; `ErrNotExported` when the shape is not exported.

## 4. Operational Context & Gotchas
- **Statefulness:** Rewrites the AST once, during validation.
- **Performance/Scale Notes:** One more scoped walk over every expression site.
- **Dependencies Risk:**
  - **Values in scope mirror runtime `evalIdent`**: locals (block lets, lambda parameters), facts, lets, visible derives, and outside a derive body the policy's rules. A runtime binder added without a matching scope entry would make this pass reject valid code.
  - **In a derive, shapes resolve against the derive's policy or namespace**, and [[runtime.eval_is_type]] validates in that same scope (`typeCheckScope` in [[runtime.typeref]]). Only `is` sites are checked here: a shape named by a cast, typed let, or parameter or return type is never looked up by the index, so a misspelt one there is still a runtime error.
  - **A two-segment qualified name never resolves to a shape**, matching the runtime, which only looks names of three or more segments up globally.
  - **Ordering matters.** A walk that runs before this pass and treats `IsExpression.Right` as an expression sees a shape name as an identifier; reference-cycle detection and derive purity rely on running after it.
