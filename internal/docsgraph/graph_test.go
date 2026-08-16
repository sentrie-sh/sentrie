// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package docsgraph

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	graphDir = "../../docs/codebase-graph"
	repoRoot = "../.."

	// overviewNode is the entrypoint matrix. It carries the catalog and the
	// normative schema instead of the four-section node structure. Named
	// overview so the Go package index/ can own index.md.
	overviewNode = "overview"

	// externalFilePath marks a node with no source file. Only Infrastructure
	// nodes, which describe resources outside the process, may use it.
	externalFilePath = "(external)"
)

// catalogNodes describe the system as a whole rather than one participant, so
// they legitimately declare edges between two other nodes.
var catalogNodes = map[string]bool{
	overviewNode:       true,
	"ext.dependencies": true,
}

var (
	requiredFrontMatter = []string{"id", "type", "language", "file_path", "tags"}

	nodeTypes = map[string]bool{
		"System / Package":    true,
		"Class":               true,
		"Function / Endpoint": true,
		"Module / File":       true,
		"Infrastructure":      true,
	}

	predicates = map[string]bool{
		"CALLS":         true,
		"IMPORTS":       true,
		"LAYERED_ON":    true,
		"DEPENDS_ON":    true,
		"READS_FROM":    true,
		"MUTATES":       true,
		"INHERITS_FROM": true,
	}

	packageType = "System / Package"

	requiredSections = []string{
		"## 1. Architectural Role & Intent",
		"## 2. Graph Edges (Strict Relational Data)",
		"## 3. Interface Contracts & Public Surface",
		"## 4. Operational Context & Gotchas",
	}

	wikiLinkRe   = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	codeSpanRe   = regexp.MustCompile("`[^`]*`")
	fenceRe      = regexp.MustCompile("(?s)```.*?```")
	tableCellRe  = regexp.MustCompile("`([^`]+)`")
	leafRe       = regexp.MustCompile(`^(ext|std)\.`)
	separatorRow = regexp.MustCompile(`^[|\s:-]+$`)
)

type edge struct {
	subject   string
	predicate string
	target    string

	// line is relative to the start of the section-2 table, which is enough to
	// locate a row given that subtests are named for the node.
	line int
}

type node struct {
	id       string
	typ      string
	filePath string
	file     string
	body     string
	edges    []edge
}

// loadGraph parses every node file in the graph directory. Parse failures are
// fatal: a file that cannot be read as a node cannot be checked as one.
func loadGraph(t *testing.T) map[string]*node {
	t.Helper()

	entries, err := filepath.Glob(filepath.Join(graphDir, "*.md"))
	require.NoError(t, err)
	require.NotEmpty(t, entries, "no node files found in %s", graphDir)

	nodes := make(map[string]*node, len(entries))
	for _, path := range entries {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)

		n := parseNode(t, path, string(raw))
		require.NotContains(t, nodes, n.id, "duplicate node id %q", n.id)
		nodes[n.id] = n
	}
	return nodes
}

func parseNode(t *testing.T, path, raw string) *node {
	t.Helper()

	name := filepath.Base(path)
	front, body := splitFrontMatter(t, name, raw)

	n := &node{
		id:       front["id"],
		typ:      front["type"],
		filePath: front["file_path"],
		file:     name,
		body:     body,
	}

	for _, key := range requiredFrontMatter {
		require.NotEmpty(t, front[key], "%s: front-matter key %q is missing or empty", name, key)
	}
	n.edges = parseEdges(body)
	return n
}

func splitFrontMatter(t *testing.T, name, raw string) (map[string]string, string) {
	t.Helper()

	rest, found := strings.CutPrefix(raw, "---\n")
	require.True(t, found, "%s: file does not open with a front-matter fence", name)

	block, body, found := strings.Cut(rest, "\n---\n")
	require.True(t, found, "%s: front-matter fence is not closed", name)

	front := map[string]string{}
	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		front[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return front, body
}

// parseEdges reads the section-2 table. Only that table declares edges; links
// anywhere else in a node file are navigational.
func parseEdges(body string) []edge {
	_, section, found := strings.Cut(body, "## 2. Graph Edges")
	if !found {
		return nil
	}
	if before, _, ok := strings.Cut(section, "\n## "); ok {
		section = before
	}

	var edges []edge
	for i, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || separatorRow.MatchString(line) {
			continue
		}

		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 3 || strings.Contains(cells[0], "Source (Subject)") {
			continue
		}

		subject := soleToken(cells[0])
		predicate := soleToken(cells[1])
		for _, target := range tokens(cells[2]) {
			edges = append(edges, edge{
				subject:   subject,
				predicate: predicate,
				target:    target,
				line:      i,
			})
		}
	}
	return edges
}

// tokens pulls every id out of a table cell. A cell holds either wiki-linked
// node references or backticked leaves, and may hold several of the latter.
func tokens(cell string) []string {
	var out []string
	for _, m := range wikiLinkRe.FindAllStringSubmatch(cell, -1) {
		out = append(out, strings.TrimSpace(m[1]))
	}
	stripped := wikiLinkRe.ReplaceAllString(cell, "")
	for _, m := range tableCellRe.FindAllStringSubmatch(stripped, -1) {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return out
}

func soleToken(cell string) string {
	got := tokens(cell)
	if len(got) != 1 {
		return ""
	}
	return got[0]
}

// stripCode removes fenced blocks and inline spans. The extractor contract in
// overview.md says a wiki-link inside code is literal text, not an edge — api.net
// documents the string "[[::1]]:7529", which is not a reference to a node.
func stripCode(body string) string {
	return codeSpanRe.ReplaceAllString(fenceRe.ReplaceAllString(body, ""), "")
}

func isLeaf(id string) bool { return leafRe.MatchString(id) }

// involves reports whether other is the node itself or one of its namespace
// members, so a package node may summarise edges belonging to its own files.
func involves(id, other string) bool {
	return other == id || strings.HasPrefix(other, id+".")
}

// expectedPackageID is the node id implied by a directory file_path:
// index/ → index, runtime/js/ → runtime.js. Globs and non-directory paths
// are not convertible and return ok=false.
func expectedPackageID(filePath string) (string, bool) {
	if strings.ContainsAny(filePath, "*,") {
		return "", false
	}
	if !strings.HasSuffix(filePath, "/") {
		return "", false
	}
	p := strings.Trim(filePath, "/")
	if p == "" || p == "." {
		return "", false
	}
	return strings.ReplaceAll(p, "/", "."), true
}

func TestFrontMatterIsWellFormed(t *testing.T) {
	for id, n := range loadGraph(t) {
		t.Run(id, func(t *testing.T) {
			stem := strings.TrimSuffix(n.file, ".md")
			require.Equal(t, stem, n.id, "id must equal the filename stem")
			require.True(t, nodeTypes[n.typ], "unknown node type %q", n.typ)

			if n.typ == packageType {
				if want, ok := expectedPackageID(n.filePath); ok {
					require.Equal(t, want, n.id,
						"package id must equal file_path with slashes as dots")
				}
			}

			if n.filePath == externalFilePath {
				require.Equal(t, "Infrastructure", n.typ,
					"only Infrastructure nodes may use %s as file_path", externalFilePath)
				return
			}

			// A node may cover several files, in which case file_path is a
			// comma-separated list of paths or globs. Every element must match.
			for _, part := range strings.Split(n.filePath, ",") {
				pattern := filepath.Join(repoRoot, strings.TrimSpace(part))
				matches, err := filepath.Glob(pattern)
				require.NoError(t, err, "file_path %q is not a valid pattern", part)
				require.NotEmpty(t, matches,
					"file_path %q does not resolve in the repository", strings.TrimSpace(part))
			}
		})
	}
}

func TestNodesHaveTheFourSections(t *testing.T) {
	for id, n := range loadGraph(t) {
		if id == overviewNode {
			continue
		}
		t.Run(id, func(t *testing.T) {
			for _, section := range requiredSections {
				require.Contains(t, n.body, section, "missing section")
			}
		})
	}
}

func TestEdgesUseTheClosedSchema(t *testing.T) {
	nodes := loadGraph(t)

	for id, n := range nodes {
		t.Run(id, func(t *testing.T) {
			for _, e := range n.edges {
				require.True(t, predicates[e.predicate],
					"line %d: predicate %q is not in the closed set", e.line, e.predicate)

				_, subjectIsNode := nodes[e.subject]
				require.True(t, subjectIsNode || isLeaf(e.subject),
					"line %d: subject %q resolves to nothing", e.line, e.subject)

				_, targetIsNode := nodes[e.target]
				require.True(t, targetIsNode || isLeaf(e.target),
					"line %d: target %q resolves to nothing", e.line, e.target)

				assert.NotEqual(t, e.subject, e.target,
					"line %d: self-edge on %q", e.line, e.subject)

				switch e.predicate {
				case "IMPORTS":
					require.False(t, targetIsNode,
						"line %d: IMPORTS is reserved for ext.*/std.* leaves, got %q", e.line, e.target)
				case "DEPENDS_ON":
					require.True(t, targetIsNode,
						"line %d: DEPENDS_ON must target a node; use IMPORTS for the leaf %q", e.line, e.target)
				case "LAYERED_ON":
					// Layering is a property of subsystems. Keeping it between
					// package nodes is what stops it re-absorbing DEPENDS_ON and
					// keeps the layer map small enough to read.
					require.Equal(t, packageType, nodes[e.subject].typ,
						"line %d: LAYERED_ON subject %q is not a package", e.line, e.subject)
					require.True(t, targetIsNode && nodes[e.target].typ == packageType,
						"line %d: LAYERED_ON target %q is not a package", e.line, e.target)
				}
			}
		})
	}
}

// TestEdgesAreLocal keeps each node's table about that node. Hub nodes such as
// ast have over a hundred inbound edges and deliberately summarise rather than
// enumerate them, so full reciprocity is not the invariant — locality is. An
// edge declared somewhere that is party to neither end is unreachable from both.
func TestEdgesAreLocal(t *testing.T) {
	nodes := loadGraph(t)

	for id, n := range nodes {
		if catalogNodes[id] {
			continue
		}
		t.Run(id, func(t *testing.T) {
			for _, e := range n.edges {
				assert.True(t, involves(id, e.subject) || involves(id, e.target),
					"line %d: edge %s -> %s is party to neither end of this node",
					e.line, e.subject, e.target)
			}
		})
	}
}

// TestNoOrphanNodes ensures every node is reachable. A node nothing links to is
// invisible to an agent traversing from the entrypoint.
func TestNoOrphanNodes(t *testing.T) {
	nodes := loadGraph(t)

	inbound := map[string]bool{}
	for id, n := range nodes {
		for _, m := range wikiLinkRe.FindAllStringSubmatch(stripCode(n.body), -1) {
			if target := strings.TrimSpace(m[1]); target != id {
				inbound[target] = true
			}
		}
	}

	for id := range nodes {
		if id == overviewNode {
			continue
		}
		require.True(t, inbound[id], "node %q has no inbound link from any other node", id)
	}
}

// TestWikiLinksResolve covers navigational links too, not just edges. A dangling
// link in prose is as much a dead end for an agent as a dangling edge.
func TestWikiLinksResolve(t *testing.T) {
	nodes := loadGraph(t)

	for id, n := range nodes {
		t.Run(id, func(t *testing.T) {
			for _, m := range wikiLinkRe.FindAllStringSubmatch(stripCode(n.body), -1) {
				target := strings.TrimSpace(m[1])
				require.Contains(t, nodes, target, "wiki-link [[%s]] resolves to nothing", target)
			}
		})
	}
}

func TestEveryNodeIsInTheCatalog(t *testing.T) {
	nodes := loadGraph(t)

	overview, ok := nodes[overviewNode]
	require.True(t, ok, "%s.md is missing", overviewNode)

	_, catalog, found := strings.Cut(overview.body, "## Node Catalog")
	require.True(t, found, "%s.md has no catalog section", overviewNode)

	listed := map[string]bool{}
	for _, m := range wikiLinkRe.FindAllStringSubmatch(catalog, -1) {
		listed[strings.TrimSpace(m[1])] = true
	}

	for id := range nodes {
		if id == overviewNode {
			continue
		}
		require.True(t, listed[id], "node %q is not listed in the %s.md catalog", id, overviewNode)
	}
}

func TestInvolvesTreatsPackageAsNamespaceParent(t *testing.T) {
	assert.True(t, involves("index", "index"))
	assert.True(t, involves("index", "index.derive"))
	assert.True(t, involves("index", "index.rule"))
	assert.True(t, involves("runtime", "runtime.eval"))
	assert.True(t, involves("runtime.js", "runtime.js.registry"))
	assert.False(t, involves("overview", "index.policy"))
	assert.False(t, involves("index.package", "index.derive"))
	assert.False(t, involves("index", "indexer"))
	assert.False(t, involves("runtime", "runtimejs"))
}
