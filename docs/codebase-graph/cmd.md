---
id: cmd
type: System / Package
language: Go
file_path: cmd/
tags: cli, entrypoint, orchestration, commands
---

# Node: cmd (Command-Line Interface)

## 1. Architectural Role & Intent
Defines the five user-facing commands and wires the loading pipeline behind each. Every command that touches a pack repeats the same five-step sequence - load pack, create index, load programs, add programs, validate - which makes this package the canonical reference for how the compilation pipeline is assembled in practice.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `cmd` | `IMPORTS` | `ext.binaek.cling` | The CLI framework: command, flag, and argument definitions plus `Hydrate` for typed binding. |
| `cmd` | `CALLS` | [[loader]] | `LoadPack` reads the manifest; `LoadPrograms` parses every policy file. |
| `cmd` | `CALLS` | [[index]] | `CreateIndex`, `SetPack`, `AddProgram`, `Validate`, and `NondeterminismWarnings` in `validate`. |
| `cmd` | `READS_FROM` | [[runtime.js.registry]] | `validate` passes `js.IsNondeterministicExport` to the index's warning pass. |
| `cmd` | `CALLS` | [[capabilities]] | `exec`, `serve`, `validate` and `bench` load `--capabilities`, run `index.CheckCapabilities`, and pass `runtime.WithCapabilities` to the executor. |
| `cmd` | `CALLS` | [[runtime.executor]] | `NewExecutor`, then `ExecPolicy` / `ExecRule`; `bench` calls them in a timed loop. |
| `cmd` | `CALLS` | [[api.http]] | `serve` constructs and drives the HTTP API. |
| `cmd` | `LAYERED_ON` | [[pack]] | `init` constructs and TOML-encodes a new pack manifest. |
| [[main]] | `CALLS` | `cmd` | `Setup` builds the CLI; `Execute` runs it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Setup(ctx context.Context, version string) -> *cling.CLI`
  - **Behavior:** Builds the CLI with pre-run and post-run debug logging and registers all five commands.
  - **Side Effects:** None until run.
  - **Exceptions:** None.

- **Signature:** `Execute(ctx context.Context, cli *cling.CLI, args []string) -> error`
  - **Behavior:** Delegates to `cli.Run`.
  - **Side Effects:** Whatever the command does.
  - **Exceptions:** **Panics** on a nil CLI - a deliberate programmer-error assertion.

- **Signature:** `--capabilities <path>` on `serve`, `exec` and `validate`
  - **Behavior:** Loads the operator's grant ([[capabilities]]) after `Validate`, fails with span-anchored `xerr.CapabilityError`s when the pack uses anything outside it, and hands the grant to the executor for the runtime gates. Empty (the default) grants everything. Shared through `capabilitiesFlag` and `newGrantedExecutor` in `cmd/capabilities.go`, which loads the grant, runs `index.CheckCapabilities`, builds the executor with `runtime.WithCapabilities`, and runs `runtime.CheckModuleGrants` so `require()` in pack TypeScript is checked before serving.
  - **Side Effects:** Reads the file.
  - **Exceptions:** Capabilities parse errors; joined capability denials.

- **Signature:** `sentrie serve [--http-port 7529] [--pack-location ./] [--http-listen local] [--eval-timeout 10s] [--eval-max-timeout <d>]`
  - **Behavior:** Runs the full load-and-validate pipeline, constructs the executor, sets up listeners, starts serving in a goroutine, and blocks on context cancellation before stopping. The default port `7529` spells "PLCY" on a phone keypad.
  - **Side Effects:** Binds sockets; serves decisions; closes the executor after the servers stop. On a shutdown signal it calls `StopServer` with a context from `context.WithoutCancel(ctx)` bounded by `api.DefaultShutdownGracePeriod` (10s; no flag), so in-flight requests finish, and returns its error: nil (exit 0) on a clean drain.
  - **`--eval-max-timeout`:** `serve` only ([#243](https://github.com/sentrie-sh/sentrie/issues/243)). The ceiling a request may ask for with the `Sentrie-Timeout` header, a Go duration parsed by `parseEvalMaxTimeout` against the parsed `--eval-timeout`. Unset (the empty default) means the ceiling is `--eval-timeout`, so requests can only shorten; with `--eval-timeout 0` too, any value is allowed. With `--eval-timeout 0` and a ceiling, requests without the header run under the ceiling. Passed to `api.WithMaxRequestTimeout`.
  - **Exceptions:** Any pipeline error; bind failures; an `--eval-max-timeout` that is unparseable, zero or negative, or below a non-zero `--eval-timeout` (checked after the pack loads, before the executor is built).

- **Signature:** `sentrie exec <rule> [--pack-location .] [--facts '{}'] [--fact-file path] [--output table|json] [--eval-timeout 10s]`
  - **Behavior:** Same pipeline, then resolves the rule path and evaluates. Facts merge from two sources with the **`--facts` flag overriding the fact file**. An empty rule segment evaluates the whole policy.
  - **Side Effects:** Full evaluation; writes to stdout; closes the executor on return.
  - **`--eval-timeout`:** a Go duration string (cling has no duration input), parsed by `parseEvalTimeout`; `0` disables the deadline and negative values are rejected. `serve` passes it to both the executor and `api.WithRequestTimeout`; the API then hands each request its own deadline, which replaces the executor's, so there is one deadline per request.
  - **Exceptions:** File read errors; JSON decode errors; evaluation errors.

- **Signature:** `sentrie bench <target> [--pack-location .] [--facts '{}'] [--fact-file path] [--benchtime 1s|Nx] [--count 1] [--eval-timeout 10s] [--capabilities path]` ([#174](https://github.com/sentrie-sh/sentrie/issues/174))
  - **Behavior:** Loads, indexes, validates and builds the executor once (`loadBenchIndex`, `newGrantedExecutor`), resolves `target` as `exec` resolves its rule (a rule, or a whole policy when the rule segment is empty), runs one untimed evaluation (warm-up, and the check that the target exists and evaluates), then prints `goos:` and `goarch:` lines and one Go benchmark line per `--count` run: `BenchmarkExec/<target>[-GOMAXPROCS] N X ns/op Y B/op Z allocs/op`, so `benchstat` accepts it. `--benchtime` means what it does for `go test`: a positive duration grows the iteration count until a run lasts that long (the last run is reported), `Nx` runs exactly N evaluations. `--count` and `--benchtime` take the double-dash form like every other flag here. `bench` has its own loop (`runBench`, `runN`) rather than `testing.Benchmark`, so it touches no global test flags; B/op and allocs/op are `runtime.MemStats` `TotalAlloc` and `Mallocs` deltas over the timed run, divided by N.
  - **Side Effects:** Full evaluations; writes to stdout; closes the executor on return.
  - **Exceptions:** Unknown target; a target that fails to evaluate (for example a required fact missing); invalid `--benchtime`, `--count`, `--facts` or `--eval-timeout`; pack, fact-file and capabilities errors. The exit status is non-zero. A failure in the warm-up evaluation prints no result line; a failure in a later `--count` run stops after the lines already printed.

- **Signature:** `sentrie validate <rule> [--pack-location .] [--facts '{}']`
  - **Behavior:** Runs the pipeline, constructs an executor and closes it, returning any error. Success means the pack parses, indexes, validates, and can be loaded. It then prints one `warning: <span>: <message>` line to stderr per use of a nondeterministic builtin or builtin-module export (`index.NondeterminismWarnings` with `js.IsNondeterministicExport`); warnings never fail validation.
  - **Side Effects:** None beyond reading files.
  - **Exceptions:** Any pipeline error.

- **Signature:** `sentrie init <name> [--directory .]`
  - **Behavior:** Validates the pack name, builds a default manifest, requires the target to be an **existing empty directory**, and writes the pack file with non-inline TOML tables.
  - **Side Effects:** Creates the pack manifest.
  - **Exceptions:** Invalid name; directory missing or not a directory; directory not empty; write or encode errors.

## 4. Operational Context & Gotchas
- **`bench` measures a warm, long-lived executor.** The executor's call-memoization cache and module loads persist across the timed iterations, as they do in `serve`, so a memoized call is paid once and then served from the cache. B/op and allocs/op are process-wide, so any concurrent goroutine (`ExecPolicy` runs one per exported rule) is included.
- **Statefulness:** Each command builds a fresh index and executor; nothing is shared between invocations.
- **Performance/Scale Notes:** The whole pack is parsed and indexed on every command, including `validate`. Startup cost scales with pack size, and `serve` pays it once before binding - so a large pack delays readiness but never serves a partially-indexed state.
- **Dependencies Risk:**
  - **`validate` accepts a `rule` argument and a `--facts` flag and uses neither.** It validates the entire pack regardless. The `rule` argument is also **required**, so `sentrie validate` alone fails and the user must supply a rule name that is then ignored. This is confusing enough to be worth either honouring or removing.
  - **`exec` checks `runErr != nil` before using it** - the correct pattern that [[api.handle_decision]] is missing. Worth noting the two call sites diverge.
  - **`serve` launches `StartServer` in a goroutine and ignores its outcome.** Combined with that function's inability to report a serve error (see [[api.http]]), a listener that fails after startup leaves the process alive and silent, blocked on `ctx.Done()` with nothing listening.
  - **`serve` checks `ctx.Err()` inside the program loop but `exec` and `validate` do not**, so only `serve` is interruptible during indexing.
  - **`exec` builds `[]*ExecutorOutput{output}` without a nil guard**, the same shape as the API handler - though the subsequent `runErr` check means the nil is not dereferenced on the failure path.
  - **`init` requires a completely empty directory**, so it cannot add a manifest to an existing project - a notable friction point for adopting Sentrie into an existing repository.
  - **`encodePackFile` is a package-level variable specifically so tests can override it**, which is worth knowing before changing its signature.
