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

package index

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/suite"
)

type PolicyTestSuite struct {
	suite.Suite
	namespace *Namespace
}

func (suite *PolicyTestSuite) SetupTest() {
	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	suite.namespace = createNamespace(nsStmt)
}

func (suite *PolicyTestSuite) TearDownTest() {
	suite.namespace = nil
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) TestCreatePolicy() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				"user",
				nil,
				false,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"allow",
				[]*ast.AttachmentClause{
					ast.NewAttachmentClause(
						"reason",
						ast.NewStringLiteral("user is allowed", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 25, Offset: 25}, To: tokens.Pos{Line: 5, Column: 25, Offset: 25}}),
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 15, Offset: 15}},
					),
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Equal(policyStmt, policy.Statement)
	suite.Equal(suite.namespace, policy.Namespace)
	suite.Equal("testPolicy", policy.Name)
	suite.Equal("com/example/testPolicy", policy.FQN.String())
	suite.Equal("test.sentra", policy.FilePath)
	suite.Equal(policyStmt.Statements, policy.Statements)

	// Check maps are initialized
	suite.NotNil(policy.Lets)
	suite.NotNil(policy.Facts)
	suite.NotNil(policy.Rules)
	suite.NotNil(policy.RuleExports)
	suite.NotNil(policy.Uses)
	suite.NotNil(policy.Shapes)
	suite.NotNil(policy.seenIdentifiers)

	// Check that facts, rules, and exports were processed
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.Len(policy.Rules, 1)
	suite.Contains(policy.Rules, "allow")
	suite.Len(policy.RuleExports, 1)
	suite.Contains(policy.RuleExports, "allow")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithoutExports() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				"user",
				nil,
				false,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
			// No rule export statement
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "does not export any rules")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithInvalidFactPosition() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "fact statement must be the first statement in a policy")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithInvalidUsePosition() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewUseStatement([]string{"com", "other", "policy"}, "", nil, "", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "'use' statement must be immediately after facts have been declared in a policy")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithUnknownRuleExport() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("unknownRule", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "cannot export unknown rule")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithDuplicateRuleExport() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "rule export conflict")
}

func (suite *PolicyTestSuite) TestAddLet() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	letStmt := ast.NewVarDeclaration(
		"testVar",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewStringLiteral("test value", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddLet(letStmt)

	suite.NoError(err)
	suite.Len(policy.Lets, 1)
	suite.Contains(policy.Lets, "testVar")
	suite.Equal(letStmt, policy.Lets["testVar"])
	suite.Contains(policy.seenIdentifiers, "testVar")
	suite.Equal(letStmt, policy.seenIdentifiers["testVar"])
}

func (suite *PolicyTestSuite) TestAddLetWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Add first let
	letStmt1 := ast.NewVarDeclaration(
		"testVar",
		ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewStringLiteral("test value 1", tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddLet(letStmt1)
	suite.NoError(err)

	// Try to add second let with same name
	letStmt2 := ast.NewVarDeclaration(
		"testVar", // Same name
		ast.NewNumberTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewIntegerLiteral(42, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err = policy.AddLet(letStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "let name conflict")
}

func (suite *PolicyTestSuite) TestAddRule() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	ruleStmt := ast.NewRuleStatement(
		"testRule",
		nil,
		ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewStringLiteral("test result", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddRule(ruleStmt)

	suite.NoError(err)
	suite.Len(policy.Rules, 1)
	suite.Contains(policy.Rules, "testRule")
	suite.Contains(policy.seenIdentifiers, "testRule")

	rule := policy.Rules["testRule"]
	suite.Equal("testRule", rule.Name)
	suite.Equal("com/example/testPolicy/testRule", rule.FQN.String())
	suite.Equal(ruleStmt, rule.Node)
	suite.Equal(policy, rule.Policy)
}

func (suite *PolicyTestSuite) TestAddRuleWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Add first rule
	ruleStmt1 := ast.NewRuleStatement(
		"testRule",
		nil,
		ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewStringLiteral("test result 1", tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddRule(ruleStmt1)
	suite.NoError(err)

	// Try to add second rule with same name
	ruleStmt2 := ast.NewRuleStatement(
		"testRule", // Same name
		nil,
		ast.NewTrinaryLiteral(trinary.False, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		ast.NewStringLiteral("test result 2", tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 25, Offset: 25}, To: tokens.Pos{Line: 1, Column: 25, Offset: 25}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err = policy.AddRule(ruleStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "rule name conflict")
}

func (suite *PolicyTestSuite) TestAddShape() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	shapeStmt := ast.NewShapeStatement(
		"testShape",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddShape(shapeStmt)

	suite.NoError(err)
	suite.Len(policy.Shapes, 1)
	suite.Contains(policy.Shapes, "testShape")

	shape := policy.Shapes["testShape"]
	suite.Equal("testShape", shape.Name)
	suite.Equal("com/example/testPolicy/testShape", shape.FQN.String())
	suite.Equal(shapeStmt, shape.Statement)
	suite.Equal(suite.namespace, shape.Namespace)
	suite.Equal(policy, shape.Policy)
}

func (suite *PolicyTestSuite) TestAddShapeWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Add first shape
	shapeStmt1 := ast.NewShapeStatement(
		"testShape",
		ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddShape(shapeStmt1)
	suite.NoError(err)

	// Try to add second shape with same name
	shapeStmt2 := ast.NewShapeStatement(
		"testShape", // Same name
		ast.NewNumberTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err = policy.AddShape(shapeStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "shape name conflict")
}

func (suite *PolicyTestSuite) TestAddFact() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	factStmt := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		ast.NewStringLiteral("default user", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		false,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddFact(factStmt)

	suite.NoError(err)
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.Equal(factStmt, policy.Facts["user"])
	suite.Contains(policy.seenIdentifiers, "user")
	suite.Equal(factStmt, policy.seenIdentifiers["user"])
}

func (suite *PolicyTestSuite) TestAddFactWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Add first fact
	factStmt1 := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		ast.NewStringLiteral("default user 1", tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		false,
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddFact(factStmt1)
	suite.NoError(err)

	// Try to add second fact with same alias
	factStmt2 := ast.NewFactStatement(
		"admin",
		ast.NewStringTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user", // Same alias
		ast.NewStringLiteral("default user 2", tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		false,
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err = policy.AddFact(factStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "fact alias conflict")
}

func (suite *PolicyTestSuite) TestPolicyString() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.namespace,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	suite.Equal("com/example/testPolicy", policy.String())
}

func (suite *PolicyTestSuite) TestCreatePolicyWithComments() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewCommentStatement("This is a comment", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewCommentStatement("Another comment", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 15, Offset: 15}, To: tokens.Pos{Line: 6, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 7, Column: 0, Offset: 0}, To: tokens.Pos{Line: 7, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Len(policy.Facts, 1)
	suite.Len(policy.Rules, 1)
	suite.Len(policy.RuleExports, 1)
	// Comments should be ignored
}

func (suite *PolicyTestSuite) TestCreatePolicyWithValidUseStatement() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, false, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewUseStatement([]string{"fn", "fn2"}, "", []string{"sentrie", "std"}, "std", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Len(policy.Uses, 1)
	suite.Equal("use fn, fn2 from @sentrie/std as std", policy.Uses["std"].String())
}
