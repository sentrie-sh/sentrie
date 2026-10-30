// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"maps"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

// ImportDecision resolves an ImportClause with `with` facts for sandboxed execution,
// and returns (decision-envelope, node, error).
func ImportDecision(ctx context.Context, exec *executorImpl, ec *ExecutionContext, p *index.Policy, t *ast.ImportClause) (box.Value, *trace.Node, error) {
	ctx, n, done := trace.New(ctx, t, "import", map[string]any{
		"what":  t.RuleToImport,
		"from":  t.FromPolicyFQN,
		"withs": len(t.Withs),
	})
	defer done()

	if len(t.FromPolicyFQN.Parts) < 2 {
		err := fmt.Errorf("import from must specify namespace/policy: got %v", t.FromPolicyFQN)
		return box.Null(), n.SetErr(err), err
	}

	rule := t.RuleToImport

	var ns, pol string
	if len(t.FromPolicyFQN.Parts) == 1 {
		// we only have a policy name - the namespace is the current policy's namespace
		ns = p.Namespace.FQN.String()
	} else {
		// we have a namespace and policy name
		ns = t.FromPolicyFQN.Parent().String()
	}
	pol = t.FromPolicyFQN.LastSegment()
	facts := make(map[string]box.Value)

	{ // resolve the policy and verify the rule is exported
		p, err := exec.index.ResolvePolicy(ns, pol)
		if err != nil {
			return box.Null(), n.SetErr(err), err
		}

		if err := p.VerifyRuleExported(rule); err != nil {
			return box.Null(), n.SetErr(err), err
		}

		for _, with := range t.Withs {
			// The target must declare the fact, keyed exactly as the index check
			// (checkImportWiths) and ExecRule key it. An unknown name is an
			// error, never skipped: a typo would otherwise run the target on
			// its default (#151). Reached only when the index check did not run.
			if _, ok := p.Facts[with.Name]; !ok {
				err := xerr.ErrInvalidInvocation(xerr.ImportUnknownFactMessage(with.Name, ns+ast.FQNSeparator+pol, p.FactNames()))
				return box.Null(), n.SetErr(err), err
			}

			// evaluate the with expression in the context of this execution context
			val, trace, err := eval(ctx, ec, exec, ec.policy, with.Expr)
			if err != nil {
				return box.Null(), n.SetErr(err), err
			}
			n.Attach(trace)

			facts[with.Name] = val
		}
	}

	injectedFacts := make(map[string]any, len(facts))
	for name, factValue := range facts {
		b, err := box.TryToBoundaryAny(factValue)
		if err != nil {
			err = fmt.Errorf("import with fact %q: %w", name, err)
			return box.Null(), n.SetErr(err), err
		}
		injectedFacts[name] = b
	}

	output, err := exec.ExecRule(ctx, ns, pol, rule, injectedFacts)
	if err != nil {
		if output != nil {
			n = n.Attach(output.RuleNode)
		}
		n.SetErr(err)
		return box.Null(), n, err
	}
	n = n.Attach(output.RuleNode)

	envelope := executorOutputEnvelope(output)
	n.SetResult(envelope)
	return envelope, n, nil
}

// executorOutputEnvelope builds the boxed value returned from import expressions:
// a map with "state" (trinary), "value" (decision payload), and one entry per rule attachment.
// "state" and "value" are applied after attachments so they cannot be shadowed by attachment names.
func executorOutputEnvelope(output *ExecutorOutput) box.Value {
	if output == nil {
		return box.Dict(map[string]box.Value{
			"state": box.Trinary(trinary.Unknown),
			"value": box.Undefined(),
		})
	}

	state := trinary.Unknown
	value := box.Undefined()
	if output.Decision != nil {
		state = output.Decision.State
		value = output.Decision.Value
	}

	m := make(map[string]box.Value, len(output.Attachments)+2)
	maps.Copy(m, map[string]box.Value(output.Attachments))
	m["state"] = box.Trinary(state)
	m["value"] = value
	return box.Dict(m)
}
