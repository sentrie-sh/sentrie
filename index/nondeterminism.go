// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/tokens"
)

// Warning is a non-fatal, span-anchored diagnostic. Validation still succeeds.
type Warning struct {
	At      tokens.Range
	Message string
}

func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.At.String(), w.Message)
}

// UseBuiltinModuleKey returns the registry key ("@sentrie/<name>") of the
// builtin module a use statement imports, mirroring how js.Registry resolves
// `use ... from @sentrie/<name>`. It reports false for local modules and for
// every other source.
func UseBuiltinModuleKey(use *ast.UseStatement) (string, bool) {
	if use == nil || len(use.LibFrom) < 2 || use.LibFrom[0] != constants.APPNAME {
		return "", false
	}
	return "@" + constants.APPNAME + "/" + filepath.ToSlash(filepath.Join(use.LibFrom[1:]...)), true
}

// NondeterminismWarnings reports every use of a nondeterministic source: each
// call to a builtin marked builtins.Decl.Nondeterministic, and each name a
// `use` statement imports from a builtin module that isNondeterministicExport
// reports as nondeterministic. The index does not import the JavaScript
// registry, so the caller supplies that lookup (js.IsNondeterministicExport);
// a nil lookup skips module imports.
//
// Warnings are sorted by file and position so output is stable.
func (idx *Index) NondeterminismWarnings(ctx context.Context, isNondeterministicExport func(moduleKey, export string) bool) ([]Warning, error) {
	var out []Warning

	err := idx.forEachBuiltinCall(ctx, func(_ *kindCheckCtx, decl *builtins.Decl, call *ast.CallExpression) {
		if !decl.Nondeterministic {
			return
		}
		out = append(out, Warning{
			At:      call.Span(),
			Message: fmt.Sprintf("builtin %q is nondeterministic: its result can differ between executions", decl.Name),
		})
	})
	if err != nil {
		return nil, err
	}

	if isNondeterministicExport != nil {
		for _, ns := range idx.Namespaces {
			for _, policy := range ns.Policies {
				for _, use := range policy.Uses {
					key, ok := UseBuiltinModuleKey(use)
					if !ok {
						continue
					}
					for _, name := range use.Modules {
						if isNondeterministicExport(key, name) {
							out = append(out, Warning{
								At:      use.Span(),
								Message: fmt.Sprintf("%s.%s is nondeterministic: its result can differ between executions", key, name),
							})
						}
					}
				}
			}
		}
	}

	slices.SortStableFunc(out, func(a, b Warning) int {
		return cmp.Or(
			cmp.Compare(a.At.File, b.At.File),
			cmp.Compare(a.At.From.Line, b.At.From.Line),
			cmp.Compare(a.At.From.Column, b.At.From.Column),
			cmp.Compare(a.Message, b.Message),
		)
	})
	return out, nil
}
