---
id: index.builtin_kind
type: Module / File
language: Go
file_path: index/builtin_kind.go
tags: type-inference, static-analysis, scoping, shape-traversal
---

# Node: index.KindCheckCtx (Static Kind Inference)

## 1. Architectural Role & Intent
A best-effort static type inferencer: given an expression and a lexical scope, it answers "what `box.ValueKind` will this produce?" — or admits it does not know. It supplies the evidence [[index.builtin_check]] needs to reject bad builtin arguments, and it is deliberately **partial**: every uncertain case returns "unknown", which the checker treats as acceptable, so the analysis never produces a false positive.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.builtin_kind` | `DEPENDS_ON` | [[box]] | `box.ValueKind` is the inference result vocabulary. |
| `index.builtin_kind` | `DEPENDS_ON` | [[ast]] | Maps every `ast.TypeRef` implementation to a kind and uses `ast.RequiredLambdaArity`. |
| `index.builtin_kind` | `DEPENDS_ON` | [[builtins]] | Reads `builtins.Table` for callee identification and `Sig.Result.Kinds` to infer a call's kind. |
| `index.builtin_kind` | `READS_FROM` | [[index.policy]] | Seeds the rule scope from `Facts`, `Lets`, `Derives`, and the namespace's derives. |
| `index.builtin_kind` | `READS_FROM` | [[index.derive]] | Seeds the derive scope from lambda params and the `DefineShort` snapshot. |
| `index.builtin_kind` | `READS_FROM` | [[index.shape]] | Hops through `Model.Fields` to type field-access chains. |
| [[index.builtin_check]] | `CALLS` | [[index.builtin_kind]] | Its sole consumer. |

## 3. Interface Contracts & Public Surface

- **Signature:** `kindCheckCtx` — `{ idx, policy, derive, scope map[string]bindingInfo }` / `bindingInfo` — `{ kindKnown, kind, typeRef, lambda, derive, letInit }`
  - **Behavior:** `bindingInfo` records everything known about a name: a resolved kind, a type annotation, a lambda or derive body for arity checks, or a let initializer to infer from lazily.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `newRuleKindCheckCtx(idx, policy) -> *kindCheckCtx` / `newDeriveKindCheckCtx(idx, d) -> *kindCheckCtx`
  - **Behavior:** The rule variant binds facts **by alias**, lets, policy derives, and then namespace derives **only where the name is not already taken** — encoding policy-over-namespace shadowing. The derive variant binds typed lambda params and the derive's visibility snapshot.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*kindCheckCtx).bindLet(vd)` / `pushLambdaScope(lam) -> *kindCheckCtx` / `cloneBindingScope(scope)`
  - **Behavior:** Scope extension. `pushLambdaScope` returns a **copy** of the context with a cloned scope, so sibling branches never see each other's bindings.
  - **Side Effects:** `bindLet` mutates the receiver's scope, which is why callers clone first.
  - **Exceptions:** None.

- **Signature:** `(*kindCheckCtx).isBuiltinCall(c) -> (*builtins.Decl, bool)`
  - **Behavior:** Only identifier callees qualify. Applies the precedence **local binding > derive > builtin**, matching the runtime's `getTarget`, so a shadowed builtin name is correctly not treated as a builtin.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `typeRefKind(idx, policy, ref) -> (box.ValueKind, bool)`
  - **Behavior:** Maps each type-ref implementation to a kind. Documented special cases: `ShapeTypeRef` delegates to `lookupShapeKind` (alias versus complex record); `NullableTypeRef` is **deliberately unknown** because runtime prechecks pass null through; `RecordTypeRef` maps to `ValueList`.
  - **Side Effects:** None.
  - **Exceptions:** None; unknown is a return value, not an error.

- **Signature:** `resolveShapeReadOnly` / `lookupShapeFieldTypeReadOnly` / `fieldTypeRefHop`
  - **Behavior:** Shape lookup and field traversal that **never mutate** — the `ReadOnly` suffix marks that these must not trigger hydration, since kind checking runs before [[index.commit]].
  - **Side Effects:** None by contract.
  - **Exceptions:** None.

- **Signature:** `(*kindCheckCtx).resolveKind(expr) -> (box.ValueKind, bool)` / `resolveKindGuarded(expr, seen)`
  - **Behavior:** Literals map directly (integer and float both to `ValueNumber`); a lambda is `ValueCallable`; a cast takes its target type; a call takes the builtin's result kind **only when the signature declares exactly one**. `NullLiteral` is explicitly **unknown**, not a null kind. Identifiers recurse through `resolveIdentKindGuarded` with a `seen` set guarding against self-referential lets.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*kindCheckCtx).resolveFieldAccessChain(e) -> (box.ValueKind, bool)`
  - **Behavior:** Flattens `a.b.c` into a root plus an ordered field list, then hops the type refs field by field. The root must be an identifier with a **type annotation**; unannotated lets are intentionally unknown.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*kindCheckCtx).resolveCallableArity(expr) -> (int, bool)`
  - **Behavior:** Returns the required parameter count for a lambda literal, a bound lambda, or a derive — the input to callback-arity diagnostics.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `kindAllowed(kinds, k) -> bool` / `builtinParamSigAt(sig, i) -> (ParamSig, bool)`
  - **Behavior:** Membership test, and positional parameter lookup that falls back to the variadic tail once fixed params are exhausted.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Contexts are values; scopes are cloned rather than shared. Safe to fork down a walk.
- **Performance/Scale Notes:** Scope cloning at every block and lambda makes deeply nested bodies quadratic in binding count. `resolveIdentKindGuarded` allocates a fresh `seen` set per let-initializer branch.
- **Dependencies Risk:**
  - **Unknown means "allowed".** Every partial case here directly weakens [[index.builtin_check]]. That trade is deliberate — no false positives — but it means static builtin checking catches far less than its presence suggests.
  - **`typeRefKind`'s switch must list every `ast.TypeRef` implementation**, as its comment states. A new type ref silently infers "unknown" instead of failing, so the gap is invisible until someone notices a missed diagnostic.
  - **Read-only shape access is a hard requirement.** These helpers run **before** commit, when shapes are not yet hydrated, so field lookups can miss inherited fields. Making them trigger hydration would reorder the lifecycle and corrupt the topological guarantee in [[index.commit]].
  - **Integer and float collapse to one kind.** The distinction preserved by the parser is erased here, matching `box`'s single numeric kind — so no arity or kind check can distinguish them.
  - **Precedence duplication with the runtime.** `isBuiltinCall` re-implements the runtime's name-resolution order; drift between the two produces checks applied to the wrong callee.
