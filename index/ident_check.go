// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/xerr"
)

// checkIdentifiers reports every identifier that names nothing at its site
// (#155), such as `yield undefined_thing`, which the runtime reports only when
// it evaluates it ("identifier not found"). A name is defined when the
// runtime's evalIdent can resolve it there (namesValue): a lambda parameter or
// block let in scope, pack data, a fact (by alias), a policy let, a rule, or a
// visible derive.
//
// Only identifiers that are evaluated as values are checked. A callee is
// resolved by getTarget, never evaluated, so the whole callee is skipped: an
// identifier callee is checkCallTargets', and `alias.fn` names a `use` module
// alias, which is not a value. Derive bodies are skipped: validateDerivePurity
// already rejects every name they cannot read.
//
// The runtime evaluates a rule or let in the context that references it, so a
// rule or let body also sees the lambda parameters and block lets of whatever
// site evaluates it (`let y = x + 1` read inside `(x) => { yield y }` sees the
// parameter). Such a name is not lexically in scope, but it may resolve, so a
// name bound by any lambda parameter or block let anywhere in the policy is
// never reported. Errors are reported in source order.
func (idx *Index) checkIdentifiers(ctx context.Context) error {
	type candidate struct {
		policy *Policy
		id     *ast.Identifier
	}
	var candidates []candidate
	locals := map[*Policy]map[string]struct{}{}
	bindLocal := func(p *Policy, name string) {
		if locals[p] == nil {
			locals[p] = map[string]struct{}{}
		}
		locals[p][name] = struct{}{}
	}
	notEvaluated := map[ast.Node]struct{}{}

	err := idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		if k.derive != nil || k.policy == nil {
			return
		}
		// Pre-order: a skipped node passes the mark on to its children
		// before they are visited.
		if _, ok := notEvaluated[e]; ok {
			_ = forEachExprChild(e, func(child ast.Node) error {
				notEvaluated[child] = struct{}{}
				return nil
			})
			return
		}
		switch n := e.(type) {
		case *ast.CallExpression:
			if n.Callee != nil {
				notEvaluated[n.Callee] = struct{}{}
			}
		case *ast.LambdaExpression:
			for _, p := range n.Params {
				bindLocal(k.policy, p)
			}
		case *ast.BlockExpression:
			for _, st := range n.Statements {
				if vd, ok := st.(*ast.VarDeclaration); ok {
					bindLocal(k.policy, vd.Name)
				}
			}
		case *ast.Identifier:
			if !k.namesValue(n.Value) {
				candidates = append(candidates, candidate{policy: k.policy, id: n})
			}
		}
	})
	if err != nil {
		return err
	}

	var errs []siteError
	for _, c := range candidates {
		if _, ok := locals[c.policy][c.id.Value]; ok {
			continue
		}
		errs = append(errs, siteError{at: c.id.Span(), err: undefinedIdentifierError(c.policy, c.id)})
	}
	return joinSiteErrors(errs)
}

func undefinedIdentifierError(p *Policy, id *ast.Identifier) error {
	name := id.Value
	if use, ok := p.Uses[name]; ok && use != nil {
		return fmt.Errorf("identifier '%s' not found at %s: '%s' is a module alias, not a value; call its functions as %s.<function>(...): %w", name, id.Span(), name, name, xerr.ErrIndex)
	}
	for _, alias := range slices.Sorted(maps.Keys(p.Uses)) {
		if use := p.Uses[alias]; use != nil && slices.Contains(use.Modules, name) {
			return fmt.Errorf("identifier '%s' not found at %s: it is imported from module alias '%s' and must be called as %s.%s(...): %w", name, id.Span(), alias, alias, name, xerr.ErrIndex)
		}
	}
	if _, ok := builtins.Table[name]; ok {
		return fmt.Errorf("identifier '%s' not found at %s: '%s' is a builtin, not a value; call it as %s(...): %w", name, id.Span(), name, name, xerr.ErrIndex)
	}
	return fmt.Errorf("identifier '%s' not found at %s: it is not a parameter, let, fact, data, rule or derive in scope: %w", name, id.Span(), xerr.ErrIndex)
}
