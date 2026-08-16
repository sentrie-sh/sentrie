---
id: version
type: System / Package
language: Go
file_path: version/
tags: build-metadata, cli-output, observability, release
---

# Node: Version (Build Information Reporter)

## 1. Architectural Role & Intent
`version` assembles and renders the binary's build identity — semantic version, git commit, tree cleanliness, build date, builder — by reading Go's embedded `debug.BuildInfo` and layering caller-supplied application details on top. It exists to give `sentrie version` and support diagnostics a single authoritative provenance string without requiring linker `-ldflags` plumbing for the VCS fields. It is a leaf presentation utility with no dependency on any other Sentrie package.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `version` | `DEPENDS_ON` | `ext.masterminds.semver` | `String()` parses `GitVersion` through `semver.MustParse` before rendering. |
| `version` | `DEPENDS_ON` | (stdlib: `runtime/debug`, `text/tabwriter`, `strings`, `fmt`) | Reads the build info embedded by the Go toolchain; aligns the build-info block for terminal output. |
| `version` | `READS_FROM` | `infra.build_metadata` | Pulls `vcs.revision`, `vcs.time`, and `vcs.modified` from the module build settings baked in at compile time. |
| [[main]] | `CALLS` | [[version]] | The single consumer: builds an `Info` with app name/description/website and the ASCII banner, then renders it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Info` struct — `{ Name, Description, Website, GitVersion, GitCommit, GitTreeState, BuildDate, BuiltBy string }` (plus unexported `asciiName`)
  - **Behavior:** The value object carrying all provenance fields. Fields are exported deliberately to mirror the `caarlos0/go-version` API this package replaces, so a caller (or release tooling) can populate them directly.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `GetVersionInfo(opts ...Option) -> Info`
  - **Behavior:** Constructs an `Info` pre-filled from `debug.ReadBuildInfo()` — mapping `vcs.revision` → `GitCommit`, `vcs.time` → `BuildDate`, and `vcs.modified` → `GitTreeState` as the literal strings `"dirty"`/`"clean"` — then applies options, which **override** the pre-filled values. `GitVersion` is taken from `Main.Version` only when it is non-empty and not the placeholder `"(devel)"`.
  - **Side Effects:** None beyond reading process build metadata.
  - **Exceptions:** None. A missing `debug.BuildInfo` is silently tolerated and yields an `Info` with empty VCS fields.

- **Signature:** `Option` type — `func(*Info)`
  - **Behavior:** Functional-options mutator applied in order by `GetVersionInfo`.
  - **Side Effects:** Mutates the `Info` under construction.
  - **Exceptions:** None.

- **Signature:** `WithAppDetails(name, description, website string) -> Option` / `WithASCIIName(asciiName string) -> Option`
  - **Behavior:** The two supplied options: identity strings and the banner art printed above them. `asciiName` is unexported and settable only through this option.
  - **Side Effects:** None until applied.
  - **Exceptions:** None.

- **Signature:** `(Info).String() -> string`
  - **Behavior:** Renders the full block: ASCII banner, `Name vX.Y.Z`, description, website, then a tab-aligned build-info table containing only the non-empty fields. Every section is conditional, so a bare `Info{}` renders as near-empty rather than a skeleton of blank labels.
  - **Side Effects:** None.
  - **Exceptions:** **Panics** when `GitVersion` is non-empty but not valid semver — `semver.MustParse` is called without a guard. See below.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless. `Info` is a value type, safe to copy and share.
- **Performance/Scale Notes:** `GetVersionInfo` iterates the module build settings once; `String()` allocates a builder and a tabwriter. Both are invoked at most once per process, at CLI startup.
- **Dependencies Risk:** No runtime failure domain, but two release-engineering hazards:
  - **`MustParse` panic.** If the module version embedded by the toolchain (or an override) is not parseable semver — for example a raw git describe string like `v0.3.1-4-gabc1234`, or a tag without the expected shape — `String()` panics and `sentrie version` crashes. This is a build-tagging concern, not a code path anyone can defend against at runtime.
  - **`BuiltBy` is never populated.** It is rendered when set but no code or option assigns it, so it only appears if release tooling constructs `Info` literally rather than through `GetVersionInfo`.
  - **Development builds report no version.** `Main.Version` is `"(devel)"` for `go run`/`go build` outside a module release, which is explicitly filtered out, so `GitVersion` is empty and the name line renders without a version — expected, not a bug.
