// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/xerr"
)

// CheckCapabilities reports every builtin call and every `use` statement in
// the index that the operator's capabilities file does not grant, each as an
// xerr.CapabilityError anchored at the offending span. A nil grant allows
// everything. Run it after Validate, at load time.
//
// The runtime repeats both checks: builtin dispatch in runtime/eval_call.go
// and module resolution in runtime/js (which also covers require() inside
// TypeScript, which this static pass cannot see).
func (idx *Index) CheckCapabilities(ctx context.Context, grant *capabilities.Capabilities) error {
	if grant == nil {
		return nil
	}

	var denied []xerr.CapabilityError

	err := idx.forEachBuiltinCall(ctx, func(decl *builtins.Decl, call *ast.CallExpression) {
		if grant.AllowsBuiltin(decl.Name) {
			return
		}
		denied = append(denied, xerr.CapabilityError{
			At:      call.Callee.Span(),
			Message: fmt.Sprintf("builtin %q", decl.Name),
		})
	})
	if err != nil {
		return err
	}

	for _, ns := range idx.Namespaces {
		for _, policy := range ns.Policies {
			for _, use := range policy.Uses {
				if msg, ok := useDenied(grant, use); ok {
					denied = append(denied, xerr.CapabilityError{At: use.Span(), Message: msg})
				}
			}
		}
	}

	slices.SortStableFunc(denied, func(a, b xerr.CapabilityError) int {
		return cmp.Or(
			cmp.Compare(a.At.File, b.At.File),
			cmp.Compare(a.At.From.Line, b.At.From.Line),
			cmp.Compare(a.At.From.Column, b.At.From.Column),
			cmp.Compare(a.Message, b.Message),
		)
	})
	errs := make([]error, 0, len(denied))
	for _, d := range denied {
		errs = append(errs, d)
	}
	return errors.Join(errs...)
}

// useDenied classifies a use statement the way js.Registry resolves it: a
// builtin module (`from @sentrie/x`), or a pack-local file (`from "./x"` or
// `from @local/x`). Any other source is rejected by the registry on its own.
func useDenied(grant *capabilities.Capabilities, use *ast.UseStatement) (string, bool) {
	if key, ok := UseBuiltinModuleKey(use); ok {
		if grant.AllowsModule(key) {
			return "", false
		}
		return fmt.Sprintf("module %q", key), true
	}
	if isLocalUse(use) && !grant.AllowsLocalModules() {
		from := use.RelativeFrom
		if from == "" {
			from = "@" + joinLibFrom(use.LibFrom)
		}
		return fmt.Sprintf("local module %q", from), true
	}
	return "", false
}

func isLocalUse(use *ast.UseStatement) bool {
	if use.RelativeFrom != "" {
		return true
	}
	return len(use.LibFrom) > 0 && use.LibFrom[0] == "local"
}

func joinLibFrom(parts []string) string {
	return strings.Join(parts, "/")
}
