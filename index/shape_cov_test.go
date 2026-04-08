// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package index

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

func (s *IndexTestSuite) TestShapeResolveDependencyFailsWhenComposedShapeNotExported_Cov() {
	idx := CreateIndex()

	currentNS := testNamespace("consumer")
	sourceNS := testNamespace("source")
	otherNS := testNamespace("backing")
	idx.Namespaces[currentNS.FQN.String()] = currentNS
	idx.Namespaces[sourceNS.FQN.String()] = sourceNS
	idx.Namespaces[otherNS.FQN.String()] = otherNS

	// Intentionally make ResolveShape return a shape whose Namespace differs from
	// the namespace map entry to exercise the not-exported cross-namespace guard.
	base := testShape(otherNS, nil, "base", nil)
	sourceNS.Shapes[base.Name] = base

	derived := testShape(currentNS, nil, "derived", ast.NewFQN([]string{"base"}, testRange()).Ptr())
	currentNS.Shapes[derived.Name] = derived

	err := derived.resolveDependency(idx, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "not exported")
}

func (s *IndexTestSuite) TestShapeResolveDependencyNotExportedMarksHydrated_Cov() {
	idx := CreateIndex()

	currentNS := testNamespace("consumer")
	sourceNS := testNamespace("source")
	idx.Namespaces[currentNS.FQN.String()] = currentNS
	idx.Namespaces[sourceNS.FQN.String()] = sourceNS

	base := testShape(sourceNS, nil, "base", nil)
	sourceNS.Shapes[base.Name] = base
	derived := testShape(currentNS, nil, "derived", ast.NewFQN([]string{"base"}, testRange()).Ptr())
	currentNS.Shapes[derived.Name] = derived

	_ = derived.resolveDependency(idx, nil)
	s.Require().NoError(derived.resolveDependency(idx, nil))
}

func (s *IndexTestSuite) TestShapeResolveDependencyInValidationFlow_Cov() {
	idx := CreateIndex()
	currentNS := testNamespace("consumer")
	sourceNS := testNamespace("source")
	idx.Namespaces[currentNS.FQN.String()] = currentNS
	idx.Namespaces[sourceNS.FQN.String()] = sourceNS
	sourceNS.Shapes["base"] = testShape(sourceNS, nil, "base", nil)
	currentNS.Shapes["derived"] = testShape(currentNS, nil, "derived", ast.NewFQN([]string{"base"}, testRange()).Ptr())

	s.Require().Error(idx.Validate(context.Background()))
}
