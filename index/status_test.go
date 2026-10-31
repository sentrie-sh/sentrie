// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"sync"

	"github.com/sentrie-sh/sentrie/ast"
)

func (suite *IndexTestSuite) TestStatus_FreshIndex() {
	st := CreateIndex().Status()
	suite.Equal(Status{}, st)
}

func (suite *IndexTestSuite) TestStatus_DoesNotValidate() {
	idx := CreateIndex()
	suite.Require().NoError(idx.AddProgram(suite.ctx, programWithRichRuleGraph(false)))
	st := idx.Status()
	suite.False(st.Validated)
	suite.False(st.Committed)
	suite.Equal(1, st.Programs)
	suite.Equal(uint32(0), idx.validated, "taking a status must not run validation")
}

func (suite *IndexTestSuite) TestStatus_ValidatedCountsDefinitions() {
	idx := CreateIndex()
	suite.Require().NoError(idx.AddProgram(suite.ctx, programWithRichRuleGraph(false)))
	suite.Require().NoError(idx.Validate(suite.ctx))

	st := idx.Status()
	suite.True(st.Validated)
	suite.True(st.Committed)
	suite.NoError(st.ValidationError)
	suite.NoError(st.CommitError)
	suite.Equal(len(idx.Namespaces), st.Namespaces)
	suite.Equal(len(idx.Programs), st.Programs)
	var policies, rules, shapes int
	for _, ns := range idx.Namespaces {
		policies += len(ns.Policies)
		shapes += len(ns.Shapes)
		for _, p := range ns.Policies {
			rules += len(p.Rules)
			shapes += len(p.Shapes)
		}
	}
	suite.Equal(policies, st.Policies)
	suite.Equal(rules, st.Rules)
	suite.Equal(shapes, st.Shapes)
	suite.Positive(st.Rules)
}

func (suite *IndexTestSuite) TestStatus_ValidationFailure() {
	idx := CreateIndex()
	suite.Require().NoError(idx.AddProgram(suite.ctx, programWithCyclicRuleImports()))
	suite.Error(idx.Validate(suite.ctx))

	st := idx.Status()
	suite.True(st.Validated, "validation finished, with an error")
	suite.False(st.Committed, "a failed validation never commits")
	suite.Error(st.ValidationError)
	suite.Contains(st.ValidationError.Error(), "cyclic dependency in rules")
	suite.NoError(st.CommitError)
}

func (suite *IndexTestSuite) TestStatus_SafeDuringValidate() {
	idx := CreateIndex()
	suite.Require().NoError(idx.AddProgram(suite.ctx, programWithRichRuleGraph(false)))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = idx.Validate(suite.ctx)
	}()
	go func() {
		defer wg.Done()
		for range 100 {
			_ = idx.Status()
		}
	}()
	wg.Wait()
}

func (suite *IndexTestSuite) TestStatus_CommitFailure() {
	rng := testRange()
	ns := testNamespace("n")
	// An alias shape cannot be composed with: validation passes, commit fails.
	aliasStmt := ast.NewShapeStatement("Alias", ast.NewStringTypeRef(rng), nil, rng)
	alias, err := createShape(ns, nil, aliasStmt)
	suite.Require().NoError(err)
	composedStmt := ast.NewShapeStatement("Composed", nil, &ast.Cmplx{
		Range:  rng,
		With:   ast.NewFQN([]string{"Alias"}, rng).Ptr(),
		Fields: map[string]*ast.ShapeField{},
	}, rng)
	composed, err := createShape(ns, nil, composedStmt)
	suite.Require().NoError(err)
	ns.Shapes["Alias"] = alias
	ns.Shapes["Composed"] = composed
	idx := CreateIndex()
	idx.Namespaces[ns.FQN.String()] = ns

	suite.Error(idx.Validate(suite.ctx))

	st := idx.Status()
	suite.True(st.Validated)
	suite.True(st.Committed, "the commit finished, with an error")
	suite.Error(st.CommitError)
	suite.Contains(st.CommitError.Error(), "commit error")
	suite.Same(st.CommitError, st.ValidationError, "Validate records a commit failure as its own error")
	suite.Equal(2, st.Shapes)
}
