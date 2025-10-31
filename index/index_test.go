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
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/suite"
)

type IndexTestSuite struct {
	suite.Suite
	ctx context.Context
	idx *Index
}

func (suite *IndexTestSuite) SetupSuite() {
	suite.ctx = context.Background()
}

func (suite *IndexTestSuite) SetupTest() {
	suite.idx = CreateIndex()
}

func (suite *IndexTestSuite) TearDownTest() {
	suite.idx = nil
}

func TestIndexTestSuite(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}

func (suite *IndexTestSuite) TestCreateIndex() {
	idx := CreateIndex()

	suite.NotNil(idx)
	suite.NotNil(idx.theLock)
	suite.NotNil(idx.Namespaces)
	suite.NotNil(idx.Programs)
	suite.NotNil(idx.validationOnce)
	suite.NotNil(idx.commitOnce)
	suite.Equal(uint32(0), idx.validated)
	suite.Equal(uint32(0), idx.committed)
	suite.Nil(idx.Pack)
	suite.Nil(idx.validationError)
	suite.Nil(idx.commitError)
}

func (suite *IndexTestSuite) TestSetPack() {
	packFile := &pack.PackFile{}

	err := suite.idx.SetPack(suite.ctx, packFile)

	suite.NoError(err)
	suite.Equal(packFile, suite.idx.Pack)
}

func (suite *IndexTestSuite) TestSetPackWithNilPack() {
	err := suite.idx.SetPack(suite.ctx, nil)

	suite.NoError(err)
	suite.Nil(suite.idx.Pack)
}

func (suite *IndexTestSuite) TestSetPackWithCancelledContext() {
	cancelledCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	packFile := &pack.PackFile{}
	err := suite.idx.SetPack(cancelledCtx, packFile)

	// SetPack doesn't check for cancelled context, so it should succeed
	suite.NoError(err)
	suite.Equal(packFile, suite.idx.Pack)
}

func (suite *IndexTestSuite) TestAddProgramWithSimpleShape() {
	stubRange := tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}
	// Create a simple program with just a shape
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, stubRange), stubRange),
			ast.NewShapeStatement("User", ast.NewStringTypeRef(stubRange), nil, stubRange),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.NoError(err)
	suite.Len(suite.idx.Programs, 1)
	suite.Len(suite.idx.Namespaces, 1)

	// Check that the program was added
	addedProgram, exists := suite.idx.Programs["test.sentra"]
	suite.True(exists)
	suite.Equal("test.sentra", addedProgram.Reference.Reference)
	suite.Equal("com/example", addedProgram.Namespace.String())

	// Check that the namespace was created
	ns, exists := suite.idx.Namespaces["com/example"]
	suite.True(exists)
	suite.Equal("com/example", ns.FQN.String())
	suite.Len(ns.Shapes, 1)
	suite.Contains(ns.Shapes, "User")
}

func (suite *IndexTestSuite) TestAddProgramWithPolicy() {
	// Create a program with a policy that exports a rule
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewPolicyStatement(
				"AuthPolicy",
				[]ast.Statement{
					ast.NewFactStatement(
						"user",
						ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 20, Offset: 20}}),
						"user",
						ast.NewStringLiteral("testuser", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 20, Offset: 20}, To: tokens.Pos{Line: 3, Column: 30, Offset: 30}}),
						false,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
					),
					ast.NewRuleStatement(
						"allow",
						nil,
						ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 25, Offset: 25}}),
						nil,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}},
					),
					ast.NewRuleExportStatement(
						"allow",
						[]*ast.AttachmentClause{
							ast.NewAttachmentClause(
								"reason",
								ast.NewStringLiteral("user is allowed", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 25, Offset: 25}, To: tokens.Pos{Line: 5, Column: 35, Offset: 35}}),
								tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 25, Offset: 25}},
							),
						},
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
					),
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.NoError(err)
	suite.Len(suite.idx.Programs, 1)
	suite.Len(suite.idx.Namespaces, 1)

	// Check that the policy was added to the namespace
	ns, exists := suite.idx.Namespaces["com/example"]
	suite.True(exists)
	suite.Len(ns.Policies, 1)
	suite.Contains(ns.Policies, "AuthPolicy")

	policy := ns.Policies["AuthPolicy"]
	suite.Equal("AuthPolicy", policy.Name)
	suite.Equal("com/example/AuthPolicy", policy.FQN.String())
	suite.Len(policy.Rules, 1)
	suite.Contains(policy.Rules, "allow")
	suite.Len(policy.RuleExports, 1)
	suite.Contains(policy.RuleExports, "allow")
}

func (suite *IndexTestSuite) TestAddProgramWithShapeExport() {
	// Create a program with a shape export
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewShapeStatement(
				"User",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 20, Offset: 20}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
			ast.NewShapeExportStatement(
				"User",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.NoError(err)
	suite.Len(suite.idx.Programs, 1)
	suite.Len(suite.idx.Namespaces, 1)

	// Check that the shape export was added
	ns, exists := suite.idx.Namespaces["com/example"]
	suite.True(exists)
	suite.Len(ns.ShapeExports, 1)
	suite.Contains(ns.ShapeExports, "User")
}

func (suite *IndexTestSuite) TestAddProgramWithCancelledContext() {
	cancelledCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		},
	}

	err := suite.idx.AddProgram(cancelledCtx, program)

	suite.Error(err)
	suite.Equal(context.Canceled, err)
}

func (suite *IndexTestSuite) TestAddProgramWithMultipleNamespaces() {
	// Add first program
	program1 := &ast.Program{
		Reference: "test1.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program1)
	suite.NoError(err)

	// Add second program with different namespace
	program2 := &ast.Program{
		Reference: "test2.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"org", "test"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		},
	}

	err = suite.idx.AddProgram(suite.ctx, program2)
	suite.NoError(err)

	suite.Len(suite.idx.Programs, 2)
	suite.Len(suite.idx.Namespaces, 2)
	suite.Contains(suite.idx.Namespaces, "com/example")
	suite.Contains(suite.idx.Namespaces, "org/test")
}

func (suite *IndexTestSuite) TestAddProgramWithParentChildNamespaces() {
	// Add parent namespace first
	parentProgram := &ast.Program{
		Reference: "parent.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "parent.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "parent.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, parentProgram)
	suite.NoError(err)

	// Add child namespace
	childProgram := &ast.Program{
		Reference: "child.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example", "sub"}, tokens.Range{File: "child.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "child.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		},
	}

	err = suite.idx.AddProgram(suite.ctx, childProgram)
	suite.NoError(err)

	suite.Len(suite.idx.Programs, 2)
	suite.Len(suite.idx.Namespaces, 2)

	// Check parent-child relationship
	parentNs := suite.idx.Namespaces["com/example"]
	childNs := suite.idx.Namespaces["com/example/sub"]

	suite.NotNil(parentNs)
	suite.NotNil(childNs)
	suite.Len(parentNs.Children, 1)
	suite.Equal(childNs, parentNs.Children[0])
	suite.Equal(parentNs, childNs.Parent)
}

func (suite *IndexTestSuite) TestEnsureNamespaceWithExistingNamespace() {
	// Create first namespace
	nsStmt1 := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	ns1, err := suite.idx.ensureNamespace(suite.ctx, nsStmt1)
	suite.NoError(err)
	suite.NotNil(ns1)

	// Try to create the same namespace again
	nsStmt2 := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	ns2, err := suite.idx.ensureNamespace(suite.ctx, nsStmt2)
	suite.NoError(err)
	suite.Equal(ns1, ns2)              // Should return the same namespace
	suite.Len(suite.idx.Namespaces, 1) // Should still have only one namespace
}

func (suite *IndexTestSuite) TestEnsureNamespaceWithComplexHierarchy() {
	// Create root namespace
	rootStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com"}, tokens.Range{File: "root.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		tokens.Range{File: "root.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	rootNs, err := suite.idx.ensureNamespace(suite.ctx, rootStmt)
	suite.NoError(err)

	// Create intermediate namespace
	intermediateStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "intermediate.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		tokens.Range{File: "intermediate.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	intermediateNs, err := suite.idx.ensureNamespace(suite.ctx, intermediateStmt)
	suite.NoError(err)

	// Create leaf namespace
	leafStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub"}, tokens.Range{File: "leaf.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		tokens.Range{File: "leaf.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	leafNs, err := suite.idx.ensureNamespace(suite.ctx, leafStmt)
	suite.NoError(err)

	suite.Len(suite.idx.Namespaces, 3)

	// Check relationships
	suite.Len(rootNs.Children, 1)
	suite.Equal(intermediateNs, rootNs.Children[0])
	suite.Equal(rootNs, intermediateNs.Parent)

	suite.Len(intermediateNs.Children, 1)
	suite.Equal(leafNs, intermediateNs.Children[0])
	suite.Equal(intermediateNs, leafNs.Parent)
}

func (suite *IndexTestSuite) TestAddProgramWithShapeAndPolicyConflict() {
	// Create a program with a shape and policy that have the same name
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewShapeStatement(
				"User",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 20, Offset: 20}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
			ast.NewPolicyStatement(
				"User",
				[]ast.Statement{
					ast.NewFactStatement(
						"user",
						ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 20, Offset: 20}}),
						"user",
						ast.NewStringLiteral("testuser", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 20, Offset: 20}, To: tokens.Pos{Line: 4, Column: 30, Offset: 30}}),
						false,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}},
					),
					ast.NewRuleStatement(
						"allow",
						nil,
						ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 25, Offset: 25}}),
						nil,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
					),

					ast.NewRuleExportStatement(
						"allow",
						[]*ast.AttachmentClause{},
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}},
					),
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.Error(err)
	suite.Contains(err.Error(), "name conflict")
}

func (suite *IndexTestSuite) TestAddProgramWithPolicyWithoutExports() {
	// Create a program with a policy that doesn't export any rules
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewPolicyStatement(
				"AuthPolicy",
				[]ast.Statement{
					ast.NewFactStatement(
						"user",
						ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 20, Offset: 20}}),
						"user",
						ast.NewStringLiteral("testuser", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 20, Offset: 20}, To: tokens.Pos{Line: 3, Column: 30, Offset: 30}}),
						false,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
					),
					ast.NewRuleStatement(
						"allow",
						nil,
						ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 25, Offset: 25}}),
						nil,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}},
					),
					// No rule export statement
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.Error(err)
	suite.Contains(err.Error(), "does not export any rules")
}

func (suite *IndexTestSuite) TestAddProgramWithInvalidUseStatementPosition() {
	// Create a program with a use statement that's not immediately after facts
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewPolicyStatement(
				"AuthPolicy",
				[]ast.Statement{
					ast.NewFactStatement(
						"user",
						ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 20, Offset: 20}}),
						"user",
						ast.NewStringLiteral("testuser", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 20, Offset: 20}, To: tokens.Pos{Line: 3, Column: 30, Offset: 30}}),
						false,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
					),
					ast.NewRuleStatement(
						"allow",
						nil,
						ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 25, Offset: 25}}),
						nil,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}},
					),
					ast.NewUseStatement(
						[]string{"com", "other", "policy"},
						"",
						nil,
						"",
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
					),
					ast.NewRuleExportStatement(
						"allow",
						[]*ast.AttachmentClause{},
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}},
					),
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.Error(err)
	suite.Contains(err.Error(), "'use' statement must be immediately after facts")
}

func (suite *IndexTestSuite) TestAddProgramWithInvalidFactStatementPosition() {
	// Create a program with a fact statement that's not the first statement
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
			ast.NewPolicyStatement(
				"AuthPolicy",
				[]ast.Statement{
					ast.NewRuleStatement(
						"allow",
						nil,
						ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 25, Offset: 25}}),
						nil,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}},
					),
					ast.NewFactStatement(
						"user",
						ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 20, Offset: 20}}),
						"user",
						ast.NewStringLiteral("testuser", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 20, Offset: 20}, To: tokens.Pos{Line: 4, Column: 30, Offset: 30}}),
						false,
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}},
					),
					ast.NewRuleExportStatement(
						"allow",
						[]*ast.AttachmentClause{},
						tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
					),
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}},
			),
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.Error(err)
	suite.Contains(err.Error(), "fact statement must be the first statement in a policy")
}
