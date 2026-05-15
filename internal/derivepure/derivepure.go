// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package derivepure lists builtins that are deterministic enough for derive bodies.
// It is shared by the runtime evaluator and the index-time derive purity checker.
package derivepure

import "slices"

// pureBuiltinNames is the single source of truth for builtins permitted in derive bodies.
// Every name must exist in runtime.Builtins.
var pureBuiltinNames = map[string]struct{}{
	"all":            {},
	"any":            {},
	"as_list":        {},
	"count":          {},
	"distinct":       {},
	"error":          {},
	"filter":         {},
	"first":          {},
	"flatten":        {},
	"flatten_deep":   {},
	"collect":        {},
	"merge":          {},
	"normalise_list": {},
	"reduce":         {},
	"now":            {},
}

// IsPureBuiltin reports whether name is a builtin allowed inside derive bodies.
func IsPureBuiltin(name string) bool {
	_, ok := pureBuiltinNames[name]
	return ok
}

// PureBuiltinNames returns a sorted copy of pure builtin names (for tests).
func PureBuiltinNames() []string {
	out := make([]string, 0, len(pureBuiltinNames))
	for k := range pureBuiltinNames {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
