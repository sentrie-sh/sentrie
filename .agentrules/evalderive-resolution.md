# `evalDerive` Resolution

When `ec.evalDerive != nil`, the execution context is inside a derive body. Every resolution path in `runtime/eval_call.go` and `runtime/eval_ident.go` must **explicitly** allow or deny that case.

**Why:** Falling through to an unrestricted path (for example TypeScript module invocation) without a derive check is a bug. Derive evaluation must stay deterministic and isolated from rules, facts, and impure builtins.

**How to apply:** When adding a new branch that resolves a callee, identifier, or builtin, decide how it behaves under `evalDerive`, and document it in code if non-obvious. The static side (`docs/codebase-graph/index.derive_purity.md`) mirrors these runtime gates, so when you change one, check the other.

Related: [Codebase knowledge graph](codebase-knowledge-graph.md), [Trace data flow for race/guard audits](trace-data-flow-for-races.md)
