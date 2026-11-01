# 174-1-bench

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-01 07:02 - Pick up issue #174
- Picked up as part of a wave-2 batch of two issues (174 first slice, 172), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no way to measure policy evaluation cost or to catch performance regressions such as #140.
- Decision: this first slice adds `sentrie bench <policy/rule>` (facts input, `-count`, `-benchtime`, Go benchmark output so `benchstat` works) and Go benchmarks for the lexer, parser, index, executor and startup on `example_pack`. `exec --profile`, `exec --explain` and performance CI are left to later slices.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-11-01 07:26 - Implement bench command and Go benchmarks
- Added `sentrie bench <target>` (`cmd/bench.go`) and Go benchmarks for the lexer, parser, index, executor and startup on `example_pack`.
- Decision: `bench` has its own measuring loop instead of `testing.Benchmark`. `testing.Benchmark` needs `testing.Init` and reads global test flags, which a production binary should not register. The loop mirrors `go test`: a duration grows N (capped at 100x growth per step, 1e9 total) and reports the last run; `Nx` runs exactly N. B/op and allocs/op are `runtime.MemStats` TotalAlloc and Mallocs deltas taken outside the timed window.
- Decision: flags are `--count` and `--benchtime` (double dash), because cling only parses the `--` form; single-dash `-count` would be read as a positional argument.
- Decision: the benchmark name carries the `-GOMAXPROCS` suffix (omitted when it is 1), as `go test` prints, and whitespace in the target becomes `_`. `goos:` and `goarch:` configuration lines are printed first; benchstat accepts them.
- Decision: one untimed warm-up evaluation runs before measuring; it also makes a missing target or a failing evaluation (for example a required fact missing) fail with a non-zero exit before any result line prints.
- Decision: the load, index and validate steps are a new small helper (`loadBenchIndex`) rather than a refactor of `exec`, to keep `exec` behaviour (and its error ordering) untouched; the repetition matches what the other commands already do.
- Decision: the `target` argument is marked `Required()` so a missing target is a clear cling error (without it cling hands the command a default of "0", which is what `exec` currently reports as `namespace: 0: not found`; left as is, out of scope).
- Decision: the index benchmark parses programs inside `b.StopTimer`/`b.StartTimer` so each iteration indexes a fresh AST rather than assuming indexing leaves it untouched.
- Problem: `make build` installs to `~/.local/bin` by default, so one build here overwrote that binary with this branch's build; later builds went to the scratch dir.
- Graph: updated `docs/codebase-graph/cmd.md` (five commands, bench contract, memoization gotcha). `make docs-graph` and lint pass; `go test -run x -bench . -benchtime 1x ./...` runs in about 7s.
- _Model: claude-sonnet-5-5, role: implementer subagent_

### 2026-11-01 07:28 - Test and CI pass for the bench slice
- Added `cmd/bench_edge_test.go`: tab-separated field layout and name shape of result lines, fact file alone supplying facts, 14 failure inputs (unknown rule, policy and namespace, missing required fact, non-object facts JSON, bad fact file, `--benchtime` overflow, scientific, bare number and negative forms, negative `--count`) that must print nothing to stdout (not even goos/goarch), `parseBenchtime` edges, duration mode stopping on a later error, and allocation deltas per op.
- Bite proof: removing `cmd/bench.go` fails to compile the tests (undefined `parseBenchtime`, `runBench`, `addBenchCmd`); restoring `main`'s `cmd/cmd.go` (no `addBenchCmd`) fails 35 bench tests.
- Gate: `make test`, lint (0 issues), gofmt, `make docs-graph`, race on cmd, lexer, parser, index, em dash scan all clean; `go test -run x -bench . -benchtime 1x` over lexer, parser, index, runtime takes about 12s including build.
- No bugs found.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-11-01 07:29 - Address review findings
- Added `bench` to the capabilities edges in the `cmd` and `capabilities` graph nodes, since `bench` builds its executor through `newGrantedExecutor` like `exec`.
- Reworded the `loadBenchIndex` comment: only `bench` calls it.
- Corrected the failure wording in the `cmd` node, PR description and this journal: only a warm-up failure prints nothing; a failure in a later `--count` run stops after the lines already printed. Buffering the lines was not worth it, since each line is a complete result.
- The executor benchmark now runs one untimed `ExecRule` before its sub-benchmarks, so a `-benchtime 1x` run of `rule` does not include first-call setup.
- Declined: a README example for `bench`; the README lists no subcommands and the website docs cover the CLI.
- _Model: claude-opus-5-5, role: parent session_
