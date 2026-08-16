---
id: cmd
type: System / Package
language: Go
file_path: cmd/
tags: cli, entrypoint, orchestration, commands
---

# Node: cmd (Command-Line Interface)

## 1. Architectural Role & Intent
Defines the four user-facing commands and wires the loading pipeline behind each. Every command that touches a pack repeats the same five-step sequence — load pack, create index, load programs, add programs, validate — which makes this package the canonical reference for how the compilation pipeline is assembled in practice.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `cmd` | `DEPENDS_ON` | `ext.cling` | The CLI framework: command, flag, and argument definitions plus `Hydrate` for typed binding. |
| `cmd` | `CALLS` | [[loader]] | `LoadPack` reads the manifest; `LoadPrograms` parses every policy file. |
| `cmd` | `CALLS` | [[index.package]] | `CreateIndex`, `SetPack`, `AddProgram`, `Validate`. |
| `cmd` | `CALLS` | [[runtime.executor]] | `NewExecutor`, then `ExecPolicy` / `ExecRule`. |
| `cmd` | `CALLS` | [[api.http]] | `serve` constructs and drives the HTTP API. |
| `cmd` | `DEPENDS_ON` | [[pack]] | `init` constructs and TOML-encodes a new pack manifest. |
| [[main]] | `CALLS` | `cmd` | `Setup` builds the CLI; `Execute` runs it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Setup(ctx context.Context, version string) -> *cling.CLI`
  - **Behavior:** Builds the CLI with pre-run and post-run debug logging and registers all four commands.
  - **Side Effects:** None until run.
  - **Exceptions:** None.

- **Signature:** `Execute(ctx context.Context, cli *cling.CLI, args []string) -> error`
  - **Behavior:** Delegates to `cli.Run`.
  - **Side Effects:** Whatever the command does.
  - **Exceptions:** **Panics** on a nil CLI — a deliberate programmer-error assertion.

- **Signature:** `sentrie serve [--http-port 7529] [--pack-location ./] [--http-listen local]`
  - **Behavior:** Runs the full load-and-validate pipeline, constructs the executor, sets up listeners, starts serving in a goroutine, and blocks on context cancellation before stopping. The default port `7529` spells "PLCY" on a phone keypad.
  - **Side Effects:** Binds sockets; serves decisions.
  - **Exceptions:** Any pipeline error; bind failures.

- **Signature:** `sentrie exec <rule> [--pack-location .] [--facts '{}'] [--fact-file path] [--output table|json]`
  - **Behavior:** Same pipeline, then resolves the rule path and evaluates. Facts merge from two sources with the **`--facts` flag overriding the fact file**. An empty rule segment evaluates the whole policy.
  - **Side Effects:** Full evaluation; writes to stdout.
  - **Exceptions:** File read errors; JSON decode errors; evaluation errors.

- **Signature:** `sentrie validate <rule> [--pack-location .] [--facts '{}']`
  - **Behavior:** Runs the pipeline and constructs an executor, returning any error. Success means the pack parses, indexes, validates, and can be loaded.
  - **Side Effects:** None beyond reading files.
  - **Exceptions:** Any pipeline error.

- **Signature:** `sentrie init <name> [--directory .]`
  - **Behavior:** Validates the pack name, builds a default manifest, requires the target to be an **existing empty directory**, and writes the pack file with non-inline TOML tables.
  - **Side Effects:** Creates the pack manifest.
  - **Exceptions:** Invalid name; directory missing or not a directory; directory not empty; write or encode errors.

## 4. Operational Context & Gotchas
- **Statefulness:** Each command builds a fresh index and executor; nothing is shared between invocations.
- **Performance/Scale Notes:** The whole pack is parsed and indexed on every command, including `validate`. Startup cost scales with pack size, and `serve` pays it once before binding — so a large pack delays readiness but never serves a partially-indexed state.
- **Dependencies Risk:**
  - **`validate` accepts a `rule` argument and a `--facts` flag and uses neither.** It validates the entire pack regardless. The `rule` argument is also **required**, so `sentrie validate` alone fails and the user must supply a rule name that is then ignored. This is confusing enough to be worth either honouring or removing.
  - **`exec` checks `runErr != nil` before using it** — the correct pattern that [[api.handle_decision]] is missing. Worth noting the two call sites diverge.
  - **`serve` launches `StartServer` in a goroutine and ignores its outcome.** Combined with that function's inability to report a serve error (see [[api.http]]), a listener that fails after startup leaves the process alive and silent, blocked on `ctx.Done()` with nothing listening.
  - **`serve` checks `ctx.Err()` inside the program loop but `exec` and `validate` do not**, so only `serve` is interruptible during indexing.
  - **`exec` builds `[]*ExecutorOutput{output}` without a nil guard**, the same shape as the API handler — though the subsequent `runErr` check means the nil is not dereferenced on the failure path.
  - **`init` requires a completely empty directory**, so it cannot add a manifest to an existing project — a notable friction point for adopting Sentrie into an existing repository.
  - **`encodePackFile` is a package-level variable specifically so tests can override it**, which is worth knowing before changing its signature.
