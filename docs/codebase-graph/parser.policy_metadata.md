---
id: parser.policy_metadata
type: Module / File
language: Go
file_path: parser/policy_metadata_title.go, parser/policy_metadata_description.go, parser/policy_metadata_version.go, parser/policy_metadata_tag.go
tags: metadata, documentation, policy-scope, declarative
---

# Node: parser.policy_metadata (Policy Metadata Statements)

## 1. Architectural Role & Intent
Four near-identical productions — `title "…"`, `description "…"`, `version "…"`, and `tag "k" = "v"` — that attach human-facing metadata to a policy. They are grouped into one graph node because they are structurally the same production with different keywords and node types. Their architectural significance is scope: they are legal **only inside a policy block**, and [[parser.statement]] carries a dedicated guard to say so rather than emitting a generic syntax error.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.policy_metadata` | `CALLS` | [[ast]] | Emits `TitleStatement`, `DescriptionStatement`, `VersionStatement`, `TagStatement`. |
| `parser.policy_metadata` | `CALLS` | [[parser.parser]] | Each uses `head`, `expect(<keyword>)`, `advanceExpected(String)`. |
| [[parser.policy]] | `CALLS` | [[parser.policy_metadata]] | All four are registered in the policy-scope handler table. |
| [[parser.statement]] | `DEPENDS_ON` | [[parser.policy_metadata]] | Guards the four keywords at top level with `'<kind>' is only allowed inside a policy`. |
| [[index.policy_stmt]] | `DEPENDS_ON` | [[ast]] | Collects metadata onto the indexed policy. |
| [[pack]] | (contrast) | [[parser.policy_metadata]] | Pack-level metadata lives in the manifest; this is *per-policy* metadata and the two are unrelated. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseTitleStatement(ctx, p) -> ast.Statement`
  - **Behavior:** `title "<string>"` → `TitleStatement`. The context parameter is explicitly discarded.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing keyword or a non-string value.

- **Signature:** `parseDescriptionStatement(ctx, p) -> ast.Statement`
  - **Behavior:** `description "<string>"` → `DescriptionStatement`. Identical shape to `title`.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** As above.

- **Signature:** `parseVersionStatement(ctx, p) -> ast.Statement`
  - **Behavior:** `version "<string>"` → `VersionStatement`. The value is kept as a **raw string** and is not parsed as semver here, unlike the pack manifest version in [[pack]].
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** As above.

- **Signature:** `parseTagStatement(ctx, p) -> ast.Statement`
  - **Behavior:** `tag "<key>" = "<value>"` → `TagStatement`. The only two-value form: both key and value must be **string literals**, not identifiers.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing keyword, a non-string key, a missing `=`, or a non-string value.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **Duplicates are unconstrained.** Nothing prevents two `title` statements or two `tag` entries with the same key in one policy; the resolution rule (first wins, last wins, or error) is left to [[index.policy_stmt]] and is not visible from the syntax.
  - **`version` is an unvalidated string.** A policy may declare `version "not-a-version"` and parse cleanly. Do not assume it is semver-comparable without checking how the index treats it.
  - **Tag keys must be quoted.** `tag env = "prod"` fails because a bare identifier is not a `String` token — an easy authoring mistake given that most configuration languages allow bare keys.
  - **Four files, one production shape.** Any change to metadata statement handling (span computation, duplicate detection, allowing identifiers) must be made in all four places; they share no helper.
