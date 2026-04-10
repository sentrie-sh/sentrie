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
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestImportWithFactBoundaryPreservesUndefined() {
	withFactValue := box.Undefined()
	boundary := box.ToBoundaryAny(withFactValue)
	decoded := box.FromBoundaryAny(boundary)
	s.Require().True(decoded.IsUndefined())
}

func (s *RuntimeTestSuite) TestImportWithFactBoundaryPreservesNestedUndefined() {
	withFactValue := box.Dict(map[string]box.Value{
		"payload": box.List([]box.Value{
			box.Number(1),
			box.Undefined(),
		}),
	})

	boundary := box.ToBoundaryAny(withFactValue)
	decoded := box.FromBoundaryAny(boundary)
	decodedMap, ok := decoded.DictValue()
	s.Require().True(ok)
	list, ok := decodedMap["payload"].ListValue()
	s.Require().True(ok)
	s.Require().Equal(1.0, list[0].Any())
	s.Require().True(list[1].IsUndefined())
}

func (s *RuntimeTestSuite) TestImportDecisionRejectsInvalidFromPolicyFQN() {
	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"policy_only"}, stubRange()).Ptr(),
		nil,
		stubRange(),
	)

	val, node, err := ImportDecision(
		s.T().Context(),
		&executorImpl{},
		&ExecutionContext{},
		&index.Policy{},
		imp,
	)
	s.Require().Error(err)
	s.Require().True(val.IsNull())
	s.Require().NotNil(node)
	s.Require().Contains(node.Err, "import from must specify namespace/policy")
}

func (s *RuntimeTestSuite) TestImportDecisionResolvePolicyFailure() {
	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"other", "policy"}, stubRange()).Ptr(),
		nil,
		stubRange(),
	)

	idx := index.CreateIndex()
	idx.Pack = &pack.PackFile{}
	exec := &executorImpl{index: idx}

	ec := NewExecutionContext(newEvalTestPolicy(), exec)
	val, _, err := ImportDecision(s.T().Context(), exec, ec, newEvalTestPolicy(), imp)
	s.Require().Error(err)
	s.Require().True(val.IsNull())
}

func (s *RuntimeTestSuite) TestExecutorOutputEnvelopeIncludesDecisionAndAttachments() {
	output := &ExecutorOutput{
		Decision: &Decision{
			State: trinary.False,
			Value: box.Number(42),
		},
		Attachments: DecisionAttachments{
			"reason": box.String("policy denied"),
		},
	}

	envelope := executorOutputEnvelope(output)
	m, ok := envelope.DictValue()
	s.Require().True(ok)
	s.Require().Equal(trinary.False, m["state"].Any())
	s.Require().Equal(42.0, m["value"].Any())
	s.Require().Equal("policy denied", m["reason"].Any())
}
