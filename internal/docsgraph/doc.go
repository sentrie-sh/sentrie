// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package docsgraph holds the conformance check for the codebase knowledge
// graph under docs/codebase-graph.
//
// The graph is consumed by agents, so its invariants — front-matter shape, a
// closed predicate set, resolvable links, and reciprocal edges — are asserted
// mechanically rather than by review. The normative schema lives in
// docs/codebase-graph/index.md; this package is its enforcement.
package docsgraph
