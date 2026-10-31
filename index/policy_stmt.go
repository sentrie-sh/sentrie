// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// policyStmtKind classifies policy body statements. Register new kinds here when adding
// policy statements; createPolicy still uses a concrete type switch for phase rules -
// policyStmtKindOf is used to skip comments, and the is* helpers are mainly for tests.
type policyStmtKind int

const (
	policyStmtComment policyStmtKind = iota
	policyStmtMetadata
	policyStmtFact
	policyStmtUse
	policyStmtBody
	policyStmtUnknown
)

func policyStmtKindOf(stmt ast.Statement) policyStmtKind {
	switch stmt.(type) {
	case *ast.CommentStatement:
		return policyStmtComment
	case *ast.TitleStatement, *ast.DescriptionStatement, *ast.VersionStatement, *ast.TagStatement:
		return policyStmtMetadata
	case *ast.FactStatement:
		return policyStmtFact
	case *ast.UseStatement:
		return policyStmtUse
	case *ast.VarDeclaration, *ast.RuleStatement, *ast.ContributeStatement, *ast.RuleExportStatement, *ast.ShapeStatement, *ast.DeriveStatement:
		return policyStmtBody
	default:
		return policyStmtUnknown
	}
}

func isMetadataStmt(stmt ast.Statement) bool { return policyStmtKindOf(stmt) == policyStmtMetadata }
func isFactStmt(stmt ast.Statement) bool     { return policyStmtKindOf(stmt) == policyStmtFact }
func isUseStmt(stmt ast.Statement) bool      { return policyStmtKindOf(stmt) == policyStmtUse }
func isBodyStmt(stmt ast.Statement) bool     { return policyStmtKindOf(stmt) == policyStmtBody }

// buildTagsByKey groups TagPairs by key (append order preserved per key). Returns nil when pairs is empty.
func buildTagsByKey(pairs []PolicyTagPair) map[string][]string {
	if len(pairs) == 0 {
		return nil
	}
	out := make(map[string][]string)
	for _, pair := range pairs {
		out[pair.Key] = append(out[pair.Key], pair.Value)
	}
	return out
}
