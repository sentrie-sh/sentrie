# Trace Data Flow When Auditing for Races or Missing Guards

When auditing for races or missing guards, always trace the data flow.

**Why:** Race conditions and missing-guard bugs are easy to miss (or false-positive) when reasoned about abstractly. Tracing where data actually originates, mutates, and is read concurrently is what surfaces real issues.

**How to apply:** During any review or audit involving concurrency, locking, or state guards, don't just pattern-match for missing locks or checks. Follow the actual read and write paths of the relevant data across goroutines, requests, and callbacks to confirm whether a race or missing guard is real. In Sentrie the concurrency lives in `runtime.executor` and `runtime.exec_ctx` (see their nodes in `docs/codebase-graph/`), so start from their section 4 gotchas. Confirm a suspected race with `go test -race` on the narrowest package that reproduces it.

Related: [Parallelize with subagents](parallelize-with-subagents.md), [Codebase knowledge graph](codebase-knowledge-graph.md)
