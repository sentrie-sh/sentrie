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
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func newExecutorAndPolicyWithFact(fact *ast.FactStatement) (*executorImpl, *index.Policy) {
	idx := index.CreateIndex()
	nsFQN := ast.NewFQN([]string{"test", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
		Children:     []*index.Namespace{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	p := &index.Policy{
		Namespace:   ns,
		Name:        "pol",
		FQN:         ast.CreateFQN(nsFQN, "pol"),
		Facts:       map[string]*ast.FactStatement{fact.Alias: fact},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Lets:        map[string]*ast.VarDeclaration{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	ruleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	rule := &index.Rule{
		Node:   ruleStmt,
		Policy: p,
		Name:   "allow",
		FQN:    ast.CreateFQN(p.FQN, "allow"),
		Body:   ruleStmt.Body,
	}
	p.Rules["allow"] = rule
	p.RuleExports["allow"] = &index.ExportedRule{RuleName: "allow"}
	ns.Policies["pol"] = p

	return &executorImpl{index: idx}, p
}

func (s *RuntimeTestSuite) TestExecRuleFactNullBranchesWrapInvalidInvocation() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", nil, false, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{"user": nil})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user' cannot be null")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestExecRuleDefaultFactEvalErrorWrapsUnresolvableFact() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", ast.NewIdentifier("missing", stubRange()), true, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "unresolvable fact: user")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestExecRuleDefaultFactNullWrapsInvalidInvocation() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", ast.NewNullLiteral(stubRange()), true, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user' cannot have null default value")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestValidateAgainstShapeTypeRefFieldErrorBranches() {
	typeRef := ast.NewShapeTypeRef(ast.NewFQN([]string{"UserShape"}, stubRange()).Ptr(), stubRange())
	policy := &index.Policy{
		Shapes: map[string]*index.Shape{
			"UserShape": {
				Model: &index.ShapeModel{
					Fields: map[string]*index.ShapeModelField{
						"name": {Name: "name", Required: true, TypeRef: ast.NewStringTypeRef(stubRange())},
					},
				},
			},
		},
		Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
	}

	err := validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field name is required")

	policy.Shapes["UserShape"] = &index.Shape{
		Model: &index.ShapeModel{
			Fields: map[string]*index.ShapeModelField{
				"name": {Name: "name", NotNullable: true, TypeRef: ast.NewStringTypeRef(stubRange())},
			},
		},
	}
	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{"name": nil}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field name cannot be null")

	policy.Shapes["UserShape"] = &index.Shape{
		Model: &index.ShapeModel{
			Fields: map[string]*index.ShapeModelField{
				"age": {Name: "age", TypeRef: ast.NewNumberTypeRef(stubRange())},
			},
		},
	}
	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{"age": "bad"}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field 'age' is not valid")
}
