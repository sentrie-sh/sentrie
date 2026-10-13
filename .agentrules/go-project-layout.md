# Go Project Layout

Sentrie is one Go module (`github.com/sentrie-sh/sentrie`) that builds one binary, `sentrie`. Its packages sit flat at the root, one per pipeline stage or subsystem. Match this layout when adding packages; don't introduce a competing structure.

## The shape

- **`main.go` lives at the repository root**, next to `go.mod` and `go.sum`. It holds the link-time build variables (`version`, `commit`, `treeState`, `date`, `builtBy`), wires them into `version.GetVersionInfo`, sets up the default `slog` logger, and hands off to `cmd.Setup` and `cmd.Execute`.
- **The `cmd` package builds the command tree.** `cmd.Setup` assembles the CLI. Each subcommand (`init`, `exec`, `serve`, `validate`) lives in its own `cmd/<name>.go` with an `add<Name>Cmd` function.
- **The CLI library is `github.com/binaek/cling`, not Cobra.**
- **Packages are flat at the root, one per stage.** The pipeline is `lexer` and `tokens`, then `parser` into `ast`, then `index` (with `dag` for cycle detection), then `runtime` (with `runtime/js` for JavaScript and TypeScript modules, and `runtime/trace` for decision trees). Around it sit `box` and `trinary` (the value model), `constraints`, `builtins`, `loader` and `pack` (pack discovery and manifests), `api` (the HTTP server), `grammar` (the reference spec), `constants`, `version` and `xerr`. Read the graph's System Overview before adding one ([Codebase knowledge graph](codebase-knowledge-graph.md)).
- **`internal/` holds repository tooling, not product code.** Today that is only `internal/docsgraph`, the knowledge-graph validator. Put a package there only when nothing outside the module should import it.
- **Fixtures and examples are not packages.** `lang_test/` holds the numbered `.sentrie` language fixtures and the single test that runs them. `example_pack/` is a sample policy pack.
- **`sandbox.go` is a `//go:build sandbox` scratch file** at the root ([Sandbox programs](sandbox.md)).

## Tests

- Use `testify` suites, following each package's existing suite ([Go test conventions](go-test-conventions.md)).

Related: [Build and test](build-and-test.md), [Go test conventions](go-test-conventions.md), [Sandbox programs](sandbox.md)
