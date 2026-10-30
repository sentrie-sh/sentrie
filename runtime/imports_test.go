// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
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

func (s *RuntimeTestSuite) TestImportDecisionMissingRequiredFactReturnsError() {
	ctx := s.T().Context()
	idx := index.CreateIndex()
	exec := &executorImpl{index: idx}

	nsFQN := ast.NewFQN([]string{"test", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	targetPolicy := &index.Policy{
		Name:      "target",
		FQN:       ast.CreateFQN(nsFQN, "target"),
		Namespace: ns,
		Lets:      map[string]*ast.VarDeclaration{},
		Facts:     map[string]*ast.FactStatement{},
		Rules:     map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{
			"allow": {RuleName: "allow"},
		},
		Uses:   map[string]*ast.UseStatement{},
		Shapes: map[string]*index.Shape{},
	}
	targetPolicy.Facts["req"] = ast.NewFactStatement("req", nil, "req", nil, false, stubRange())
	targetRuleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	targetPolicy.Rules["allow"] = &index.Rule{
		Node:   targetRuleStmt,
		Policy: targetPolicy,
		Name:   "allow",
		FQN:    ast.CreateFQN(targetPolicy.FQN, "allow"),
		Body:   targetRuleStmt.Body,
	}
	ns.Policies[targetPolicy.Name] = targetPolicy

	callerPolicy := &index.Policy{
		Name:        "caller",
		FQN:         ast.CreateFQN(nsFQN, "caller"),
		Namespace:   ns,
		Lets:        map[string]*ast.VarDeclaration{},
		Facts:       map[string]*ast.FactStatement{},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	ec := NewExecutionContext(callerPolicy, exec)

	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"test", "ns", "target"}, stubRange()).Ptr(),
		nil,
		stubRange(),
	)

	val, node, err := ImportDecision(ctx, exec, ec, callerPolicy, imp)
	s.Require().Error(err)
	s.ErrorIs(err, xerr.InvalidInvocationError{})
	s.Require().Contains(err.Error(), "required fact not found: req")
	s.Require().True(val.IsNull())
	s.Require().NotNil(node)
	s.Require().Contains(node.Err, "required fact not found: req")
}

// TestImportDecisionUnknownWithFactReturnsError covers the runtime fallback of
// #151: a pack whose index never ran the import check must still fail closed on
// a with name the target does not declare, instead of running it on its default.
func (s *RuntimeTestSuite) TestImportDecisionUnknownWithFactReturnsError() {
	ctx := s.T().Context()
	idx := index.CreateIndex()
	exec := &executorImpl{index: idx}

	nsFQN := ast.NewFQN([]string{"test", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	targetPolicy := &index.Policy{
		Name:      "target",
		FQN:       ast.CreateFQN(nsFQN, "target"),
		Namespace: ns,
		Lets:      map[string]*ast.VarDeclaration{},
		Facts:     map[string]*ast.FactStatement{},
		Rules:     map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{
			"allow": {RuleName: "allow"},
		},
		Uses:   map[string]*ast.UseStatement{},
		Shapes: map[string]*index.Shape{},
	}
	// An optional fact: without the check the typo'd binding would be skipped
	// and the rule would run on the fact's absence.
	targetPolicy.Facts["level"] = ast.NewFactStatement("level", nil, "level", nil, true, stubRange())
	targetRuleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	targetPolicy.Rules["allow"] = &index.Rule{
		Node:   targetRuleStmt,
		Policy: targetPolicy,
		Name:   "allow",
		FQN:    ast.CreateFQN(targetPolicy.FQN, "allow"),
		Body:   targetRuleStmt.Body,
	}
	ns.Policies[targetPolicy.Name] = targetPolicy

	callerPolicy := &index.Policy{
		Name:        "caller",
		FQN:         ast.CreateFQN(nsFQN, "caller"),
		Namespace:   ns,
		Lets:        map[string]*ast.VarDeclaration{},
		Facts:       map[string]*ast.FactStatement{},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	ec := NewExecutionContext(callerPolicy, exec)

	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"test", "ns", "target"}, stubRange()).Ptr(),
		[]*ast.WithClause{
			ast.NewWithClause("levl", ast.NewIntegerLiteral(3, stubRange()), stubRange()),
		},
		stubRange(),
	)

	val, node, err := ImportDecision(ctx, exec, ec, callerPolicy, imp)
	s.Require().Error(err)
	s.ErrorIs(err, xerr.ErrIndex)
	s.Contains(err.Error(), `import binds unknown fact "levl": policy test/ns/target accepts fact aliases: level`)
	s.True(val.IsNull())
	s.Require().NotNil(node)
	s.Contains(node.Err, `unknown fact "levl"`)
}
