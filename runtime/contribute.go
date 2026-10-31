// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/tokens"
)

// validateRuleValue checks a value produced by a rule's body or default
// against the rule's declared type. A rule without a type, and a body that
// yields undefined, are not checked. An imported decision is checked as the
// value it reports.
func validateRuleValue(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, r *index.Rule, val box.Value, at tokens.Range) error {
	if r.Node.Type == nil {
		return nil
	}
	val = DecisionOf(val).reported().Value
	if val.IsUndefined() {
		return nil
	}
	return validateValueAgainstTypeRef(ctx, ec, e, p, val, r.Node.Type, at)
}

// applyContributions combines the rule's own outcome, base, with its
// contributions, evaluated one after another in declaration order so the
// result never depends on scheduling (#143).
//
// A contribution whose `when` does not hold, or whose body yields undefined,
// is skipped. Every other contribution's value must match the type the rule
// takes (the element type of a list[T] rule, or the boolean type itself).
//
//   - list[T]: the result is the base list followed by each contribution's
//     value. A base that is not a list (undefined, null, or the rule's guard
//     did not hold and it has no default) starts the list empty.
//   - boolean or trinary: the result is the Kleene OR of the base state and
//     each contribution's value.
//
// On any error the decision is nil, so ExecRule reports Unknown rather than
// the base or a partly combined value.
func applyContributions(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, r *index.Rule, base *Decision, rn *trace.Node) (*Decision, error) {
	kind, yieldType := index.ContributionKindOf(r.Node.Type)

	var items []box.Value
	state := base.State
	if kind == index.ContributionList {
		if xs, ok := base.Value.ListValue(); ok {
			items = slices.Clone(xs)
		} else {
			items = []box.Value{}
		}
	}

	for i, c := range r.Contributions {
		val, applied, err := evalContribution(ctx, ec, e, p, i, c, rn)
		if err != nil {
			return nil, err
		}
		if !applied {
			continue
		}
		if err := validateValueAgainstTypeRef(ctx, ec, e, p, val, yieldType, c.Body.Span()); err != nil {
			return nil, err
		}
		if kind == index.ContributionList {
			items = append(items, val)
			continue
		}
		tv, ok := val.TrinaryValue()
		if !ok {
			// An imported decision passes the boolean type check by its state.
			tv, _ = val.DecisionState()
		}
		state = state.Or(tv)
	}

	var combined box.Value
	if kind == index.ContributionList {
		combined = box.List(items)
	} else {
		combined = box.Trinary(state)
	}
	// The combined value must satisfy the whole declared type too (list
	// constraints such as a maximum length apply to the combined list).
	if err := validateValueAgainstTypeRef(ctx, ec, e, p, combined, r.Node.Type, r.Span()); err != nil {
		return nil, err
	}
	return DecisionOf(combined), nil
}

// evalContribution evaluates one contribution's guard and body, attaching a
// "rule-contribution" trace node to rn. applied is false when the guard does
// not hold or the body yields undefined.
func evalContribution(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, i int, c *index.Contribution, rn *trace.Node) (box.Value, bool, error) {
	meta := map[string]any{"index": i}
	ctx, cn, done := trace.New(ctx, c.Node, "rule-contribution", meta)
	defer done()
	rn.Attach(cn)

	if c.When != nil {
		ctx, wn, done := trace.New(ctx, c.When, "contribution-when", map[string]any{})
		defer done()

		cond, condNode, err := eval(ctx, ec, e, p, c.When)
		wn.Attach(condNode).SetResult(cond).SetErr(err)
		cn.Attach(wn)
		if err != nil {
			cn.SetErr(err)
			return box.Undefined(), false, err
		}
		if !box.TrinaryFrom(cond).IsTrue() {
			meta["skipped"] = "when"
			return box.Undefined(), false, nil
		}
	}

	ctx, bn, done := trace.New(ctx, c.Body, "contribution-body", map[string]any{})
	defer done()

	val, bodyNode, err := eval(ctx, ec, e, p, c.Body)
	bn.Attach(bodyNode).SetResult(val).SetErr(err)
	cn.Attach(bn)
	if err != nil {
		cn.SetErr(err)
		return box.Undefined(), false, err
	}
	if val.IsUndefined() {
		meta["skipped"] = "undefined"
		return val, false, nil
	}
	cn.SetResult(val)
	return val, true, nil
}
