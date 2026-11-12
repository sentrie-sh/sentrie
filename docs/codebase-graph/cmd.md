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
| `cmd` | `CALLS` | [[format]] | `fmt` runs `format.Source` on each `.sentrie` file it finds. |
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

- **Signature:** `--skip-engine-check` on `serve`, `exec`, `validate` and `bench` ([#157](https://github.com/sentrie-sh/sentrie/issues/157))
  - **Behavior:** After `LoadPack`, `checkEngine` (`cmd/engine.go`) compares the pack's `[engine] sentrie` constraint with the running binary's version, which `main` hands over with `SetEngineVersion`. Comparison is strict semver: a prerelease sorts before its release, so `0.0.5-rc1` does not satisfy `>=0.0.5` but does satisfy `>=0.0.4` (the constraint is copied with `IncludePrerelease` set, because Masterminds otherwise drops prereleases). A mismatch is an error naming the constraint and the binary version, under `validate` too. A dev or unversioned binary, one whose version is not semver, and one built from source (a Go pseudo-version, with or without a reachable tag, or a `+dirty` stamp) pass with a `warning:` line on stderr. A pack with no `[engine]` table passes silently. The flag skips the check.
  - **Side Effects:** Writes the warning to stderr.
  - **Exceptions:** The constraint mismatch error.

- **Signature:** `sentrie serve [--http-port 7529] [--pack-location ./] [--http-listen local] [--http-max-body-bytes 1048576] [--eval-timeout 10s] [--eval-max-timeout <d>]`
  - **Behavior:** Runs the full load-and-validate pipeline, constructs the executor, sets up listeners, starts serving in a goroutine, and blocks on context cancellation before stopping. The default port `7529` spells "PLCY" on a phone keypad.
  - **Side Effects:** Binds sockets; serves decisions; closes the executor after the servers stop. On a shutdown signal it calls `StopServer` with a context from `context.WithoutCancel(ctx)` bounded by `api.DefaultShutdownGracePeriod` (10s; no flag), so in-flight requests finish, and returns its error: nil (exit 0) on a clean drain.
  - **`--eval-max-timeout`:** `serve` only ([#243](https://github.com/sentrie-sh/sentrie/issues/243)). The ceiling a request may ask for with the `Sentrie-Timeout` header, a Go duration parsed by `parseEvalMaxTimeout` against the parsed `--eval-timeout`. Unset (the empty default) means the ceiling is `--eval-timeout`, so requests can only shorten; with `--eval-timeout 0` too, any value is allowed. With `--eval-timeout 0` and a ceiling, requests without the header run under the ceiling. Passed to `api.WithMaxRequestTimeout`.
  - **`--http-max-body-bytes`:** `serve` only ([#115](https://github.com/sentrie-sh/sentrie/issues/115)). The largest request body, a plain integer byte count (cling int input), default `api.DefaultMaxBodyBytes` (1 MiB); `0` disables the limit. Checked by `validateMaxBodyBytes` before the pack loads and passed to `api.WithMaxBodyBytes`.
  - **Exceptions:** Any pipeline error; bind failures; a negative `--http-max-body-bytes`; an `--eval-max-timeout` that is unparseable, zero or negative, or below a non-zero `--eval-timeout` (checked after the pack loads, before the executor is built).

- **Signature:** `sentrie exec <rule> [--pack-location .] [--facts '{}'] [--fact-file path] [--output table|json] [--eval-timeout 10s]`
  - **Behavior:** Same pipeline, then resolves the rule path and evaluates. Facts merge from two sources with the **`--facts` flag overriding the fact file**: objects merge deeply, any other value, a list included, is replaced (`mergeFacts`, [#147](https://github.com/sentrie-sh/sentrie/issues/147); `bench` merges the same way). Facts are keyed by declared name, not alias. Both are decoded with `json.Decoder.UseNumber`, so an integral number that fits int64 stays exact, and checked with `box.CheckJSONNumbers`, so a number float64 cannot hold is an error ([#166](https://github.com/sentrie-sh/sentrie/issues/166)). With `--output json` an encoding failure (a non-finite result) is returned as an error instead of an empty stdout and exit 0. An empty rule segment evaluates the whole policy. For a whole policy, a failing rule does not hide the others ([#143](https://github.com/sentrie-sh/sentrie/issues/143)): every rule's result is printed (the JSON output gives a failed rule an `error` object, the table an `Errors:` section) and exec then fails with `<n> of <m> rules failed`, exit 1. A named rule that fails, and a fact error on a policy (a missing required fact, say), print only the error.
  - **Side Effects:** Full evaluation; writes to stdout; closes the executor on return.
  - **`--eval-timeout`:** a Go duration string (cling has no duration input), parsed by `parseEvalTimeout`; `0` disables the deadline and negative values are rejected. `serve` passes it to both the executor and `api.WithRequestTimeout`; the API then hands each request its own deadline, which replaces the executor's, so there is one deadline per request.
  - **Exceptions:** File read errors; JSON decode errors; evaluation errors.

- **Signature:** `sentrie bench <target> [--pack-location .] [--facts '{}'] [--fact-file path] [--benchtime 1s|Nx] [--count 1] [--eval-timeout 10s] [--capabilities path] [--skip-engine-check]` ([#174](https://github.com/sentrie-sh/sentrie/issues/174))
  - **Behavior:** Loads the pack, checks its `[engine]` constraint (`checkEngine`, as `exec` does), indexes, validates and builds the executor once (`loadBenchIndex`, `newGrantedExecutor`), resolves `target` as `exec` resolves its rule (a rule, or a whole policy when the rule segment is empty), runs one untimed evaluation (warm-up, and the check that the target exists and evaluates), then prints `goos:` and `goarch:` lines and one Go benchmark line per `--count` run: `BenchmarkExec/<target>[-GOMAXPROCS] N X ns/op Y B/op Z allocs/op`, so `benchstat` accepts it. `--benchtime` means what it does for `go test`: a positive duration grows the iteration count until a run lasts that long (the last run is reported), `Nx` runs exactly N evaluations. `--count` and `--benchtime` take the double-dash form like every other flag here. `bench` has its own loop (`runBench`, `runN`) rather than `testing.Benchmark`, so it touches no global test flags; B/op and allocs/op are `runtime.MemStats` `TotalAlloc` and `Mallocs` deltas over the timed run, divided by N.
  - **Side Effects:** Full evaluations; writes to stdout; closes the executor on return.
  - **Exceptions:** Unknown target; a target that fails to evaluate (for example a required fact missing); invalid `--benchtime`, `--count`, `--facts` or `--eval-timeout`; pack, engine constraint, fact-file and capabilities errors. The exit status is non-zero. A failure in the warm-up evaluation prints no result line; a failure in a later `--count` run stops after the lines already printed.

- **Signature:** `sentrie validate [--pack-location .] [--capabilities <path>]`
  - **Behavior:** Takes no positional argument: a non-flag word is an error (`validate takes no arguments`), a single-dash word such as `-x` is an error (`validate: unknown flag`), and `--facts` is no longer a flag, so cling rejects it ([#155](https://github.com/sentrie-sh/sentrie/issues/155)). Runs the pipeline, constructs an executor and closes it, returning any error. Success means the pack parses, indexes, validates, and can be loaded. It then prints one `warning: <span>: <message>` line to stderr per use of a nondeterministic builtin or builtin-module export (`index.NondeterminismWarnings` with `js.IsNondeterministicExport`); warnings never fail validation.
  - **Side Effects:** None beyond reading files.
  - **Exceptions:** Any pipeline error.

- **Signature:** `sentrie fmt <pattern> [--check] [--diff]` ([#172](https://github.com/sentrie-sh/sentrie/issues/172))
  - **Behavior:** Formats `.sentrie` files with [[format]]. The one required `pattern` follows `go fmt`: `./...` or `dir/...` walks every `*.sentrie` file below the directory (hidden directories and symlinks met on the walk are skipped; a symlinked root is followed); `dir` (including `.`) covers only the files directly in it; a file path is formatted whatever its extension. No pattern, or a second positional, is an error. Rewrites a file in place (keeping its mode) only when its content changes. `--check` writes nothing, prints each file that would change and fails when there are any. `--diff` writes nothing and prints a unified diff per changed file; like `gofmt -d` it exits 0 alone, and with `--check` it fails too. A file that cannot be read, parsed or formatted is reported on stderr (each error is printed as is, since it already names the file), makes the exit status non-zero, and the other files are still processed.
  - **Side Effects:** Rewrites files (unless `--check`/`--diff`); prints to stdout and stderr. The diff is a small LCS helper in `cmd/fmt_diff.go`, which above `maxLCSCells` (4M cells) reports the changed middle as one delete-all/insert-all hunk. Files are rewritten following `gofmt -w`: the original bytes go to a backup beside the file, the file is opened write-only without truncation (a file that cannot be opened, such as a read-only one, is never touched and the backup is removed), the new bytes are written over it and it is truncated to their length, then the backup is removed. Writing in place keeps symlinks, hard links, owner and xattrs. If bytes were written and the write then fails, the original is written back and the backup removed; the backup is kept and named in the error only when that restore also fails. A backup that cannot be made leaves the file untouched.
  - **Exceptions:** A summary error (`N file(s) could not be formatted` / `are not formatted`) after all files are processed.

- **Signature:** `sentrie init <name> [--directory .]`
  - **Behavior:** Validates the pack name, builds a default manifest, requires the target to be an **existing empty directory**, and writes the pack file with non-inline TOML tables.
  - **Side Effects:** Creates the pack manifest.
  - **Exceptions:** Invalid name; directory missing or not a directory; directory not empty; write or encode errors.

## 4. Operational Context & Gotchas
- **`bench` measures a warm, long-lived executor.** The executor's call-memoization cache and module loads persist across the timed iterations, as they do in `serve`, so a memoized call is paid once and then served from the cache. B/op and allocs/op are process-wide, so any concurrent goroutine (`ExecPolicy` runs one per exported rule) is included.
- **Statefulness:** Each command builds a fresh index and executor; nothing is shared between invocations.
- **Performance/Scale Notes:** The whole pack is parsed and indexed on every command, including `validate`. Startup cost scales with pack size, and `serve` pays it once before binding - so a large pack delays readiness but never serves a partially-indexed state.
- **Dependencies Risk:**
  - **`validate` checks the whole pack and evaluates nothing.** It used to accept a rule argument and a `--facts` flag and ignore both; both are gone (#155). cling binds positionals without complaint, so `validateCmd` rejects any non-flag word itself, as `fmtCmd` does. cling treats a single-dash word as a positional, so only a `--` prefix counts as a flag there.
  - **`exec` checks `runErr != nil` before using it** - the correct pattern that [[api.handle_decision]] is missing. Worth noting the two call sites diverge.
  - **`serve` launches `StartServer` in a goroutine and ignores its outcome.** Combined with that function's inability to report a serve error (see [[api.http]]), a listener that fails after startup leaves the process alive and silent, blocked on `ctx.Done()` with nothing listening.
  - **`serve` checks `ctx.Err()` inside the program loop but `exec` and `validate` do not**, so only `serve` is interruptible during indexing.
  - **`exec` builds `[]*ExecutorOutput{output}` without a nil guard.** `ExecRule` returns an output on success and on failure ([#143](https://github.com/sentrie-sh/sentrie/issues/143)), and a named rule's failure returns before the outputs are used.
  - **`init` requires a completely empty directory**, so it cannot add a manifest to an existing project - a notable friction point for adopting Sentrie into an existing repository.
  - **`fmt` declares its inputs with cling like the other commands, and the pattern is required for a reason.** cling reads the word after a bare flag as that flag's value, so with a default pattern `sentrie fmt --check f` would mean check=false and rewrite `.`. With a required pattern it fails instead (and `--check ./...` fails on ParseBool), so flags go after the pattern: `sentrie fmt ./... --check`. cling binds only one positional and ignores the rest, so `fmtCmd` rejects extra non-flag words itself (every flag reaches the handler as `--name=value`). A bare `--check` at the end hydrates to true and reaches the handler as `--check=`, so `fmtCmd` rejects any `--check=`/`--diff=` value other than empty, `true` or `false` (`fmt --check f sub` would otherwise read check=false and rewrite `sub`); this validates cling's output, it does not reparse. A literal `true` or `false` after a flag still binds as its value (`fmt --check false sub` rewrites `sub`), since it cannot be told apart from `--check=false`. `dir/...` on a file is an error as in `go fmt`; a bare `/...` is also an error, a sentrie choice (Go reads it as a pattern under the filesystem root).
  - **`encodePackFile` is a package-level variable specifically so tests can override it**, which is worth knowing before changing its signature.
