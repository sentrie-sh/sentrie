// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

// ImportDecision resolves an ImportClause with `with` facts for sandboxed execution,
// and returns (value+attachments-map, node, error).
func ImportDecision(ctx context.Context, exec *executorImpl, ec *ExecutionContext, currentPolicy *index.Policy, imp *ast.ImportClause) (*ExecutorOutput, *trace.Node, error) {
	n, done := trace.New("import", imp.RuleToImport, imp, map[string]any{
		"what": imp.RuleToImport,
		"from": imp.FromPolicyFQN,
		"with": len(imp.Withs),
	})
	defer done()

	if len(imp.FromPolicyFQN) < 2 {
		err := fmt.Errorf("import from must specify namespace/policy: got %v", imp.FromPolicyFQN)
		return nil, n.SetErr(err), err
	}

	rule := imp.RuleToImport

	var ns, pol string
	if len(imp.FromPolicyFQN) == 1 {
		// we only have a policy name - the namespace is the current policy's namespace
		ns = currentPolicy.Namespace.FQN.String()
	} else {
		// we have a namespace and policy name
		ns = imp.FromPolicyFQN.Parent().String()
	}
	pol = imp.FromPolicyFQN.LastSegment()
	facts := make(map[string]any)

	{ // resolve the policy and verify the rule is exported
		p, err := exec.index.ResolvePolicy(ns, pol)
		if err != nil {
			return nil, n.SetErr(err), err
		}

		if err := p.VerifyRuleExported(rule); err != nil {
			return nil, n.SetErr(err), err
		}

		for _, with := range imp.Withs {
			// find the fact in the target policy
			if _, ok := p.Facts[with.Name]; !ok {
				// no point evaluating - the target policy does not need this fact
				continue
			}

			// evaluate the with expression in the context of this execution context
			val, trace, err := eval(ctx, ec, exec, ec.policy, with.Expr)
			if err != nil {
				return nil, n.SetErr(err), err
			}
			n.Attach(trace)

			facts[with.Name] = val
		}
	}

	output, err := exec.ExecRule(ctx, ns, pol, rule, facts)
	n = n.Attach(output.RuleNode)
	if err != nil {
		n.SetErr(err)
		return nil, n, err
	}

	n.SetResult(output)

	return output, n, nil
}
