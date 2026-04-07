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

package index

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func pr(line int) tokens.Range {
	return tokens.Range{
		File: "cov.sentrie",
		From: tokens.Pos{Line: line, Column: 0, Offset: 0},
		To:   tokens.Pos{Line: line, Column: 1, Offset: 1},
	}
}

func (suite *IndexTestSuite) TestValidate_StringTypeStringMethod() {
	var s String = "node"
	suite.Equal("node", s.String())
}

func (suite *IndexTestSuite) TestValidate_IsValidSameAsValidate() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithRichRuleGraph(false)))
	err := suite.idx.IsValid(ctx)
	suite.NoError(err)
}

func (suite *IndexTestSuite) TestValidate_ContextCancelledInDetectReference() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	suite.Require().NoError(suite.idx.AddProgram(context.Background(), programWithRichRuleGraph(false)))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "validation cancelled")
}

func (suite *IndexTestSuite) TestValidate_ReferenceCycleSelfIdentifier() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithSelfReferencingRule()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "infinite recursion")
}

func (suite *IndexTestSuite) TestValidate_RuleImportCycleTwoPolicies() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithCyclicRuleImports()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "cyclic dependency in rules")
}

func (suite *IndexTestSuite) TestValidate_RuleImportMissingPolicy() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithBrokenRuleImport()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "error resolving policy")
}

func (suite *IndexTestSuite) TestValidate_RuleImportQualifiedFQN() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programTargetPolForQualifiedImport()))
	suite.Require().NoError(suite.idx.AddProgram(ctx, programConsumerQualifiedImport()))
	err := suite.idx.Validate(ctx)
	suite.NoError(err)
}

func (suite *IndexTestSuite) TestValidate_ShapePolicyLocalMissingBase() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithPolicyShapeMissingNamespaceBase()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "shape not found")
}

func (suite *IndexTestSuite) TestValidate_DetectShapeCycleNamespaceWith() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithNamespaceShapesCycleViaWith()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "cyclic dependencies in shapes")
}

func (suite *IndexTestSuite) TestValidate_RuleImportSelfLoopAddEdgeError() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithRuleImportSelfLoop()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "error adding edge")
}

func (suite *IndexTestSuite) TestValidate_ShapeComposeSelfLoopAddEdgeError() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programWithShapeComposeSelfLoop()))
	err := suite.idx.Validate(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "error adding edge")
}

func (suite *IndexTestSuite) TestValidate_ChildNamespaceShapeComposesParentShape() {
	ctx := context.Background()
	suite.Require().NoError(suite.idx.AddProgram(ctx, programParentShapeOnly()))
	suite.Require().NoError(suite.idx.AddProgram(ctx, programChildNamespaceComposesParentShape()))
	err := suite.idx.Validate(ctx)
	suite.NoError(err)
}

// programWithRichRuleGraph builds lets and a rule body that exercises addNodes switch cases (no cycle).
func programWithRichRuleGraph(selfCycle bool) *ast.Program {
	r := pr(1)
	id := func(name string, line int) *ast.Identifier {
		return ast.NewIdentifier(name, pr(line))
	}
	body := id("chainB", 20)
	if selfCycle {
		body = id("chainA", 20)
	}
	innerLet := ast.NewVarDeclaration(
		"innerLet",
		ast.NewStringTypeRef(pr(21)),
		id("chainC", 21),
		pr(21),
	)
	ruleInBlock := ast.NewRuleStatement("innerR", nil, nil, id("innerRBody", 23), pr(23))
	blockYield := ast.NewBlockExpression(
		[]ast.Statement{innerLet, ruleInBlock},
		id("chainD", 22),
		pr(22),
	)
	call := ast.NewCallExpression(
		id("calleeX", 30),
		[]ast.Expression{id("argA", 31), id("argB", 32)},
		false,
		nil,
		pr(30),
	)
	list := ast.NewListLiteral([]ast.Expression{id("le1", 40), id("le2", 41)}, pr(40))
	mapLit := ast.NewMapLiteral([]ast.MapEntry{
		{Key: ast.NewStringLiteral("k", pr(50)), Value: id("mapVal", 51)},
	}, pr(50))
	field := ast.NewFieldAccessExpression(id("objZ", 60), "f", pr(60))
	unary := ast.NewUnaryExpression("!", id("u1", 70), pr(70))
	tern := ast.NewTernaryExpression(
		id("cond1", 80),
		id("then1", 81),
		id("else1", 82),
		pr(80),
	)
	infix := ast.NewInfixExpression(id("leftI", 90), id("rightI", 91), "and", pr(90))
	complexBody := ast.NewInfixExpression(
		infix,
		ast.NewInfixExpression(
			tern,
			ast.NewInfixExpression(unary, field, "or", pr(92)),
			"and",
			pr(91),
		),
		"and",
		pr(93),
	)
	complexBody = ast.NewInfixExpression(complexBody, mapLit, "and", pr(94))
	complexBody = ast.NewInfixExpression(complexBody, list, "and", pr(95))
	complexBody = ast.NewInfixExpression(complexBody, call, "and", pr(96))
	complexBody = ast.NewInfixExpression(complexBody, blockYield, "and", pr(97))
	complexBody = ast.NewInfixExpression(complexBody, body, "and", pr(99))

	pol := ast.NewPolicyStatement(
		"GraphPol",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(pr(2)),
				"u",
				nil,
				true,
				pr(2),
			),
			ast.NewVarDeclaration(
				"toBRef",
				ast.NewStringTypeRef(pr(3)),
				id("chainB", 3),
				pr(3),
			),
			ast.NewRuleStatement(
				"chainA",
				nil,
				ast.NewTrinaryLiteral(trinary.True, pr(4)),
				complexBody,
				pr(4),
			),
			ast.NewRuleStatement(
				"chainB",
				nil,
				nil,
				ast.NewTrinaryLiteral(trinary.True, pr(5)),
				pr(5),
			),
			ast.NewRuleExportStatement("chainA", nil, pr(6)),
		},
		pr(1),
	)
	return &ast.Program{
		Reference: "cov.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(
				ast.NewFQN([]string{"com", "example"}, r),
				r,
			),
			pol,
		},
	}
}

func programWithSelfReferencingRule() *ast.Program {
	r := pr(1)
	return &ast.Program{
		Reference: "self.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"SelfPol",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement(
						"ra",
						nil,
						nil,
						ast.NewIdentifier("rb", pr(3)),
						pr(3),
					),
					ast.NewRuleStatement(
						"rb",
						nil,
						nil,
						ast.NewIdentifier("ra", pr(4)),
						pr(4),
					),
					ast.NewRuleExportStatement("ra", nil, pr(5)),
				},
				pr(1),
			),
		},
	}
}

func programWithCyclicRuleImports() *ast.Program {
	r := pr(1)
	importAToB := ast.NewImportClause(
		"rb",
		ast.NewFQN([]string{"PolB"}, r).Ptr(),
		nil,
		pr(10),
	)
	importBToA := ast.NewImportClause(
		"ra",
		ast.NewFQN([]string{"PolA"}, r).Ptr(),
		nil,
		pr(11),
	)
	return &ast.Program{
		Reference: "cyc.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"PolA",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("ra", nil, nil, importAToB, pr(3)),
					ast.NewRuleExportStatement("ra", nil, pr(4)),
				},
				pr(1),
			),
			ast.NewPolicyStatement(
				"PolB",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("rb", nil, nil, importBToA, pr(5)),
					ast.NewRuleExportStatement("rb", nil, pr(6)),
				},
				pr(1),
			),
		},
	}
}

func programWithBrokenRuleImport() *ast.Program {
	r := pr(1)
	badImport := ast.NewImportClause(
		"missingRule",
		ast.NewFQN([]string{"NoSuchPol"}, r).Ptr(),
		nil,
		pr(10),
	)
	return &ast.Program{
		Reference: "badimp.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"HasImp",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r1", nil, nil, badImport, pr(3)),
					ast.NewRuleExportStatement("r1", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programConsumerQualifiedImport() *ast.Program {
	r := pr(1)
	qualImport := ast.NewImportClause(
		"tr",
		ast.NewFQN([]string{"com", "other", "TargetPol"}, r).Ptr(),
		nil,
		pr(10),
	)
	return &ast.Program{
		Reference: "qual-consumer.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"Consumer",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("cr", nil, nil, qualImport, pr(3)),
					ast.NewRuleExportStatement("cr", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programTargetPolForQualifiedImport() *ast.Program {
	r := pr(1)
	return &ast.Program{
		Reference: "qual-target.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "other"}, r), r),
			ast.NewPolicyStatement(
				"TargetPol",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("tr", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(5)), pr(5)),
					ast.NewRuleExportStatement("tr", nil, pr(6)),
				},
				r,
			),
		},
	}
}

func programWithPolicyShapeMissingNamespaceBase() *ast.Program {
	r := pr(1)
	withMissing := ast.NewFQN([]string{"NoSuchNsShape"}, r).Ptr()
	polShape := ast.NewShapeStatement(
		"LocalBad",
		nil,
		&ast.Cmplx{
			Range:  pr(5),
			With:   withMissing,
			Fields: map[string]*ast.ShapeField{},
		},
		pr(5),
	)
	return &ast.Program{
		Reference: "polshape.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"PS",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					polShape,
					ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(3)), pr(3)),
					ast.NewRuleExportStatement("r", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programWithRuleImportSelfLoop() *ast.Program {
	r := pr(1)
	selfImp := ast.NewImportClause(
		"r1",
		ast.NewFQN([]string{"LoopPol"}, r).Ptr(),
		nil,
		pr(10),
	)
	return &ast.Program{
		Reference: "loopr.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			ast.NewPolicyStatement(
				"LoopPol",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r1", nil, nil, selfImp, pr(3)),
					ast.NewRuleExportStatement("r1", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programWithShapeComposeSelfLoop() *ast.Program {
	r := pr(1)
	withSelf := ast.NewFQN([]string{"SelfShape"}, r).Ptr()
	selfShape := ast.NewShapeStatement(
		"SelfShape",
		nil,
		&ast.Cmplx{
			Range:  pr(10),
			With:   withSelf,
			Fields: map[string]*ast.ShapeField{},
		},
		pr(10),
	)
	return &ast.Program{
		Reference: "loops.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			selfShape,
			ast.NewPolicyStatement(
				"P",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(3)), pr(3)),
					ast.NewRuleExportStatement("r", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programParentShapeOnly() *ast.Program {
	r := pr(1)
	parentShape := ast.NewShapeStatement(
		"ParentS",
		nil,
		&ast.Cmplx{
			Range: r,
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"p": {Range: pr(5), Name: "p", NotNullable: true, Required: true, Type: ast.NewStringTypeRef(pr(5))},
			},
		},
		pr(4),
	)
	return &ast.Program{
		Reference: "parentshape.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "ex"}, r), r),
			parentShape,
			ast.NewPolicyStatement(
				"PP",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(3)), pr(3)),
					ast.NewRuleExportStatement("r", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programChildNamespaceComposesParentShape() *ast.Program {
	r := pr(1)
	withParent := ast.NewFQN([]string{"com", "ex", "ParentS"}, r).Ptr()
	kid := ast.NewShapeStatement(
		"Kid",
		nil,
		&ast.Cmplx{
			Range:  pr(10),
			With:   withParent,
			Fields: map[string]*ast.ShapeField{},
		},
		pr(10),
	)
	return &ast.Program{
		Reference: "childshape.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "ex", "sub"}, r), r),
			kid,
			ast.NewPolicyStatement(
				"CP",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(3)), pr(3)),
					ast.NewRuleExportStatement("r", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func programWithNamespaceShapesCycleViaWith() *ast.Program {
	r := pr(1)
	sA := ast.NewShapeStatement(
		"NSA",
		nil,
		&ast.Cmplx{
			Range: pr(10),
			With:  ast.NewFQN([]string{"NSB"}, r).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"a": {Range: pr(11), Name: "a", NotNullable: true, Required: true, Type: ast.NewStringTypeRef(pr(11))},
			},
		},
		pr(10),
	)
	sB := ast.NewShapeStatement(
		"NSB",
		nil,
		&ast.Cmplx{
			Range: pr(20),
			With:  ast.NewFQN([]string{"NSA"}, r).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"b": {Range: pr(21), Name: "b", NotNullable: true, Required: true, Type: ast.NewStringTypeRef(pr(21))},
			},
		},
		pr(20),
	)
	return &ast.Program{
		Reference: "nsshapecyc.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, r), r),
			sA,
			sB,
			ast.NewPolicyStatement(
				"P",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(pr(2)), "u", nil, true, pr(2)),
					ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, pr(3)), pr(3)),
					ast.NewRuleExportStatement("r", nil, pr(4)),
				},
				pr(1),
			),
		},
	}
}

func TestProgramWithRichRuleGraphBuilds(t *testing.T) {
	p := programWithRichRuleGraph(false)
	require.NotNil(t, p)
}
