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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/pack"
	"github.com/binaek/sentra/tokens"
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
	// Create a simple program with just a shape
	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.ShapeStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "User",
				Simple: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 2, Column: 10},
				},
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.PolicyStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "AuthPolicy",
				Statements: []ast.Statement{
					&ast.FactStatement{
						Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
						Name:  "user",
						Alias: "user",
						Type: &ast.StringTypeRef{
							Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
						},
						Default: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 20},
							Value: "testuser",
						},
					},
					&ast.RuleStatement{
						Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
						RuleName: "allow",
						When: &ast.TrinaryLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
							Value: 1, // true in trinary
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
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.ShapeStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "User",
				Simple: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 2, Column: 10},
				},
			},
			&ast.ShapeExportStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name: "User",
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program1)
	suite.NoError(err)

	// Add second program with different namespace
	program2 := &ast.Program{
		Reference: "test2.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"org", "test"},
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "parent.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
		},
	}

	err := suite.idx.AddProgram(suite.ctx, parentProgram)
	suite.NoError(err)

	// Add child namespace
	childProgram := &ast.Program{
		Reference: "child.sentra",
		Statements: []ast.Statement{
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "child.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example", "sub"},
			},
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
	nsStmt1 := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "test1.sentra", Line: 1, Column: 0},
		Name: []string{"com", "example"},
	}

	ns1, err := suite.idx.ensureNamespace(suite.ctx, nsStmt1)
	suite.NoError(err)
	suite.NotNil(ns1)

	// Try to create the same namespace again
	nsStmt2 := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "test2.sentra", Line: 1, Column: 0},
		Name: []string{"com", "example"},
	}

	ns2, err := suite.idx.ensureNamespace(suite.ctx, nsStmt2)
	suite.NoError(err)
	suite.Equal(ns1, ns2)              // Should return the same namespace
	suite.Len(suite.idx.Namespaces, 1) // Should still have only one namespace
}

func (suite *IndexTestSuite) TestEnsureNamespaceWithComplexHierarchy() {
	// Create root namespace
	rootStmt := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "root.sentra", Line: 1, Column: 0},
		Name: []string{"com"},
	}

	rootNs, err := suite.idx.ensureNamespace(suite.ctx, rootStmt)
	suite.NoError(err)

	// Create intermediate namespace
	intermediateStmt := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "intermediate.sentra", Line: 1, Column: 0},
		Name: []string{"com", "example"},
	}

	intermediateNs, err := suite.idx.ensureNamespace(suite.ctx, intermediateStmt)
	suite.NoError(err)

	// Create leaf namespace
	leafStmt := &ast.NamespaceStatement{
		Pos:  tokens.Position{Filename: "leaf.sentra", Line: 1, Column: 0},
		Name: []string{"com", "example", "sub"},
	}

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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.ShapeStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "User",
				Simple: &ast.StringTypeRef{
					Pos: tokens.Position{Filename: "test.sentra", Line: 2, Column: 10},
				},
			},
			&ast.PolicyStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
				Name: "User", // Same name as shape - should cause conflict
				Statements: []ast.Statement{
					&ast.FactStatement{
						Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
						Name:  "user",
						Alias: "user",
						Type: &ast.StringTypeRef{
							Pos: tokens.Position{Filename: "test.sentra", Line: 4, Column: 10},
						},
						Default: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 20},
							Value: "testuser",
						},
					},
					&ast.RuleStatement{
						Pos:      tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
						RuleName: "allow",
						When: &ast.TrinaryLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 5, Column: 15},
							Value: 1, // true in trinary
						},
					},
					&ast.RuleExportStatement{
						Pos:         tokens.Position{Filename: "test.sentra", Line: 6, Column: 0},
						Of:          "allow",
						Attachments: []*ast.AttachmentClause{},
					},
				},
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.PolicyStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "AuthPolicy",
				Statements: []ast.Statement{
					&ast.FactStatement{
						Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
						Name:  "user",
						Alias: "user",
						Type: &ast.StringTypeRef{
							Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
						},
						Default: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 20},
							Value: "testuser",
						},
					},
					&ast.RuleStatement{
						Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
						RuleName: "allow",
						When: &ast.TrinaryLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
							Value: 1, // true in trinary
						},
					},
					// No rule export statement
				},
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.PolicyStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "AuthPolicy",
				Statements: []ast.Statement{
					&ast.FactStatement{
						Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
						Name:  "user",
						Alias: "user",
						Type: &ast.StringTypeRef{
							Pos: tokens.Position{Filename: "test.sentra", Line: 3, Column: 10},
						},
						Default: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 20},
							Value: "testuser",
						},
					},
					&ast.RuleStatement{
						Pos:      tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
						RuleName: "allow",
						When: &ast.TrinaryLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 15},
							Value: 1, // true in trinary
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
			},
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
			&ast.NamespaceStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 1, Column: 0},
				Name: ast.FQN{"com", "example"},
			},
			&ast.PolicyStatement{
				Pos:  tokens.Position{Filename: "test.sentra", Line: 2, Column: 0},
				Name: "AuthPolicy",
				Statements: []ast.Statement{
					&ast.RuleStatement{
						Pos:      tokens.Position{Filename: "test.sentra", Line: 3, Column: 0},
						RuleName: "allow",
						When: &ast.TrinaryLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 3, Column: 15},
							Value: 1, // true in trinary
						},
					},
					&ast.FactStatement{
						Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 0},
						Name:  "user",
						Alias: "user",
						Type: &ast.StringTypeRef{
							Pos: tokens.Position{Filename: "test.sentra", Line: 4, Column: 10},
						},
						Default: &ast.StringLiteral{
							Pos:   tokens.Position{Filename: "test.sentra", Line: 4, Column: 20},
							Value: "testuser",
						},
					},
					&ast.RuleExportStatement{
						Pos:         tokens.Position{Filename: "test.sentra", Line: 5, Column: 0},
						Of:          "allow",
						Attachments: []*ast.AttachmentClause{},
					},
				},
			},
		},
	}

	err := suite.idx.AddProgram(suite.ctx, program)

	suite.Error(err)
	suite.Contains(err.Error(), "fact statement must be the first statement in a policy")
}
