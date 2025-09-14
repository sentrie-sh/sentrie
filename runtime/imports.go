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
	"strings"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
)

type ImportDecisionResult struct {
	Decision    *Decision
	Attachments DecisionAttachments
}

// ImportDecision resolves an ImportClause with `with` facts for sandboxed execution,
// and returns (value+attachments-map, node, error).
func ImportDecision(ctx context.Context, exec *executorImpl, ec *ExecutionContext, currentPolicy *index.Policy, imp *ast.ImportClause) (*ImportDecisionResult, *trace.Node, error) {
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
		pol = imp.FromPolicyFQN[0]
	} else {
		// we have a namespace and policy name
		ns = strings.Join(imp.FromPolicyFQN[:len(imp.FromPolicyFQN)-1], ast.FQNSeparator)
		pol = imp.FromPolicyFQN[len(imp.FromPolicyFQN)-1]
	}

	p, err := exec.index.ResolvePolicy(ns, pol)
	if err != nil {
		return nil, n.SetErr(err), err
	}

	if err := p.VerifyRuleExported(rule); err != nil {
		return nil, n.SetErr(err), err
	}

	facts := make(map[string]any)
	for _, with := range imp.Withs {
		// find the fact in the target policy
		if _, ok := p.Facts[with.Name]; !ok {
			// no point evaluating - the target policy does not need this fact
			continue
		}

		// evaluate the with expression in the context of this execution context
		val, trace, err := eval(ctx, ec, exec, p, with.Expr)
		if err != nil {
			return nil, n.SetErr(err), err
		}
		n.Attach(trace)

		facts[with.Name] = val
	}

	decision, attachments, node, err := exec.ExecRule(ctx, ns, pol, rule, facts)
	n = n.Attach(node)
	if err != nil {
		n.SetErr(err)
		return nil, n, err
	}

	importDecisionResult := &ImportDecisionResult{
		Decision:    decision,
		Attachments: attachments,
	}

	n.SetResult(importDecisionResult)

	return importDecisionResult, n, nil
}
