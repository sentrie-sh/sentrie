## Binaek Sarkar (binaek89@gmail.com)

### 22:03 - Pick up #210
- Branch cut from `209-regex-replace-first`. Both change `builtin_regex.go`, so this layer stacks on #209 to avoid a conflict. Third layer of the chain 208, 209, 210, 211, 212, 213.
- Plan: replace the unbounded global map with a `perch` LRU of a named-constant entry count, as #140 did for the executor caches.
- _Model: claude-opus-5-5_

### 22:04 - Bound the regex cache with perch
- Replaced the `sync.RWMutex` + map in `builtin_regex.go` with `perch.New[*regexp.Regexp](regexCacheEntries)`, `regexCacheEntries = 1024`. Reuses the dependency #140 already sized by entry count (`journal/140-perch-entry-count-caches.md`); no new module.
- Why 1024: matches `moduleBindingCacheEntries`. A policy's constant patterns number in the tens, so 1024 keeps every real pattern hot while capping a flood of request-derived ones at roughly a thousand small compiled programs.
- `Get` with ttl 0 caches indefinitely, LRU eviction past capacity, singleflight on concurrent misses. perch does not keep failed loads (it removes the entry after waking waiters), so invalid patterns are recompiled and keep throwing, as before.
- `context.Background()` for the load: goja callbacks carry no context and compiling is pure and non-blocking, so there is nothing to cancel. Commented at the call.
- Did not add a pattern-length limit (the issue's "consider"): RE2 compile cost is linear in pattern length, and a limit would be a new failure mode for existing policies. Left for the maintainer; noted in Review notes and the graph.
- Tests: `TestRegexCacheStaysBounded` (3x capacity distinct patterns, size stays at or under capacity; hits return the same `*Regexp`; a bad pattern is not cached) and `TestRegexCacheConcurrentVMs` (8 VMs sharing the cache). `go test -race ./runtime/js/` passes.
- Graph: `runtime.js.builtins` gains the `ext.binaek.perch` edge and a gotcha; `ext.dependencies` lists the new importer.
- _Model: claude-opus-5-5_

### 13:59 - Restack onto the rebased 209
- Carried onto 209 after it was rebased onto `main` (cfe249a). Conflicts only in the graph. In `ext.dependencies.md` I kept this layer's perch regex row and left out the puddle row, which #194 removed. In `runtime.js.builtins.md` I kept the perch row with `main`'s JWT-aware `LAYERED_ON [[constants]]` row, and both Dependencies Risk bullets (JWT clock, regex cache).
- _Model: claude-opus-5-5_
