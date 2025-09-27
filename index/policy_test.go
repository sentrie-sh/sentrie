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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/tokens"
	"github.com/stretchr/testify/suite"
)

type PolicyTestSuite struct {
	suite.Suite
	namespace *Namespace
}

func (suite *PolicyTestSuite) SetupTest() {
	// Create namespace
	nsStmt := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
		Name: ast.FQN{"com", "example"},
	}
	suite.namespace = createNamespace(nsStmt)
}

func (suite *PolicyTestSuite) TearDownTest() {
	suite.namespace = nil
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) TestCreatePolicy() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
					Value: 1,
				},
			},
			&ast.RuleExportStatement{
				Pos: tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Of:  "allow",
				Attachments: []*ast.AttachmentClause{
					{
						Pos:  tokens.Position{Filename: "test.sentra", Line: 5, Column: 15},
						What: "reason",
						As: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 5, Column: 25},
							Value: "user is allowed",
						},
					},
				},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
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
	suite.NotNil(policy.knownIdentifiers)

	// Check that facts, rules, and exports were processed
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.Len(policy.Rules, 1)
	suite.Contains(policy.Rules, "allow")
	suite.Len(policy.RuleExports, 1)
	suite.Contains(policy.RuleExports, "allow")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithoutExports() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
					Value: 1,
				},
			},
			// No rule export statement
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "does not export any rules")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithInvalidFactPosition() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 15},
					Value: 1,
				},
			},
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 4, Column: 10},
				},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Of:          "allow",
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "fact statement must be the first statement in a policy")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithInvalidUsePosition() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
					Value: 1,
				},
			},
			&ast.UseStatement{
				Pos:     tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Modules: []string{"com", "other", "policy"},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 6, Column: 0},
				Of:          "allow",
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "'use' statement must be immediately after facts have been declared in a policy")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithUnknownRuleExport() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
					Value: 1,
				},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Of:          "unknownRule", // Rule that doesn't exist
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.Error(err)
	suite.Nil(policy)
	suite.Contains(err.Error(), "cannot export unknown rule")
}

func (suite *PolicyTestSuite) TestCreatePolicyWithDuplicateRuleExport() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
					Value: 1,
				},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Of:          "allow",
				Attachments: []*ast.AttachmentClause{},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 6, Column: 0},
				Of:          "allow", // Duplicate export
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
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
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	letStmt := &ast.VarDeclaration{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
		Name: "testVar",
		Type: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test.sentra", Line: 1, Column: 15},
		},
		Value: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test.sentra", Line: 1, Column: 25},
			Value: "test value",
		},
	}

	err := policy.AddLet(letStmt)

	suite.NoError(err)
	suite.Len(policy.Lets, 1)
	suite.Contains(policy.Lets, "testVar")
	suite.Equal(letStmt, policy.Lets["testVar"])
	suite.Contains(policy.knownIdentifiers, "testVar")
	suite.Equal(letStmt, policy.knownIdentifiers["testVar"])
}

func (suite *PolicyTestSuite) TestAddLetWithNameConflict() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	// Add first let
	letStmt1 := &ast.VarDeclaration{
		Pos:  tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
		Name: "testVar",
		Type: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test1.sentra", Line: 1, Column: 15},
		},
		Value: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test1.sentra", Line: 1, Column: 25},
			Value: "test value 1",
		},
	}

	err := policy.AddLet(letStmt1)
	suite.NoError(err)

	// Try to add second let with same name
	letStmt2 := &ast.VarDeclaration{
		Pos:  tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
		Name: "testVar", // Same name
		Type: &ast.IntTypeRef{
			Pos: tokens.Position{Filename: "test2.sentra", Line: 1, Column: 15},
		},
		Value: &ast.IntegerLiteral{
			Pos:   tokens.Position{Filename: "test2.sentra", Line: 1, Column: 25},
			Value: 42,
		},
	}

	err = policy.AddLet(letStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "let name conflict")
}

func (suite *PolicyTestSuite) TestAddRule() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	ruleStmt := &ast.RuleStatement{
		Pos:      tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
		RuleName: "testRule",
		When: &ast.TrinaryLiteral{
			Pos:   tokens.Position{Filename: "test.sentra", Line: 1, Column: 15},
			Value: 1,
		},
		Body: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test.sentra", Line: 1, Column: 25},
			Value: "test result",
		},
	}

	err := policy.AddRule(ruleStmt)

	suite.NoError(err)
	suite.Len(policy.Rules, 1)
	suite.Contains(policy.Rules, "testRule")
	suite.Contains(policy.knownIdentifiers, "testRule")

	rule := policy.Rules["testRule"]
	suite.Equal("testRule", rule.Name)
	suite.Equal("com/example/testPolicy/testRule", rule.FQN.String())
	suite.Equal(ruleStmt, rule.Node)
	suite.Equal(policy, rule.Policy)
}

func (suite *PolicyTestSuite) TestAddRuleWithNameConflict() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	// Add first rule
	ruleStmt1 := &ast.RuleStatement{
		Pos:      tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
		RuleName: "testRule",
		When: &ast.TrinaryLiteral{
			Pos:   tokens.Position{Filename: "test1.sentra", Line: 1, Column: 15},
			Value: 1,
		},
		Body: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test1.sentra", Line: 1, Column: 25},
			Value: "test result 1",
		},
	}

	err := policy.AddRule(ruleStmt1)
	suite.NoError(err)

	// Try to add second rule with same name
	ruleStmt2 := &ast.RuleStatement{
		Pos:      tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
		RuleName: "testRule", // Same name
		When: &ast.TrinaryLiteral{
			Pos:   tokens.Position{Filename: "test2.sentra", Line: 1, Column: 15},
			Value: 0,
		},
		Body: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test2.sentra", Line: 1, Column: 25},
			Value: "test result 2",
		},
	}

	err = policy.AddRule(ruleStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "rule name conflict")
}

func (suite *PolicyTestSuite) TestAddShape() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	shapeStmt := &ast.ShapeStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
		Name: "testShape",
		Simple: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test.sentra", Line: 1, Column: 15},
		},
	}

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
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	// Add first shape
	shapeStmt1 := &ast.ShapeStatement{
		Pos:  tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
		Name: "testShape",
		Simple: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test1.sentra", Line: 1, Column: 15},
		},
	}

	err := policy.AddShape(shapeStmt1)
	suite.NoError(err)

	// Try to add second shape with same name
	shapeStmt2 := &ast.ShapeStatement{
		Pos:  tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
		Name: "testShape", // Same name
		Simple: &ast.IntTypeRef{
			Pos: tokens.Position{Filename: "test2.sentra", Line: 1, Column: 15},
		},
	}

	err = policy.AddShape(shapeStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "shape name conflict")
}

func (suite *PolicyTestSuite) TestAddFact() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	factStmt := &ast.FactStatement{
		Pos:   tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
		Name:  "user",
		Alias: "user",
		Type: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test.sentra", Line: 1, Column: 10},
		},
		Default: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test.sentra", Line: 1, Column: 20},
			Value: "default user",
		},
	}

	err := policy.AddFact(factStmt)

	suite.NoError(err)
	suite.Len(policy.Facts, 1)
	suite.Contains(policy.Facts, "user")
	suite.Equal(factStmt, policy.Facts["user"])
	suite.Contains(policy.knownIdentifiers, "user")
	suite.Equal(factStmt, policy.knownIdentifiers["user"])
}

func (suite *PolicyTestSuite) TestAddFactWithNameConflict() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	// Add first fact
	factStmt1 := &ast.FactStatement{
		Pos:   tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
		Name:  "user",
		Alias: "user",
		Type: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test1.sentra", Line: 1, Column: 10},
		},
		Default: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test1.sentra", Line: 1, Column: 20},
			Value: "default user 1",
		},
	}

	err := policy.AddFact(factStmt1)
	suite.NoError(err)

	// Try to add second fact with same alias
	factStmt2 := &ast.FactStatement{
		Pos:   tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
		Name:  "admin",
		Alias: "user", // Same alias
		Type: &ast.StringTypeRef{
			Pos: tokens.Position{Filename: "test2.sentra", Line: 1, Column: 10},
		},
		Default: &ast.StringLiteral{
			Pos:   tokens.Position{Filename: "test2.sentra", Line: 1, Column: 20},
			Value: "default user 2",
		},
	}

	err = policy.AddFact(factStmt2)

	suite.Error(err)
	suite.Contains(err.Error(), "fact alias conflict")
}

func (suite *PolicyTestSuite) TestPolicyString() {
	policy := &Policy{
		Statement:        &ast.PolicyStatement{},
		Namespace:        suite.namespace,
		Name:             "testPolicy",
		FQN:              ast.FQN{"com", "example", "testPolicy"},
		FilePath:         "test.sentra",
		Statements:       []ast.Statement{},
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	suite.Equal("com/example/testPolicy", policy.String())
}

func (suite *PolicyTestSuite) TestCreatePolicyWithComments() {
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.CommentStatement{
				Pos:     tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Content: "This is a comment",
			},
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 4, Column: 10},
				},
			},
			&ast.CommentStatement{
				Pos:     tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				Content: "Another comment",
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 6, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 6, Column: 15},
					Value: 1,
				},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 7, Column: 0},
				Of:          "allow",
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
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
	policyStmt := &ast.PolicyStatement{
		Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
		Name: "testPolicy",
		Statements: []ast.Statement{
			&ast.FactStatement{
				Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name:  "user",
				Alias: "user",
				Type: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
				},
			},
			&ast.UseStatement{
				Pos:     tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
				Modules: []string{"com", "other", "policy"},
			},
			&ast.RuleStatement{
				Pos:      tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
				RuleName: "allow",
				When: &ast.TrinaryLiteral{
					Pos:   tokens.Position{Filename: "test.sentra", Line: 5, Column: 15},
					Value: 1,
				},
			},
			&ast.RuleExportStatement{
				Pos:         tokens.Position{Filename: "test.sentra", Line: 6, Column: 0},
				Of:          "allow",
				Attachments: []*ast.AttachmentClause{},
			},
		},
	}

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.namespace, policyStmt, program)

	suite.NoError(err)
	suite.NotNil(policy)
	suite.Len(policy.Uses, 1)
	suite.Equal("use com, other, policy from  as ", policy.Uses[0].String())
}
