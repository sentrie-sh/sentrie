// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/suite"
)

// IndexTestSuite is the single suite for all index package tests.
type IndexTestSuite struct {
	suite.Suite
	ctx context.Context
	idx *Index
	// policyNs is the namespace fixture for policy unit tests.
	policyNs *Namespace
	parentNs *Namespace
	childNs  *Namespace
}

func (suite *IndexTestSuite) SetupSuite() {
	suite.ctx = suite.T().Context()
}

// BeforeTest dispatches per-method fixtures (segments index graph, policy namespace,
// namespace hierarchy, or a fresh CreateIndex for integration tests).
func (suite *IndexTestSuite) BeforeTest(suiteName, testName string) {
	suite.ctx = suite.T().Context()
	suite.idx = nil
	suite.policyNs = nil
	suite.parentNs = nil
	suite.childNs = nil

	if strings.HasPrefix(testName, "TestShapeDependency") {
		return
	}
	if strings.HasPrefix(testName, "TestResolveSegments") {
		suite.idx = CreateIndex()
		suite.setupSegmentsIndexFixture()
		return
	}
	if policyFixtureTests[testName] {
		nsStmt := ast.NewNamespaceStatement(
			ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
		)
		suite.policyNs = createNamespace(nsStmt)
		return
	}
	if namespaceFixtureTests[testName] {
		parentStmt := ast.NewNamespaceStatement(
			ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "parent.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			tokens.Range{File: "parent.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
		)
		suite.parentNs = createNamespace(parentStmt)
		childStmt := ast.NewNamespaceStatement(
			ast.NewFQN([]string{"com", "example", "sub"}, tokens.Range{File: "child.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
			tokens.Range{File: "child.sentrie", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
		)
		suite.childNs = createNamespace(childStmt)
		return
	}
	suite.idx = CreateIndex()
}

func TestIndexTestSuite(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}

var policyFixtureTests = map[string]bool{
	"TestCreatePolicy":                        true,
	"TestCreatePolicyWithoutExports":          true,
	"TestCreatePolicyWithInvalidFactPosition": true,
	"TestCreatePolicyWithInvalidUsePosition":  true,
	"TestCreatePolicyWithUnknownRuleExport":   true,
	"TestCreatePolicyWithDuplicateRuleExport": true,
	"TestAddLet":                                             true,
	"TestAddLetWithNameConflict":                             true,
	"TestAddRule":                                            true,
	"TestAddRuleWithNameConflict":                            true,
	"TestPolicyAddShape":                                     true,
	"TestPolicyAddShapeWithNameConflict":                     true,
	"TestAddFact":                                            true,
	"TestAddFactRequiredCannotHaveDefault":                   true,
	"TestAddFactOptionalCanHaveDefault":                      true,
	"TestAddFactRequiredWithoutDefault":                      true,
	"TestAddFactWithNameConflict":                            true,
	"TestPolicyString":                                       true,
	"TestCreatePolicyWithComments":                           true,
	"TestCreatePolicyWithValidUseStatement":                  true,
	"TestCreatePolicyWithMetadataAndPhases":                  true,
	"TestCreatePolicyUseWithoutFacts":                        true,
	"TestCreatePolicyMetadataThenUseWithoutFacts":            true,
	"TestCreatePolicyVersionVPrefixAccepted":                 true,
	"TestCreatePolicyFactAfterUseErrors":                     true,
	"TestCreatePolicyMetadataAfterFactErrors":                true,
	"TestCreatePolicyShapeBeforeFactErrors":                  true,
	"TestCreatePolicyDuplicateTitle":                         true,
	"TestCreatePolicyInvalidSemVer":                          true,
	"TestCreatePolicyEmptyTitle":                             true,
	"TestCreatePolicyEmptyTagKey":                            true,
	"TestCreatePolicyVersionWhitespacePaddedLiteralAccepted": true,
	"TestCreatePolicyUnsupportedShapeExportInBody":           true,
	"TestCreatePolicyDuplicateUseAliasRebind":                true,
	"TestCreatePolicyDuplicateRuleExportAttachmentName":      true,
	"TestPolicyAddShapeDuplicateInnerFieldName":              true,
	"TestCreatePolicyTitleAfterRuleLateHeader":               true,
	"TestCreatePolicyDuplicateDescription":                   true,
	"TestCreatePolicyDuplicateVersion":                       true,
	"TestCreatePolicyTagAfterUseMetadataContiguous":          true,
	"TestCreatePolicyTwoFactsSecondPhaseFactsCase":           true,
	"TestCreatePolicyRequiredFactWithDefaultAddFactError":    true,
	"TestCreatePolicyDuplicateLetName":                       true,
	"TestCreatePolicyDuplicateRuleName":                      true,
	"TestCreatePolicyDescriptionAfterRuleLateHeader":         true,
	"TestCreatePolicyVersionAfterRuleLateHeader":             true,
	"TestCreatePolicyTagAfterRuleLateHeader":                 true,
	"TestCreatePolicyExportAfterUseBeforeRule":               true,
	"TestCreatePolicyShapeBodyDuplicateFieldErrors":          true,
	"TestCreatePolicyDescriptionAfterFactContiguousMetadata": true,
	"TestCreatePolicyVersionAfterFactContiguousMetadata":     true,
}

var namespaceFixtureTests = map[string]bool{
	"TestCreateNamespace":                true,
	"TestAddChild":                       true,
	"TestAddChildWithNameConflict":       true,
	"TestCheckNameAvailable":             true,
	"TestAddPolicy":                      true,
	"TestAddPolicyWithNameConflict":      true,
	"TestAddShape":                       true,
	"TestAddShapeWithNameConflict":       true,
	"TestAddShapeExport":                 true,
	"TestAddShapeExportWithNameConflict": true,
	"TestIsChildOf":                      true,
	"TestIsParentOf":                     true,
	"TestComplexHierarchy":               true,
	"TestMultipleChildren":               true,
}
