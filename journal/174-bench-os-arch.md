# 174-bench-os-arch

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-13 15:55 - Pick up the bench os/arch rename
- Review of the v0.0.5 website release PR (sentrie-sh/website#73, thread on `cli-reference/bench.md` line 144) asked for `os` and `arch` in place of `goos` and `goarch` in the bench output. The user approved fixing it in the engine first, so the docs can follow.
- Scope: `sentrie bench` prints `os: <os>` and `arch: <arch>` configuration lines instead of `goos:` / `goarch:`. Result lines and their `BenchmarkExec/...` names are unchanged.
- Supersedes the `174-1-bench` decision to print `goos:` and `goarch:` lines. `benchstat` treats any `key: value` line before the results as configuration, so comparison still works.
- Branch `174-bench-os-arch` off `origin/main` at 74ead0f. Updates #174 (the issue stays open for its later slices).
