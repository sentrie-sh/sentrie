// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

// detectSelfReference reports a rule or let whose own value reads itself,
// such as `rule r = r` or `let x = x + 1` (#155). The runtime fails on the
// first evaluation with "'t/p/r' references itself".
//
// The identifier graph in detectReferenceCycle ignores scope, so its self
// edges cannot be trusted: a block let, lambda parameter or fact with the
// rule's name would read as a self-reference. This check resolves each
// identifier the way the runtime's evalIdent does instead: a lambda parameter
// of any enclosing lambda, then pack data and facts, then the nearest let
// (block lets inside out, then the policy's), then a rule. A self-reference is
// an identifier that resolves to the rule or let whose value it is in, except
// a let whose name is bound anywhere else in the policy, which the runtime may
// resolve to a cached value of that other binding (selfRefBinders).
//
// A lambda body is not part of the value it is written in: a let bound to a
// lambda may call itself (#161), and a lambda passed to a call may never run.
// So identifiers inside a lambda body are not reported, which can miss a real
// recursion through a callback but never reports one that is not there.
// An export attachment is walked with its rule as owner, since the runtime
// evaluates it with the rule still on the reference stack.
// Callees are not evaluated as identifiers (getTarget), and contributions
// reading their own rule are reported by detectContributionSelfReference.
// Derive bodies cannot read rules or lets (validateDerivePurity) and are not
// walked. Errors are reported in source order.
func (idx *Index) detectSelfReference(ctx context.Context) error {
	binders, err := idx.selfRefBinders(ctx)
	if err != nil {
		return err
	}
	var errs []siteError
	for _, ns := range idx.Namespaces {
		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			w := &selfRefWalk{policy: policy, errs: &errs, binders: binders[policy]}
			for _, rule := range policy.Rules {
				for _, e := range []ast.Expression{rule.Default, rule.When, rule.Body} {
					if err := w.walk(e, nil, rule); err != nil {
						return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
					}
				}
				for _, c := range rule.Contributions {
					for _, e := range []ast.Expression{c.When, c.Body} {
						if err := w.walk(e, nil, nil); err != nil {
							return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
						}
					}
				}
			}
			for _, let := range policy.Lets {
				if err := w.walk(let.Value, nil, selfRefOwnerOfLet(let)); err != nil {
					return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
				}
			}
			for _, fact := range policy.Facts {
				if err := w.walk(fact.Default, nil, nil); err != nil {
					return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
				}
			}
			for _, export := range policy.RuleExports {
				if export == nil {
					continue
				}
				// execRule evaluates attachments while the rule is still on
				// the reference stack, so an attachment reading its rule
				// fails with "references itself" just like the rule's body.
				// The owner is the exported rule, or nil if there is none.
				var owner any
				if r, ok := policy.Rules[export.RuleName]; ok && r != nil {
					owner = r
				}
				for _, attachment := range export.Attachments {
					if attachment == nil {
						continue
					}
					if err := w.walk(attachment.Value, nil, owner); err != nil {
						return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
					}
				}
			}
		}
	}
	return joinSiteErrors(errs)
}

// selfRefOwnerOfLet is the owner a let's value is walked with: the let
// itself, or nil when the value is a lambda, which is not evaluated when the
// let is (#161).
func selfRefOwnerOfLet(let *ast.VarDeclaration) any {
	if _, ok := let.Value.(*ast.LambdaExpression); ok {
		return nil
	}
	return let
}

// selfRefFrame is one lexical scope: the parameters of a lambda or the lets
// of a block.
type selfRefFrame struct {
	parent *selfRefFrame
	params map[string]struct{}
	lets   map[string]*ast.VarDeclaration
}

type selfRefWalk struct {
	policy  *Policy
	errs    *[]siteError
	binders *selfRefPolicyBinders
}

// selfRefPolicyBinders is every name a policy binds outside its policy lets,
// rules, facts and data: each block let declaration, and each lambda
// parameter name.
type selfRefPolicyBinders struct {
	blockLets map[string]map[*ast.VarDeclaration]struct{}
	params    map[string]struct{}
}

// selfRefBinders collects, per policy, the block lets and lambda parameters
// of every expression site (forEachScopedExpr), so a let's self-reference can
// be skipped when its name is bound anywhere else in the policy
// (selfRefWalk.boundElsewhere). Derive bodies are detached and not counted.
func (idx *Index) selfRefBinders(ctx context.Context) (map[*Policy]*selfRefPolicyBinders, error) {
	out := map[*Policy]*selfRefPolicyBinders{}
	err := idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		if k.derive != nil || k.policy == nil {
			return
		}
		b := out[k.policy]
		if b == nil {
			b = &selfRefPolicyBinders{blockLets: map[string]map[*ast.VarDeclaration]struct{}{}, params: map[string]struct{}{}}
			out[k.policy] = b
		}
		switch n := e.(type) {
		case *ast.LambdaExpression:
			for _, p := range n.Params {
				b.params[p] = struct{}{}
			}
		case *ast.BlockExpression:
			for _, st := range n.Statements {
				if vd, ok := st.(*ast.VarDeclaration); ok {
					if b.blockLets[vd.Name] == nil {
						b.blockLets[vd.Name] = map[*ast.VarDeclaration]struct{}{}
					}
					b.blockLets[vd.Name][vd] = struct{}{}
				}
			}
		}
	})
	return out, err
}

// resolve returns what name refers to at a site with the given scope: a
// *ast.VarDeclaration for a let, a *Rule for a rule, or nil for anything
// else (a parameter, pack data, a fact, a derive, or nothing). The order is
// evalIdent's: locals (parameters) first, across the whole scope chain, then
// data and facts, then lets nearest first, then rules.
func (w *selfRefWalk) resolve(name string, scope *selfRefFrame) any {
	for f := scope; f != nil; f = f.parent {
		if _, ok := f.params[name]; ok {
			return nil
		}
	}
	if _, ok := w.policy.Data[name]; ok {
		return nil
	}
	if _, ok := w.policy.Facts[name]; ok {
		return nil
	}
	for f := scope; f != nil; f = f.parent {
		if vd, ok := f.lets[name]; ok {
			if vd == nil {
				return nil
			}
			return vd
		}
	}
	if vd, ok := w.policy.Lets[name]; ok {
		return vd
	}
	if r, ok := w.policy.Rules[name]; ok {
		return r
	}
	return nil
}

// boundElsewhere reports whether name, the name of the let vd whose value
// reads it, is bound anywhere else in the policy: another block let or policy
// let, a lambda parameter, a rule, a fact or pack data.
//
// evalIdent reads a cached local (ExecutionContext.GetLocal, across every
// ancestor context) before it looks a let up, and an evaluated let or rule is
// cached as a local under its name. A same-policy rule or a policy let is
// evaluated in the context that reads it, so its block contexts are children
// of the caller's block or lambda context. So once any other binding of the
// name is in an ancestor context, the let's own read returns that value
// instead of recursing: with `let x = 1`, `rule r = (x + { let x = x + 1 yield
// x }) == 2` evaluates to true, and so does `rule r = { let x = 1 yield x + r2
// } == 2` with `rule r2 = { let x = x + 1 yield x }`. Whether that happens
// depends on evaluation order and call paths, which this check does not
// model, so such a self-reference is skipped as ambiguous rather than
// reported. This is a known gap: when no other binding is in an ancestor
// context, the pack still fails at runtime with "references itself".
func (w *selfRefWalk) boundElsewhere(name string, vd *ast.VarDeclaration) bool {
	if w.binders != nil {
		if _, ok := w.binders.params[name]; ok {
			return true
		}
		for other := range w.binders.blockLets[name] {
			if other != vd {
				return true
			}
		}
	}
	if pl, ok := w.policy.Lets[name]; ok && pl != vd {
		return true
	}
	if _, ok := w.policy.Rules[name]; ok {
		return true
	}
	if _, ok := w.policy.Facts[name]; ok {
		return true
	}
	_, ok := w.policy.Data[name]
	return ok
}

// walk visits e with the given scope. owner is the rule or let whose value e
// is part of, or nil where a self-reference is not reported.
func (w *selfRefWalk) walk(e ast.Node, scope *selfRefFrame, owner any) error {
	if e == nil {
		return nil
	}
	switch n := e.(type) {
	case *ast.Identifier:
		if owner == nil {
			return nil
		}
		target := w.resolve(n.Value, scope)
		if target == nil || target != owner {
			return nil
		}
		if vd, isLet := target.(*ast.VarDeclaration); isLet && w.boundElsewhere(n.Value, vd) {
			return nil
		}
		*w.errs = append(*w.errs, siteError{at: n.Span(), err: selfReferenceError(owner, n)})
		return nil
	case *ast.CallExpression:
		// The callee is resolved by getTarget, never evaluated as an
		// identifier; calling a rule or a let lambda is checkCallTargets'.
		for _, a := range n.Arguments {
			if err := w.walk(a, scope, owner); err != nil {
				return err
			}
		}
		return nil
	case *ast.LambdaExpression:
		frame := &selfRefFrame{parent: scope, params: make(map[string]struct{}, len(n.Params))}
		for _, p := range n.Params {
			frame.params[p] = struct{}{}
		}
		if n.Body == nil {
			return nil
		}
		return w.walk(n.Body, frame, nil)
	case *ast.BlockExpression:
		// The runtime injects every let of a block before evaluating any of
		// them (evalBlock), so each initializer sees all of them.
		// A name declared twice fails the block at runtime with a let
		// conflict, so it resolves to neither declaration and is bound to
		// nil: never reported here.
		frame := &selfRefFrame{parent: scope, lets: map[string]*ast.VarDeclaration{}}
		for _, st := range n.Statements {
			if vd, ok := st.(*ast.VarDeclaration); ok {
				if _, dup := frame.lets[vd.Name]; dup {
					frame.lets[vd.Name] = nil
					continue
				}
				frame.lets[vd.Name] = vd
			}
		}
		for _, st := range n.Statements {
			// Only lets are evaluated; evalBlock ignores any other statement.
			if vd, ok := st.(*ast.VarDeclaration); ok {
				if err := w.walk(vd.Value, frame, selfRefOwnerOfLet(vd)); err != nil {
					return err
				}
			}
		}
		return w.walk(n.Yield, frame, owner)
	default:
		return forEachExprChild(e, func(child ast.Node) error {
			return w.walk(child, scope, owner)
		})
	}
}

func selfReferenceError(owner any, at *ast.Identifier) error {
	switch o := owner.(type) {
	case *Rule:
		fqn := o.FQN.String()
		return fmt.Errorf("rule '%s' references itself at %s: %w: %w", fqn, at.Span(), xerr.ErrInfiniteRecursion([]string{fqn, fqn}), xerr.ErrIndex)
	case *ast.VarDeclaration:
		return fmt.Errorf("let '%s' references itself at %s: %w: %w", o.Name, at.Span(), xerr.ErrInfiniteRecursion([]string{o.Name, o.Name}), xerr.ErrIndex)
	}
	return fmt.Errorf("'%s' references itself at %s: %w", at.Value, at.Span(), xerr.ErrIndex)
}
