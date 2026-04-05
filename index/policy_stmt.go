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
)

// policyStmtKind classifies policy body statements for header phase rules.
// New statement kinds that belong in the header or body should be registered here.
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
	case *ast.VarDeclaration, *ast.RuleStatement, *ast.RuleExportStatement, *ast.ShapeStatement:
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
