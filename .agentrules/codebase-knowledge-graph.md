# Codebase Knowledge Graph

Read [`docs/codebase-graph/overview.md`](../docs/codebase-graph/overview.md) before exploring the codebase. It is a knowledge graph of every package, file, and boundary in the repository: what each one is for, what it depends on, its public contract, and its known defects and gotchas. Start at the overview and follow the wiki-links to the areas you need. Use it instead of grepping blindly.

The graph is only useful while it is true, so keeping it current is a standing convention, not a one-off task.

## When to update it

Update the affected node files in the **same change** as the code whenever you:

- add, remove, or rename a package or a file significant enough to have its own node;
- change a public contract, a signature, a side effect, or a failure mode a node documents;
- add, remove, or change a dependency edge (an import, a call across packages, a layering relationship);
- add, remove, or bump a module in `go.mod` (`docs/codebase-graph/ext.dependencies.md` catalogues the manifest).

A pure internal refactor that changes no documented contract or edge needs no graph change. When in doubt, read the affected node and check whether anything it asserts is now false.

## How the graph is shaped

`overview.md` is the normative schema and the node catalog. In short:

- **Five front-matter keys on every node:** `id` (equal to the filename stem), `type`, `language`, `file_path`, `tags`.
- **Five node types:** `System / Package`, `Class`, `Function / Endpoint`, `Module / File`, `Infrastructure`.
- **Four body sections:** 1. Architectural Role & Intent, 2. Graph Edges, 3. Interface Contracts & Public Surface, 4. Operational Context & Gotchas.
- **Edges live only in the section-2 table** and use a closed set of seven predicates: `CALLS`, `IMPORTS`, `LAYERED_ON`, `DEPENDS_ON`, `READS_FROM`, `MUTATES`, `INHERITS_FROM`. `LAYERED_ON` is for package-to-package edges only.
- **Ids:** a package node's id is its path with slashes turned to dots (`runtime/js/` is `runtime.js`). Internal and `infra.*` nodes are `[[wiki-link]]`s. Third-party (`ext.<org>.<module>`) and standard-library (`std.<import/path>`) references are backticked leaves, never wiki-links.
- **Edge locality and no orphans:** every edge has its own node at one end, and every node has at least one inbound link.
- **A new node goes into the catalog in `overview.md`**, or the validator fails.
- **No source code in nodes.** Record relationships, contracts, and intent. Copied code goes stale silently.

## Enforcement

`make docs-graph` runs `go test -count=1 ./internal/docsgraph/`. The `docs-graph` CI job runs the same thing, and `go test ./...` includes it. It checks structure, not truth: front-matter shape, id and filename agreement, `file_path` resolution, the closed predicate set, link resolution, locality, orphans, and catalog completeness. Nothing detects prose drift yet ([#121](https://github.com/sentrie-sh/sentrie/issues/121)), so the accuracy of every claim is still on the author.

## Defects found while indexing

The graph records relationships and intent, not fixes. If you find a bug while documenting a node, file it as its own GitHub issue and reference the issue number in that node's Operational Context. Don't fix it in a documentation change.

**Why:** Architectural facts are stable, and rediscovering them from source on every task is slow and error-prone. A graph that is confidently wrong is worse than none, which is why staleness is something to guard against.

**How to apply:** Read the overview and the relevant nodes before reading sources. Before you finish a change, check every node whose package you touched, and run `make docs-graph`.

Related: [Token efficiency](token-efficiency.md), [Go project layout](go-project-layout.md), [Build and test](build-and-test.md), [Journal convention](journal-convention.md)
