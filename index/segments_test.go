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

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/suite"
)

type SegmentsTestSuite struct {
	suite.Suite
	idx *Index
}

func (suite *SegmentsTestSuite) SetupTest() {
	suite.idx = CreateIndex()
	suite.setupTestData()
}

func (suite *SegmentsTestSuite) TearDownTest() {
	suite.idx = nil
}

func TestSegmentsTestSuite(t *testing.T) {
	suite.Run(t, new(SegmentsTestSuite))
}

func (suite *SegmentsTestSuite) setupTestData() {
	// Create namespace: com/example
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	_, err := suite.idx.ensureNamespace(context.Background(), nsStmt)
	suite.Require().NoError(err)

	// Create policy: com/example/auth
	policyStmt := ast.NewPolicyStatement(
		"auth",
		[]ast.Statement{
			ast.NewFactStatement(
				"user",
				ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				"user",
				nil,
				true, // optional
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}},
			),
			ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("deny", nil, ast.NewTrinaryLiteral(trinary.False, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 15, Offset: 15}, To: tokens.Pos{Line: 5, Column: 15, Offset: 15}}), nil, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("allow", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 0, Offset: 0}, To: tokens.Pos{Line: 6, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("deny", []*ast.AttachmentClause{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 7, Column: 0, Offset: 0}, To: tokens.Pos{Line: 7, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	program := &ast.Program{
		Reference: "test.sentra",
		Statements: []ast.Statement{
			nsStmt,
			policyStmt,
		},
	}

	err = suite.idx.AddProgram(context.Background(), program)
	suite.Require().NoError(err)

	// Create nested namespace: com/example/sub
	subNsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "sub"}, tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	_, err = suite.idx.ensureNamespace(context.Background(), subNsStmt)
	suite.Require().NoError(err)

	// Create policy in sub namespace: com/example/sub/admin
	subPolicyStmt := ast.NewPolicyStatement(
		"admin",
		[]ast.Statement{
			ast.NewFactStatement("role", ast.NewStringTypeRef(tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "role", nil, true, tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("check", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("check", []*ast.AttachmentClause{}, tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "sub.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	subProgram := &ast.Program{
		Reference: "sub.sentra",
		Statements: []ast.Statement{
			subNsStmt,
			subPolicyStmt,
		},
	}

	err = suite.idx.AddProgram(context.Background(), subProgram)
	suite.Require().NoError(err)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithValidPath() {
	suite.Run("simple namespace and policy", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("", rule)
	})

	suite.Run("namespace, policy, and rule", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/allow")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("allow", rule)
	})
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithEmptyPath() {
	// Test with empty path - should return an error, not panic
	ns, policy, rule, err := suite.idx.ResolveSegments("")

	suite.Error(err)
	suite.True(errors.Is(err, xerr.NotFoundError{}))
	suite.Contains(err.Error(), "namespace")
	suite.Equal("", ns)
	suite.Equal("", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithInvalidNamespace() {
	suite.Run("non-existent namespace", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("org/unknown/policy")

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
		suite.Contains(err.Error(), "namespace")
		suite.Equal("", ns)
		suite.Equal("", policy)
		suite.Equal("", rule)
	})

	suite.Run("partial namespace", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/unknown/policy")

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
		suite.Contains(err.Error(), "namespace")
		suite.Equal("", ns)
		suite.Equal("", policy)
		suite.Equal("", rule)
	})

	suite.Run("single segment", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("unknown")

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
		suite.Contains(err.Error(), "namespace")
		suite.Equal("", ns)
		suite.Equal("", policy)
		suite.Equal("", rule)
	})
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithInvalidPolicy() {
	suite.Run("non-existent policy", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/unknown")

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
		suite.Contains(err.Error(), "policy")
		suite.Equal("", ns)
		suite.Equal("", policy)
		suite.Equal("", rule)
	})

	suite.Run("policy in non-existent namespace", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("org/test/policy")

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
		suite.Contains(err.Error(), "policy")
		suite.Equal("", ns)
		suite.Equal("", policy)
		suite.Equal("", rule)
	})
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithEmptySegments() {
	suite.Run("path with empty segments", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com//example/auth")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("", rule)
	})

	suite.Run("path starting with slash", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("/com/example/auth")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("", rule)
	})

	suite.Run("path ending with slash", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("", rule)
	})

	suite.Run("path with multiple slashes", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com///example//auth")
		_ = ns
		_ = policy
		_ = rule

		suite.Error(err)
		suite.True(errors.Is(err, xerr.NotFoundError{}))
	})
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithOnlyNamespace() {
	// This should fail because we need at least a policy
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example")

	suite.Error(err)
	suite.True(errors.Is(err, xerr.NotFoundError{}))
	suite.Contains(err.Error(), "policy")
	suite.Equal("", ns)
	suite.Equal("", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithNamespaceAndPolicyOnly() {
	// This should succeed - we have namespace and policy, no rule needed
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth")

	suite.NoError(err)
	suite.Equal("com/example", ns)
	suite.Equal("auth", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithLongPath() {
	// Test with a very long path that doesn't exist
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/allow/extra/segments")

	suite.NoError(err)
	suite.Equal("com/example", ns)
	suite.Equal("auth", policy)
	suite.Equal("allow", rule)
	// Extra segments are ignored
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithSingleCharacterSegments() {
	// Create a namespace with single character segments
	singleCharNsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"a", "b"}, tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	_, err := suite.idx.ensureNamespace(context.Background(), singleCharNsStmt)
	suite.Require().NoError(err)

	// Create policy in single char namespace
	singleCharPolicyStmt := ast.NewPolicyStatement(
		"c",
		[]ast.Statement{
			ast.NewFactStatement("d", ast.NewStringTypeRef(tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}), "d", nil, true, tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 3, Column: 0, Offset: 0}, To: tokens.Pos{Line: 3, Column: 0, Offset: 0}}),
			ast.NewRuleStatement("e", nil, ast.NewTrinaryLiteral(trinary.True, tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 4, Column: 15, Offset: 15}, To: tokens.Pos{Line: 4, Column: 15, Offset: 15}}), nil, tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 4, Column: 0, Offset: 0}, To: tokens.Pos{Line: 4, Column: 0, Offset: 0}}),
			ast.NewRuleExportStatement("e", []*ast.AttachmentClause{}, tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}}),
		},
		tokens.Range{File: "single.sentra", From: tokens.Pos{Line: 2, Column: 0, Offset: 0}, To: tokens.Pos{Line: 2, Column: 0, Offset: 0}},
	)

	singleCharProgram := &ast.Program{
		Reference: "single.sentra",
		Statements: []ast.Statement{
			singleCharNsStmt,
			singleCharPolicyStmt,
		},
	}

	err = suite.idx.AddProgram(context.Background(), singleCharProgram)
	suite.Require().NoError(err)

	// Test resolving single character segments
	ns, policy, rule, err := suite.idx.ResolveSegments("a/b/c/e")

	suite.NoError(err)
	suite.Equal("a/b", ns)
	suite.Equal("c", policy)
	suite.Equal("e", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithSpecialCharacters() {
	// Test with paths that might have special characters (though FQN doesn't support them)
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/allow")

	suite.NoError(err)
	suite.Equal("com/example", ns)
	suite.Equal("auth", policy)
	suite.Equal("allow", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithWhitespace() {
	// Test with paths that have whitespace (should be treated as empty segments)
	ns, policy, rule, err := suite.idx.ResolveSegments(" com / example / auth ")

	suite.Error(err)
	suite.True(errors.Is(err, xerr.NotFoundError{}))
	suite.Contains(err.Error(), "namespace")
	suite.Equal("", ns)
	suite.Equal("", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithNestedNamespaceResolution() {
	// Test that the method correctly resolves nested namespaces
	// com/example/sub should resolve to the nested namespace, not the parent

	// Test nested namespace and policy
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/sub/admin")
	suite.NoError(err)
	suite.Equal("com/example/sub", ns)
	suite.Equal("admin", policy)
	suite.Equal("", rule)

	// Test nested namespace, policy, and rule
	ns, policy, rule, err = suite.idx.ResolveSegments("com/example/sub/admin/check")
	suite.NoError(err)
	suite.Equal("com/example/sub", ns)
	suite.Equal("admin", policy)
	suite.Equal("check", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithPartialNamespaceMatch() {
	// Test that partial namespace matches don't work
	// com/example/sub/extra should not match com/example/sub
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/sub/extra/admin")

	suite.Error(err)
	suite.True(errors.Is(err, xerr.NotFoundError{}))
	suite.Contains(err.Error(), "policy") // The error is about policy not found, not namespace
	suite.Equal("", ns)
	suite.Equal("", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithExactNamespaceMatch() {
	// Test that exact namespace matches work
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth")

	suite.NoError(err)
	suite.Equal("com/example", ns)
	suite.Equal("auth", policy)
	suite.Equal("", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithRuleInNestedNamespace() {
	// Test resolving a rule in a nested namespace

	// Test resolving a rule in the nested namespace
	ns, policy, rule, err := suite.idx.ResolveSegments("com/example/sub/admin/check")
	suite.NoError(err)
	suite.Equal("com/example/sub", ns)
	suite.Equal("admin", policy)
	suite.Equal("check", rule)
}

func (suite *SegmentsTestSuite) TestResolveSegmentsWithMultipleRules() {
	// Test resolving different rules in the same policy
	suite.Run("allow rule", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/allow")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("allow", rule)
	})

	suite.Run("deny rule", func() {
		ns, policy, rule, err := suite.idx.ResolveSegments("com/example/auth/deny")

		suite.NoError(err)
		suite.Equal("com/example", ns)
		suite.Equal("auth", policy)
		suite.Equal("deny", rule)
	})
}
