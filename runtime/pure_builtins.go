// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

// derivePureBuiltins is the whitelist of builtins permitted inside derive bodies.
// Keep in sync with Builtins in builtins.go: every name here must exist in Builtins.
var derivePureBuiltins = map[string]struct{}{
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

func isBuiltinAllowedInDerive(name string) bool {
	_, ok := derivePureBuiltins[name]
	return ok
}
