---
id: infra.build_metadata
type: Infrastructure
language: N/A
file_path: (external)
tags: infrastructure, boundary, build, provenance, observability
---

# Node: Build Metadata (Go Toolchain Build Info)

## 1. Architectural Role & Intent
Provenance data that the Go toolchain embeds into the binary at link time and exposes at runtime through `runtime/debug.ReadBuildInfo`. Sentrie uses it so that a deployed binary can report exactly which commit it was built from without that information having to be threaded through ldflags by hand or kept in sync in source.

This is the only node in the graph whose data originates from the build system rather than from an operator, a policy author, or a request. It is read exactly once, by [[version]], and surfaced through the CLI.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[version]] | `READS_FROM` | `infra.build_metadata` | Pulls `vcs.revision`, `vcs.time`, and `vcs.modified` from the module build settings baked in at compile time. |

## 3. Interface Contracts & Public Surface

- **Signature:** `vcs.revision`
  - **Behavior:** Full commit SHA of the source tree the binary was built from.
  - **Side Effects:** None.
  - **Exceptions:** Absent when the build did not happen inside a VCS checkout.

- **Signature:** `vcs.time`
  - **Behavior:** Commit timestamp, RFC 3339.
  - **Side Effects:** None.
  - **Exceptions:** Absent under the same conditions as `vcs.revision`.

- **Signature:** `vcs.modified`
  - **Behavior:** `"true"` when the working tree had uncommitted changes at build time — the dirty-build marker.
  - **Side Effects:** None.
  - **Exceptions:** Absent under the same conditions as above.

## 4. Operational Context & Gotchas
- **Statefulness:** Immutable for the lifetime of the binary; fixed at link time.
- **Performance/Scale Notes:** Read once, on demand. No measurable cost.
- **Dependencies Risk:**
  - **The settings are absent for builds outside a VCS checkout**, which includes `go install` from a module proxy and most container builds that copy sources rather than clone. Version output degrades silently to whatever default [[version]] carries rather than reporting that provenance is unavailable.
  - **`vcs.modified` is the only signal that a binary was built from a dirty tree.** If it is dropped from the surfaced output, a locally-patched binary becomes indistinguishable from a released one.
  - **This is provenance, not identity.** It records the commit, not the release tag, so it cannot answer "which released version is this" on its own.
