// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestEvaluateRuleOutcomeWhenFalseDefaultBranches() {
	p := newEvalTestPolicy()
	ruleStmt := ast.NewRuleStatement("r", nil, ast.NewTrinaryLiteral(trinary.False, stubRange()), ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	rule := &index.Rule{
		Node:    ruleStmt,
		Policy:  p,
		Name:    "r",
		FQN:     ast.CreateFQN(p.FQN, "r"),
		When:    ruleStmt.When,
		Body:    ruleStmt.Body,
		Default: nil,
	}
	ec := NewExecutionContext(p, &executorImpl{})
	decision, _, err := evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().NoError(err)
	s.Require().Equal(trinary.Unknown, decision.State)

	rule.Default = ast.NewTrinaryLiteral(trinary.True, stubRange())
	decision, _, err = evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().NoError(err)
	s.Require().Equal(trinary.True, decision.State)
}
