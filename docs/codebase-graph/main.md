---
id: main
type: Module / File
language: Go
file_path: main.go
tags: entrypoint, bootstrap, logging, signals, versioning
---

# Node: main (Process Entrypoint)

## 1. Architectural Role & Intent
The process bootstrap: installs signal handling, resolves build-time version information, configures structured logging, and hands control to the CLI. It is deliberately thin — every behaviour beyond process setup lives in [[cmd]] — which keeps the binary's startup path easy to audit.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `main` | `CALLS` | [[cmd]] | `cmd.Setup` builds the CLI, `cmd.Execute` runs it with `os.Args`. |
| `main` | `CALLS` | [[version]] | `GetVersionInfo` assembles the version string from ldflags and build info. |
| `main` | `DEPENDS_ON` | [[constants]] | `EnvDebug` and `EnvLogLevel` control logging. |
| `main` | `DEPENDS_ON` | `ext.google_uuid` | Generates a per-process instance identifier for log correlation. |
| `main` | `MUTATES` | `ext.os_environment` | Sets `EnvLogLevel` to `DEBUG` when debug mode is detected. |

## 3. Interface Contracts & Public Surface

- **Signature:** `main()`
  - **Behavior:** Creates a background context cancelled by `SIGINT` or `SIGKILL`, resolves the version once, installs the default logger, builds the CLI, and executes it. On error it prints `Error: <err>` to stdout and exits 1.
  - **Side Effects:** Signal handlers; global logger; process exit.
  - **Exceptions:** None — errors become exit codes.

- **Signature:** `setupDefaultLogger(version string) -> *slog.Logger`
  - **Behavior:** Resolves the level from `EnvLogLevel` (defaulting to INFO), forces DEBUG when `EnvDebug` is set, and builds a **JSON handler writing to stdout** with source locations enabled. Every line carries the version and a per-process instance UUID; debug mode additionally logs the full argument vector and executable path.
  - **Side Effects:** Sets `EnvLogLevel` in the process environment as a side effect of debug detection.
  - **Exceptions:** None.

- **Signature:** Build-time variables — `version`, `commit`, `treeState`, `date`, `builtBy`
  - **Behavior:** Overridden via `-ldflags "-X main.version=…"`. Each is applied only when non-empty, so an unstamped build falls back to whatever [[version]] derives from Go build info.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Process-global — the default `slog` logger and the signal-derived context. `versionOnce` uses `sync.OnceValue` so version resolution happens at most once.
- **Performance/Scale Notes:** Irrelevant; runs once. `AddSource: true` adds a small per-log-line cost that is paid for the process lifetime.
- **Dependencies Risk:**
  - **`os.Kill` cannot be caught.** `signal.NotifyContext(ctx, os.Interrupt, os.Kill)` registers `SIGKILL`, which the operating system does not deliver to the process — it is unstoppable by design. Only the `SIGINT` half works. Notably **`SIGTERM` is not registered**, which is the signal container orchestrators actually send, so a containerised Sentrie is killed rather than shut down cleanly. Since `serve` blocks on `ctx.Done()` to trigger its shutdown path, this means the graceful path is effectively unreachable in a container.
  - **Logs go to stdout, and so does command output.** `exec` writes its table or JSON result to stdout as well, so at INFO level or above the machine-readable output is interleaved with JSON log lines. Anything piping `sentrie exec --output json` into a parser must suppress logging first. Logs conventionally belong on stderr.
  - **Errors are printed with `fmt.Printf`, not the logger**, so the failure message is unstructured while everything around it is JSON — and the comment describing red colouring does not match the code, which applies none.
  - **Debug mode logs the full `os.Args`.** Since `exec` accepts `--facts` as a JSON string, enabling debug can write potentially sensitive fact data into the log stream.
  - **`setupDefaultLogger` mutates the process environment** to propagate the debug level rather than setting the level directly — a side effect that is invisible to any later reader of `EnvLogLevel`.
  - **`os.Exit` bypasses deferred functions**, including `stop()`. Harmless at process end, but it means no cleanup can be added to `main` via `defer`.
