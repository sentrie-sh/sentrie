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
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestImportWithFactBoundaryPreservesUndefined(t *testing.T) {
	withFactValue := box.Undefined()
	boundary := box.ToBoundaryAny(withFactValue)
	decoded := box.FromBoundaryAny(boundary)
	require.True(t, decoded.IsUndefined())
}

func TestImportWithFactBoundaryPreservesNestedUndefined(t *testing.T) {
	withFactValue := box.Map(map[string]box.Value{
		"payload": box.List([]box.Value{
			box.Number(1),
			box.Undefined(),
		}),
	})

	boundary := box.ToBoundaryAny(withFactValue)
	decoded := box.FromBoundaryAny(boundary)
	decodedMap, ok := decoded.MapValue()
	require.True(t, ok)
	list, ok := decodedMap["payload"].ListValue()
	require.True(t, ok)
	require.Equal(t, 1.0, list[0].Any())
	require.True(t, list[1].IsUndefined())
}

func TestImportDecisionRejectsInvalidFromPolicyFQN(t *testing.T) {
	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"policy_only"}, stubRange()).Ptr(),
		nil,
		stubRange(),
	)

	val, node, err := ImportDecision(
		t.Context(),
		&executorImpl{},
		&ExecutionContext{},
		&index.Policy{},
		imp,
	)
	require.Error(t, err)
	require.True(t, val.IsNull())
	require.NotNil(t, node)
	require.Contains(t, node.Err, "import from must specify namespace/policy")
}

func TestImportDecisionResolvePolicyFailure(t *testing.T) {
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
	val, _, err := ImportDecision(t.Context(), exec, ec, newEvalTestPolicy(), imp)
	require.Error(t, err)
	require.True(t, val.IsNull())
}

func TestExecutorOutputEnvelopeIncludesDecisionAndAttachments(t *testing.T) {
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
	m, ok := envelope.MapValue()
	require.True(t, ok)
	require.Equal(t, trinary.False, m["state"].Any())
	require.Equal(t, 42.0, m["value"].Any())
	require.Equal(t, "policy denied", m["reason"].Any())
}
