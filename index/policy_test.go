// SPDX-License-Identifier: Apache-2.0
//
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
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func (suite *IndexTestSuite) TestCreatePolicy() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional (old required=false meant optional)
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Equal(policyStmt, policy.Statement)
	suite.Equal(suite.policyNs, policy.Namespace)
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

func (suite *IndexTestSuite) TestCreatePolicyWithoutExports() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional (old required=false meant optional)
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "does not export any rules")
}

func (suite *IndexTestSuite) TestCreatePolicyWithInvalidFactPosition() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "'fact' must appear before rules, exports, lets, and shapes")
}

func (suite *IndexTestSuite) TestCreatePolicyWithInvalidUsePosition() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "'use' must appear before rules, exports, lets, and shapes")
}

func (suite *IndexTestSuite) TestCreatePolicyWithUnknownRuleExport() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "cannot export unknown rule")
}

func (suite *IndexTestSuite) TestCreatePolicyWithDuplicateRuleExport() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "conflict: rule export")
}

func (suite *IndexTestSuite) TestAddLet() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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

func (suite *IndexTestSuite) TestAddLetWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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
	suite.Contains(err.Error(), "conflict: let declaration")
}

func (suite *IndexTestSuite) TestAddRule() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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

func (suite *IndexTestSuite) TestAddRuleWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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
	suite.Contains(err.Error(), "conflict: rule declaration")
}

func (suite *IndexTestSuite) TestPolicyAddShape() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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
	suite.Equal(suite.policyNs, shape.Namespace)
	suite.Equal(policy, shape.Policy)
}

func (suite *IndexTestSuite) TestPolicyAddShapeWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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
	suite.Contains(err.Error(), "conflict: shape declaration")
}

func (suite *IndexTestSuite) TestAddFact() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	factStmt := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		ast.NewStringLiteral("default user", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		true, // optional (has default, so was optional)
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

func (suite *IndexTestSuite) TestAddFactRequiredCannotHaveDefault() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Required fact (not optional) cannot have default value
	factStmt := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		ast.NewStringLiteral("default user", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		false, // required (not optional)
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddFact(factStmt)

	suite.Error(err)
	suite.Contains(err.Error(), "required fact")
	suite.Contains(err.Error(), "cannot have a default value")
}

func (suite *IndexTestSuite) TestAddFactOptionalCanHaveDefault() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Optional fact can have default value
	factStmt := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		ast.NewStringLiteral("default user", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 20, Offset: 20}, To: tokens.Pos{Line: 1, Column: 20, Offset: 20}}),
		true, // optional
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddFact(factStmt)

	suite.NoError(err)
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.True(policy.Facts["user"].Optional, "Fact should be optional")
}

func (suite *IndexTestSuite) TestAddFactRequiredWithoutDefault() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	// Required fact without default is valid
	factStmt := ast.NewFactStatement(
		"user",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		"user",
		nil,   // no default
		false, // required (not optional)
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err := policy.AddFact(factStmt)

	suite.NoError(err)
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.False(policy.Facts["user"].Optional, "Fact should be required")
	suite.Nil(policy.Facts["user"].Default, "Required fact should not have default")
}

func (suite *IndexTestSuite) TestAddFactWithNameConflict() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
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
		true, // optional (has default, so was optional)
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
		true, // optional (has default, so was optional)
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	err = policy.AddFact(factStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "conflict: fact declaration")
}

func (suite *IndexTestSuite) TestPolicyString() {
	policy := &Policy{
		Statement:       &ast.PolicyStatement{},
		Namespace:       suite.policyNs,
		Name:            "testPolicy",
		FQN:             ast.NewFQN([]string{"com", "example", "testPolicy"}, tokens.Range{}),
		FilePath:        "test.sentra",
		Statements:      []ast.Statement{},
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	suite.Equal("com/example/testPolicy", policy.String())
}

func (suite *IndexTestSuite) TestCreatePolicyWithComments() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewCommentStatement("This is a comment", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Len(policy.Facts, 1)
	suite.Len(policy.Rules, 1)
	suite.Len(policy.RuleExports, 1)
	// Comments should be ignored
}

func (suite *IndexTestSuite) TestCreatePolicyWithValidUseStatement() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "user", nil, true, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
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

	policy, err := createPolicy(suite.policyNs, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Len(policy.Uses, 1)
	suite.Equal("use fn, fn2 from @sentrie/std as std", policy.Uses["std"].String())
}

func (suite *IndexTestSuite) TestCreatePolicyWithMetadataAndPhases() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewCommentStatement("-- meta", r(3)),
			ast.NewTitleStatement("Hello", r(4)),
			ast.NewDescriptionStatement("", r(5)),
			ast.NewVersionStatement("1.2.3", r(6)),
			ast.NewTagStatement("a", "1", r(7)),
			ast.NewTagStatement("a", "2", r(8)),
			ast.NewTagStatement("b", "  \t  ", r(9)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(10)), "user", nil, true, r(10)),
			ast.NewUseStatement([]string{"x"}, "", []string{"sentrie", "std"}, "std", r(11)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(12)), nil, r(12)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(13)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	policy, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.NoError(err)
	require.NotNil(suite.T(), policy)
	suite.Equal("Hello", *policy.Title)
	suite.Equal("", *policy.Description)
	suite.Equal("1.2.3", policy.VersionLiteral)
	suite.NotNil(policy.Version)
	suite.Len(policy.TagPairs, 3)
	suite.Equal("a", policy.TagPairs[0].Key)
	suite.Equal("1", policy.TagPairs[0].Value)
	suite.Equal("2", policy.TagPairs[1].Value)
	suite.Equal("b", policy.TagPairs[2].Key)
	suite.Equal("  \t  ", policy.TagPairs[2].Value)
	require.NotNil(suite.T(), policy.TagsByKey)
	suite.Equal([]string{"1", "2"}, policy.TagsByKey["a"])
	suite.Equal([]string{"  \t  "}, policy.TagsByKey["b"])
}

func (suite *IndexTestSuite) TestCreatePolicyUseWithoutFacts() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewUseStatement([]string{"x"}, "", []string{"sentrie", "std"}, "std", r(3)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(4)), nil, r(4)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(5)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	policy, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.NoError(err)
	suite.Len(policy.Uses, 1)
}

func (suite *IndexTestSuite) TestCreatePolicyMetadataThenUseWithoutFacts() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewTitleStatement("T", r(3)),
			ast.NewUseStatement([]string{"x"}, "", []string{"sentrie", "std"}, "std", r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	policy, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.NoError(err)
	require.Equal(suite.T(), "T", *policy.Title)
}

func (suite *IndexTestSuite) TestCreatePolicyVersionVPrefixAccepted() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewVersionStatement("v1.2.3", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	policy, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.NoError(err)
	suite.Equal("v1.2.3", policy.VersionLiteral)
	suite.NotNil(policy.Version)
}

func (suite *IndexTestSuite) TestCreatePolicyFactAfterUseErrors() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewUseStatement([]string{"x"}, "", []string{"sentrie", "std"}, "std", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "fact statements must appear before any use statements")
}

func (suite *IndexTestSuite) TestCreatePolicyMetadataAfterFactErrors() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(3)), "user", nil, true, r(3)),
			ast.NewTitleStatement("Late", r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "title, description, version, and tag may only appear in one contiguous block")
}

func (suite *IndexTestSuite) TestCreatePolicyShapeBeforeFactErrors() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	shapeStmt := ast.NewShapeStatement(
		"S",
		ast.NewStringTypeRef(r(3)),
		nil,
		r(3),
	)
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			shapeStmt,
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "'fact' must appear before rules, exports, lets, and shapes")
}

func (suite *IndexTestSuite) TestCreatePolicyDuplicateTitle() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewTitleStatement("A", r(3)),
			ast.NewTitleStatement("B", r(4)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(5)), "user", nil, true, r(5)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(6)), nil, r(6)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(7)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "conflict: policy title")
}

func (suite *IndexTestSuite) TestCreatePolicyInvalidSemVer() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewVersionStatement("not-a-version", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), `Invalid policy version: expected SemVer string (e.g., "1.2.3")`)
}

func (suite *IndexTestSuite) TestCreatePolicyEmptyTitle() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewTitleStatement("   ", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "policy title must not be empty or whitespace-only")
}

func (suite *IndexTestSuite) TestCreatePolicyEmptyTagKey() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewTagStatement("   ", "v", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "tag key must not be empty or whitespace-only")
}

func (suite *IndexTestSuite) TestCreatePolicyUseBeforeFactWhenBothPresent() {
	r := func(line int) tokens.Range {
		return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: line, Column: 0, Offset: 0}, To: tokens.Pos{Line: line, Column: 1, Offset: 1}}
	}
	policyStmt := ast.NewPolicyStatement(
		"p",
		[]ast.Statement{
			ast.NewUseStatement([]string{"x"}, "", []string{"sentrie", "std"}, "std", r(3)),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r(4)), "user", nil, true, r(4)),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, r(5)), nil, r(5)),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, r(6)),
		},
		r(2),
	)
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r(1)), r(1)),
			policyStmt,
		},
	}
	_, err := createPolicy(suite.policyNs, policyStmt, program)
	suite.Error(err)
	suite.Contains(err.Error(), "fact statements must appear before any use statements")
}
