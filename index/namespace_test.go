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
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/suite"
)

type NamespaceTestSuite struct {
	suite.Suite
	parentNs *Namespace
	childNs  *Namespace
}

func (suite *NamespaceTestSuite) SetupTest() {
	// Create parent namespace
	parentStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "parent.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "parent.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	suite.parentNs = createNamespace(parentStmt)

	// Create child namespace
	childStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub"}, tokens.Range{File: "child.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "child.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	suite.childNs = createNamespace(childStmt)
}

func (suite *NamespaceTestSuite) TearDownTest() {
	suite.parentNs = nil
	suite.childNs = nil
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}

func (suite *NamespaceTestSuite) TestCreateNamespace() {
	stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "test"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	ns := createNamespace(stmt)

	suite.NotNil(ns)
	suite.Equal(stmt, ns.Statement)
	suite.Equal("com/example/test", ns.FQN.String())
	suite.Nil(ns.Parent)
	suite.NotNil(ns.Children)
	suite.Len(ns.Children, 0)
	suite.NotNil(ns.Policies)
	suite.Len(ns.Policies, 0)
	suite.NotNil(ns.Shapes)
	suite.Len(ns.Shapes, 0)
	suite.NotNil(ns.ShapeExports)
	suite.Len(ns.ShapeExports, 0)
}

func (suite *NamespaceTestSuite) TestAddChild() {
	err := suite.parentNs.addChild(suite.childNs)

	suite.NoError(err)
	suite.Len(suite.parentNs.Children, 1)
	suite.Equal(suite.childNs, suite.parentNs.Children[0])
	suite.Equal(suite.parentNs, suite.childNs.Parent)
}

func (suite *NamespaceTestSuite) TestAddChildWithNameConflict() {
	// Create a policy with the same name as the child namespace
	policyStmt := ast.NewPolicyStatement(
		"sub", // Same as child namespace last segment
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"allow",
				[]*ast.AttachmentClause{},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.parentNs, policyStmt, program)
	suite.NoError(err)

	err = suite.parentNs.addPolicy(policy)
	suite.NoError(err)

	// Now try to add child with conflicting name
	err = suite.parentNs.addChild(suite.childNs)

	suite.Error(err)
	suite.Contains(err.Error(), "conflict: policy declaration")
}

func (suite *NamespaceTestSuite) TestCheckNameAvailable() {
	// Test with no conflicts
	err := suite.parentNs.checkNameAvailable("testName")
	suite.NoError(err)

	// Add a policy
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"allow",
				[]*ast.AttachmentClause{},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.parentNs, policyStmt, program)
	suite.NoError(err)

	err = suite.parentNs.addPolicy(policy)
	suite.NoError(err)

	// Test conflict with policy name
	err = suite.parentNs.checkNameAvailable("testPolicy")
	suite.Error(err)
	suite.Contains(err.Error(), "conflict: policy declaration")

	// Add a shape
	shapeStmt := ast.NewShapeStatement(
		"testShape",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	shape, err := createShape(suite.parentNs, nil, shapeStmt)
	suite.NoError(err)

	err = suite.parentNs.addShape(shape)
	suite.NoError(err)

	// Test conflict with shape name
	err = suite.parentNs.checkNameAvailable("testShape")
	suite.Error(err)
	suite.Contains(err.Error(), "conflict: shape declaration")

	// Add a child namespace
	err = suite.parentNs.addChild(suite.childNs)
	suite.NoError(err)

	// Test conflict with child namespace name
	err = suite.parentNs.checkNameAvailable("sub")
	suite.Error(err)
	suite.Contains(err.Error(), "conflict: namespace declaration")
}

func (suite *NamespaceTestSuite) TestAddPolicy() {
	policyStmt := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"allow",
				[]*ast.AttachmentClause{},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt,
		},
	}

	policy, err := createPolicy(suite.parentNs, policyStmt, program)
	suite.NoError(err)

	err = suite.parentNs.addPolicy(policy)

	suite.NoError(err)
	suite.Len(suite.parentNs.Policies, 1)
	suite.Contains(suite.parentNs.Policies, "testPolicy")
	suite.Equal(policy, suite.parentNs.Policies["testPolicy"])
}

func (suite *NamespaceTestSuite) TestAddPolicyWithNameConflict() {
	// Add first policy
	policyStmt1 := ast.NewPolicyStatement(
		"testPolicy",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				"user",
				nil,
				false, // required
				tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"allow",
				nil,
				ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"allow",
				[]*ast.AttachmentClause{},
				tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	program1 := &ast.Program{
		Reference: "test1.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt1,
		},
	}

	policy1, err := createPolicy(suite.parentNs, policyStmt1, program1)
	suite.NoError(err)

	err = suite.parentNs.addPolicy(policy1)
	suite.NoError(err)

	// Try to add second policy with same name
	policyStmt2 := ast.NewPolicyStatement(
		"testPolicy", // Same name
		[]ast.Statement{
			ast.NewFactStatement(
				"admin",
				ast.NewStringTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				"admin",
				nil,
				false, // required
				tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement(
				"deny",
				nil,
				ast.NewTrinaryLiteral(trinary.False, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 3, Column: 15, Offset: 15}, To: tokens.Pos{Line: 3, Column: 15, Offset: 15}}),
				nil,
				tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleExportStatement(
				"deny",
				[]*ast.AttachmentClause{},
				tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}},
			),
		},
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	program2 := &ast.Program{
		Reference: "test2.sentra",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}), tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			policyStmt2,
		},
	}

	policy2, err := createPolicy(suite.parentNs, policyStmt2, program2)
	suite.NoError(err)

	err = suite.parentNs.addPolicy(policy2)

	suite.Error(err)
	suite.Contains(err.Error(), "conflict: policy declaration")
}

func (suite *NamespaceTestSuite) TestAddShape() {
	shapeStmt := ast.NewShapeStatement(
		"testShape",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	shape, err := createShape(suite.parentNs, nil, shapeStmt)
	suite.NoError(err)

	err = suite.parentNs.addShape(shape)

	suite.NoError(err)
	suite.Len(suite.parentNs.Shapes, 1)
	suite.Contains(suite.parentNs.Shapes, "testShape")
	suite.Equal(shape, suite.parentNs.Shapes["testShape"])
}

func (suite *NamespaceTestSuite) TestAddShapeWithNameConflict() {
	// Add first shape
	shapeStmt1 := ast.NewShapeStatement(
		"testShape",
		ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	shape1, err := createShape(suite.parentNs, nil, shapeStmt1)
	suite.NoError(err)

	err = suite.parentNs.addShape(shape1)
	suite.NoError(err)

	// Try to add second shape with same name
	shapeStmt2 := ast.NewShapeStatement(
		"testShape", // Same name
		ast.NewNumberTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 15, Offset: 15}, To: tokens.Pos{Line: 1, Column: 15, Offset: 15}}),
		nil,
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	shape2, err := createShape(suite.parentNs, nil, shapeStmt2)
	suite.NoError(err)

	err = suite.parentNs.addShape(shape2)

	suite.Error(err)
	suite.Contains(err.Error(), "conflict: shape declaration")
}

func (suite *NamespaceTestSuite) TestAddShapeExport() {
	exportStmt := ast.NewShapeExportStatement(
		"testShape",
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	export := &ExportedShape{
		Statement: exportStmt,
		Name:      "testShape",
	}

	err := suite.parentNs.addShapeExport(export)

	suite.NoError(err)
	suite.Len(suite.parentNs.ShapeExports, 1)
	suite.Contains(suite.parentNs.ShapeExports, "testShape")
	suite.Equal(export, suite.parentNs.ShapeExports["testShape"])
}

func (suite *NamespaceTestSuite) TestAddShapeExportWithNameConflict() {
	// Add first export
	exportStmt1 := ast.NewShapeExportStatement(
		"testShape",
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	export1 := &ExportedShape{
		Statement: exportStmt1,
		Name:      "testShape",
	}

	err := suite.parentNs.addShapeExport(export1)
	suite.NoError(err)

	// Try to add second export with same name
	exportStmt2 := ast.NewShapeExportStatement(
		"testShape", // Same name
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	export2 := &ExportedShape{
		Statement: exportStmt2,
		Name:      "testShape",
	}

	err = suite.parentNs.addShapeExport(export2)

	suite.Error(err)
	suite.Contains(err.Error(), "conflict: shape export")
}

func (suite *NamespaceTestSuite) TestIsChildOf() {
	// Test parent-child relationship
	suite.True(suite.childNs.IsChildOf(suite.parentNs))
	suite.False(suite.parentNs.IsChildOf(suite.childNs))

	// Test with unrelated namespaces
	unrelatedStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"org", "different"}, tokens.Range{File: "unrelated.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "unrelated.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	unrelatedNs := createNamespace(unrelatedStmt)

	suite.False(suite.childNs.IsChildOf(unrelatedNs))
	suite.False(unrelatedNs.IsChildOf(suite.parentNs))
}

func (suite *NamespaceTestSuite) TestIsParentOf() {
	// Test parent-child relationship
	suite.True(suite.parentNs.IsParentOf(suite.childNs))
	suite.False(suite.childNs.IsParentOf(suite.parentNs))

	// Test with unrelated namespaces
	unrelatedStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"org", "different"}, tokens.Range{File: "unrelated.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "unrelated.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	unrelatedNs := createNamespace(unrelatedStmt)

	suite.False(suite.parentNs.IsParentOf(unrelatedNs))
	suite.False(unrelatedNs.IsParentOf(suite.childNs))
}

func (suite *NamespaceTestSuite) TestComplexHierarchy() {
	// Create a complex hierarchy: com.example -> com.example.sub -> com.example.sub.deep
	grandchildStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub", "deep"}, tokens.Range{File: "grandchild.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "grandchild.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	grandchildNs := createNamespace(grandchildStmt)

	// Add child to parent
	err := suite.parentNs.addChild(suite.childNs)
	suite.NoError(err)

	// Add grandchild to child
	err = suite.childNs.addChild(grandchildNs)
	suite.NoError(err)

	// Test relationships
	suite.True(suite.childNs.IsChildOf(suite.parentNs))
	suite.True(grandchildNs.IsChildOf(suite.childNs))
	// grandchildNs is not a direct child of parentNs (it's a grandchild)
	suite.False(grandchildNs.IsChildOf(suite.parentNs))

	suite.True(suite.parentNs.IsParentOf(suite.childNs))
	suite.True(suite.childNs.IsParentOf(grandchildNs))
	// parentNs is not a direct parent of grandchildNs (it's a grandparent)
	suite.False(suite.parentNs.IsParentOf(grandchildNs))

	// Test that grandchild is not child of itself
	suite.False(grandchildNs.IsChildOf(grandchildNs))
	suite.False(grandchildNs.IsParentOf(grandchildNs))
}

func (suite *NamespaceTestSuite) TestMultipleChildren() {
	// Create multiple children
	child2Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub2"}, tokens.Range{File: "child2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "child2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	child2Ns := createNamespace(child2Stmt)

	child3Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub3"}, tokens.Range{File: "child3.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "child3.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	child3Ns := createNamespace(child3Stmt)

	// Add all children
	err := suite.parentNs.addChild(suite.childNs)
	suite.NoError(err)

	err = suite.parentNs.addChild(child2Ns)
	suite.NoError(err)

	err = suite.parentNs.addChild(child3Ns)
	suite.NoError(err)

	// Verify all children are added
	suite.Len(suite.parentNs.Children, 3)
	suite.Contains(suite.parentNs.Children, suite.childNs)
	suite.Contains(suite.parentNs.Children, child2Ns)
	suite.Contains(suite.parentNs.Children, child3Ns)

	// Verify parent relationships
	suite.Equal(suite.parentNs, suite.childNs.Parent)
	suite.Equal(suite.parentNs, child2Ns.Parent)
	suite.Equal(suite.parentNs, child3Ns.Parent)
}
