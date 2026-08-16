---
id: index.segments
type: Function / Endpoint
language: Go
file_path: index/segments.go
tags: path-parsing, disambiguation, cli-entrypoint, resolution
---

# Node: index.ResolveSegments (Path Splitting)

## 1. Architectural Role & Intent
Splits a user-supplied slash path such as `acme/security/s3/public_bucket` into its namespace, policy, and rule parts. Because namespaces are themselves multi-segment and there is no delimiter distinguishing a namespace boundary from a policy name, the split is inherently ambiguous - this routine resolves it by **greedily preferring the longest prefix that is a registered namespace**. It is the bridge between the flat strings a CLI user types and the exact-match lookups in [[index.resolve]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.segments` | `DEPENDS_ON` | [[ast]] | Joins candidate prefixes with `ast.FQNSeparator`. |
| `index.segments` | `DEPENDS_ON` | [[xerr]] | Returns `ErrNamespaceNotFound` and `ErrPolicyNotFound`. |
| `index.segments` | `CALLS` | [[index.resolve]] | Probes `ResolveNamespace` for each prefix and `ResolvePolicy` for the chosen split. |
| [[cmd]] | `CALLS` | [[index.segments]] | Turns a command-line target into a namespace/policy/rule triple. |
| [[runtime]] | `CALLS` | [[index.segments]] | Same, for evaluation targets supplied as strings. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).ResolveSegments(path string) -> (ns, policy, rule string, err error)`
  - **Behavior:** Splits on `/`, rejects an all-empty path, then walks the parts left to right accumulating a candidate namespace and remembering the **last** prefix that resolved - so the longest matching namespace wins. The next remaining segment is the policy name (verified to exist), and the one after it, if present, is the rule name. The rule component is **optional** and returns as `""` when absent.
  - **Side Effects:** None.
  - **Exceptions:** `ErrNamespaceNotFound` when the path is empty or no prefix matches; `ErrPolicyNotFound` when nothing remains after the namespace; whatever `ResolvePolicy` returns when the policy segment does not exist.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** One `ResolveNamespace` probe per path segment - cheap map lookups, linear in path depth.
- **Dependencies Risk:**
  - **Longest-match can pick the wrong split.** If both `a/b` and `a/b/c` are registered namespaces and `c` is also a policy in `a/b`, the greedy loop prefers the longer namespace and then fails to find the policy - even though a shorter split would have succeeded. There is no backtracking.
  - **Empty segments are skipped, not rejected.** `a//b` and `a/b` resolve identically, and a leading or trailing slash is tolerated, so malformed input is silently normalised rather than reported.
  - **Trailing segments beyond the rule are dropped.** Only `parts[0]` is read as the rule name; anything after it is discarded without an error, so `ns/policy/rule/extra` succeeds and silently ignores `extra`.
  - **The returned rule is not verified to exist.** Unlike the namespace and policy, the rule name is passed through unchecked - callers must resolve it themselves and handle the miss.
  - **The empty-rule return is indistinguishable from a rule literally named `""`**, which cannot occur in practice but means callers should test for the empty string rather than expecting an error.
