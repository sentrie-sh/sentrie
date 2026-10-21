// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"slices"
	"strings"

	"github.com/sentrie-sh/sentrie/constants"
)

// nondeterministicExports lists, per builtin module key, the exports whose
// result can differ between two executions given the same arguments because
// they read a clock or a random source. It is the JavaScript counterpart of
// builtins.Decl.Nondeterministic: `sentrie validate` warns on every `use` that
// imports one of these.
//
// Any export added to a builtin module that reads the clock or a random source
// must be listed here. That includes exports that compare against the
// execution clock, such as JWT `exp`/`nbf` checks (#197, #202) once they land. TestNondeterministicExportsExist in runtime checks that
// every listed module is registered and every listed export exists.
var nondeterministicExports = map[string][]string{
	// Date.now and Math.random, re-exported from the JavaScript globals.
	BuiltinModuleKey("js"): {"now", "random"},
	// Pinned to the execution start, but differs between executions.
	BuiltinModuleKey("time"): {"now"},
	// Fresh random (v4) or time-ordered (v6, v7) identifiers on every call.
	BuiltinModuleKey("uuid"): {"v4", "v6", "v7"},
}

// BuiltinModuleKey returns the registry key of the builtin module name, for
// example "@sentrie/uuid" for "uuid". It is the key RegisterGoBuiltin and
// RegisterTSBuiltin register under.
func BuiltinModuleKey(name string) string {
	return "@" + constants.APPNAME + "/" + name
}

// IsBuiltinModuleKey reports whether key names a builtin module ("@sentrie/...").
func IsBuiltinModuleKey(key string) bool {
	return strings.HasPrefix(key, "@"+constants.APPNAME+"/")
}

// IsNondeterministicExport reports whether export of the builtin module
// moduleKey (for example "@sentrie/uuid", "v4") is nondeterministic.
func IsNondeterministicExport(moduleKey, export string) bool {
	return slices.Contains(nondeterministicExports[moduleKey], export)
}

// NondeterministicExports returns a sorted copy of the nondeterministic
// exports of the builtin module moduleKey, or nil when it has none.
func NondeterministicExports(moduleKey string) []string {
	exports := nondeterministicExports[moduleKey]
	if len(exports) == 0 {
		return nil
	}
	out := slices.Clone(exports)
	slices.Sort(out)
	return out
}

// NondeterministicModules returns the sorted keys of the builtin modules that
// have at least one nondeterministic export.
func NondeterministicModules() []string {
	out := make([]string, 0, len(nondeterministicExports))
	for key := range nondeterministicExports {
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}
