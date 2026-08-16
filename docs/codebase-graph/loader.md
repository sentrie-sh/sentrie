---
id: loader
type: System / Package
language: Go
file_path: loader/
tags: filesystem, manifest-discovery, schema-validation, bootstrap, toml
---

# Node: Loader (Pack Discovery and Manifest Ingestion)

## 1. Architectural Role & Intent
`loader` is the filesystem boundary of Sentrie: it locates a `sentrie.pack.toml` by walking up from a supplied path, decodes and validates it into a [[pack]] `PackFile`, and then walks the pack root to parse every `.sentrie` source into an `ast.Program`. It exists so that every entrypoint — CLI [[cmd]] and HTTP [[api]] alike — resolves "which pack am I running?" through one deterministic algorithm. It is the only package that performs both TOML decoding and JSON-Schema validation of the manifest.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `loader` | `MUTATES` | [[pack]] | Decodes TOML into a `pack.PackFile` and injects the resolved `Location` directory. |
| `loader` | `LAYERED_ON` | [[constants]] | Composes `PackFileName` from `constants.APPNAME` + `constants.PackFileExtension`; filters sources by `constants.PolicyFileExtension`. |
| `loader` | `CALLS` | [[parser]] | `LoadPrograms` constructs a `parser.NewParser` per discovered file and calls `ParseProgram`. |
| `loader` | `LAYERED_ON` | [[ast]] | Returns `[]*ast.Program`. |
| `loader` | `IMPORTS` | `ext.pelletier.go-toml` | Manifest decoding, performed twice: once into a raw map for unknown-key detection, once into the typed struct. |
| `loader` | `IMPORTS` | `ext.xeipuuv.gojsonschema` | Compiles the embedded Draft-7 `schema.json` at package init and validates the marshalled manifest. |
| `loader` | `READS_FROM` | [[infra.filesystem]] | `os.Stat`, `os.ReadFile`, and `fs.WalkDir` over the pack directory tree. |
| [[cmd]] | `CALLS` | [[loader]] | `exec`, `init`, `serve`, and `validate` all bootstrap through `LoadPack`/`LoadPrograms`. |
| [[index.package]] | `LAYERED_ON` | [[loader]] | Consumes the programs produced here as its indexing input. |

## 3. Interface Contracts & Public Surface

- **Signature:** `LoadPack(ctx: context.Context, root: string) -> (*pack.PackFile, error)`
  - **Behavior:** The pack bootstrap. Locates the manifest (see below), rejects an empty file, decodes the TOML into a raw map to reject **unknown top-level tables** (only `schema`, `pack`, `engine`, `permissions`, `metadata` are allowed), pre-checks that an `[engine]` table carries a non-empty `sentrie` field, decodes into `pack.PackFile`, enforces that `schema` and `pack.name` are present and that the name matches the dotted-identifier regex, runs JSON-Schema validation, and finally sets `Location` to the manifest's directory.
  - **Side Effects:** Reads the filesystem; **mutates** the returned `PackFile.Location`.
  - **Exceptions:** `ctx.Err()` if already cancelled; `locate pack file: …` wrapping `ErrPackFileNotFound`; `pack file is empty`; `failed to parse pack file` for malformed TOML; `unknown top-level table '[x]'`; `engine table exists but 'sentrie' field is required`; `schema version is required`; `name is required`; `name must be a valid identity`; `schema validation failed: …` with per-field descriptions.

- **Signature:** `LoadPrograms(ctx: context.Context, packFile: *pack.PackFile) -> ([]*ast.Program, error)`
  - **Behavior:** Recursively walks `packFile.Location`, parsing every file whose extension matches the policy extension into an `ast.Program`. Programs are accumulated in walk order.
  - **Side Effects:** Opens files (see the leak note below); drives [[parser]] and therefore [[lexer]].
  - **Exceptions:** Aborts the entire walk on the first parse error, open error, or context cancellation — the returned slice is then incomplete and must be discarded. A parser returning a nil program is skipped silently.

- **Signature:** `ValidatePackFile(packFile: *pack.PackFile) -> error`
  - **Behavior:** Marshals the manifest to JSON and validates it against the embedded Draft-7 schema, flattening all violations into a single multi-line error with `field: description` per line (the root is rendered as `root`). Exported so callers can validate a synthesized manifest without touching disk.
  - **Side Effects:** None.
  - **Exceptions:** JSON marshal failure; schema evaluation failure; the aggregated validation error.

- **Signature:** `IsValidPackName(name: string) -> bool`
  - **Behavior:** Matches `^([a-zA-Z][a-zA-Z0-9_-]*)(\.[a-zA-Z][a-zA-Z0-9_-]*)*$` — dot-separated segments, each starting with a letter, allowing hyphens and underscores.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `PackFileName` (var) / `NameRegex` (var) / `ErrPackFileNotFound`, `ErrPackFileLoadFailed` (sentinels)
  - **Behavior:** `PackFileName` is derived from [[constants]] rather than hardcoded. `ErrPackFileNotFound` is the only sentinel actually returned — `ErrPackFileLoadFailed` is declared but unused, so do not match on it.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless functions over package-level singletons. The compiled JSON schema and its loader are process-wide, built once in `init()`.
- **Performance/Scale Notes:** The manifest is decoded **twice** (raw map, then struct) and marshalled to JSON a third time for schema validation — irrelevant for a one-shot CLI run, worth knowing for a long-lived [[api]] server that reloads packs per request. `LoadPrograms` parses files **sequentially**; parse time scales linearly with pack size.
- **Dependencies Risk:**
  - **`init()` panics.** If the embedded `schema.json` fails to compile, the package panics at import time, taking down the binary before `main` runs. Any edit to `schema.json` must be exercised by simply running the binary.
  - **File-handle leak.** `LoadPrograms` calls `os.Open` per policy file and never closes the handle — the parser holds the reader and nothing releases it. On a large pack or a long-lived server this exhausts descriptors.
  - **Upward search can escape the intended pack.** `locatePackFile` walks parent directories until it hits `/` (or a Windows drive root), so running from a subdirectory of an *unrelated* project can silently bind to an ancestor's manifest. It refuses to start from `/` or an empty path, but there is no other containment.
  - **Cancellation is coarse.** `ctx` is checked at entry and per directory entry during the walk, but not during file read or parse; a large file will finish parsing after cancellation.
  - **Test seams are package globals.** `statPackFile`, `filepathAbs`, and `validatePackDocument` are mutable package-level function variables swapped by tests — they are process-global and not safe to reassign concurrently.
  - **`loader/load.go` is empty** (license header only); the naming suggests an entrypoint that does not exist. The real entrypoints are `pack.go` and `file.go`.
