// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sast "github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

// walkSource wraps a policy body in a namespace, adds the sibling policy `q`
// that import tests target, and returns the Validate result.
func (suite *IndexTestSuite) walkSource(body string) error {
	ctx := suite.T().Context()
	idx := CreateIndex()
	src := "namespace com/ex\n" +
		"policy q {\n  fact f?: number as f default 0\n  rule ok = f\n  export decision of ok\n}\n" +
		"policy pol {\n  " + body + "\n  export decision of a\n}\n"
	prog, err := parser.NewParserFromString(src, "walk.sentrie").ParseProgram(ctx)
	suite.Require().NoError(err)
	suite.Require().NoError(idx.AddProgram(ctx, prog))
	return idx.Validate(ctx)
}

// TestValidate_ReferenceCycleThroughEveryExpressionKind pins #269: a rule cycle
// and a let cycle that pass through an expression kind addNodes used to skip
// must be reported, and the same kinds without a cycle must still validate.
func (suite *IndexTestSuite) TestValidate_ReferenceCycleThroughEveryExpressionKind() {
	cases := []struct {
		name string
		// wrap renders the kind around the referenced name.
		wrap func(ref string) string
	}{
		{"index access left", func(ref string) string { return ref + "[0]" }},
		{"index access index", func(ref string) string { return "[1][" + ref + "]" }},
		{"cast", func(ref string) string { return ref + " as number" }},
		{"transform", func(ref string) string { return `transform ` + ref + ` with "."` }},
		{"trailing comment", func(ref string) string { return ref + " -- note" }},
		{"preceding comment", func(ref string) string { return "-- note\n    " + ref }},
		{"map key", func(ref string) string { return "{ [" + ref + "]: 1 }" }},
	}
	for _, tc := range cases {
		suite.Run("rule cycle through "+tc.name, func() {
			body := "rule a = { yield " + tc.wrap("b") + "\n  }\n  rule b = { yield " + tc.wrap("a") + "\n  }"
			err := suite.walkSource(body)
			suite.Require().Error(err)
			suite.Contains(err.Error(), "infinite recursion")
		})
		suite.Run("let cycle through "+tc.name, func() {
			body := "let x = " + tc.wrap("y") + "\n  let y = " + tc.wrap("x") + "\n  rule a = { yield x }"
			err := suite.walkSource(body)
			suite.Require().Error(err)
			suite.Contains(err.Error(), "infinite recursion")
		})
		suite.Run("no cycle through "+tc.name, func() {
			body := "rule a = { yield " + tc.wrap("b") + "\n  }\n  rule b = { yield [1] }"
			suite.NoError(suite.walkSource(body))
		})
	}
}

// TestValidate_ReferenceCycleThroughImportWith pins that an `import decision`
// `with` value is walked: it is evaluated in the importing policy, so it can
// reference that policy's rules and lets.
func (suite *IndexTestSuite) TestValidate_ReferenceCycleThroughImportWith() {
	suite.Run("rule cycle", func() {
		err := suite.walkSource("rule a = import decision ok from com/ex/q with f as b\n  rule b = { yield a }")
		suite.Require().Error(err)
		suite.Contains(err.Error(), "infinite recursion")
	})
	suite.Run("with value through a let", func() {
		err := suite.walkSource("let x = a\n  rule a = import decision ok from com/ex/q with f as x")
		suite.Require().Error(err)
		suite.Contains(err.Error(), "infinite recursion")
	})
	suite.Run("no cycle", func() {
		suite.NoError(suite.walkSource("rule a = import decision ok from com/ex/q with f as b\n  rule b = { yield 1 }"))
	})
}

// TestValidate_ReferenceCycleAcrossMixedKinds pins cycles that span more than
// two nodes, change expression kind at every hop, or bury the reference several
// kinds deep.
func (suite *IndexTestSuite) TestValidate_ReferenceCycleAcrossMixedKinds() {
	cases := []struct {
		name  string
		body  string
		names []string
	}{
		{
			name:  "three rules, index access then map key then transform",
			body:  "rule a = { yield b[0]\n  }\n  rule b = { yield { [c]: 1 }\n  }\n  rule c = { yield transform a with \".\"\n  }",
			names: []string{"a", "b", "c"},
		},
		{
			name:  "rule and lets, cast then index then comment",
			body:  "rule a = { yield x as number\n  }\n  let x = [1][y]\n  let y = a -- note",
			names: []string{"a", "x", "y"},
		},
		{
			name:  "cast inside index access inside a map key",
			body:  "rule a = { yield { [[1][b as number]]: 1 }\n  }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "default clause",
			body:  "rule a = default b[0] { yield 1 }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "left of is defined",
			body:  "rule a = { yield b is defined }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "left of is empty",
			body:  "rule a = { yield b is empty }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "left of an is type check",
			body:  "rule a = { yield b is number }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "value on the right of is",
			body:  "rule a = { yield 1 is b }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "lambda body",
			body:  "rule a = { yield any([1], (n) => { yield b }) }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
		{
			name:  "lambda body behind a field access and a ternary",
			body:  "rule a = { yield true ? any([1], (n) => { yield b }) : false }\n  rule b = { yield [a].first }",
			names: []string{"a", "b"},
		},
		{
			name:  "when clause",
			body:  "rule a = when b[0] { yield 1 }\n  rule b = { yield a }",
			names: []string{"a", "b"},
		},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			err := suite.walkSource(tc.body)
			suite.Require().Error(err)
			suite.Contains(err.Error(), "infinite recursion")
			for _, n := range tc.names {
				suite.Contains(err.Error(), n)
			}
		})
	}
}

// TestValidate_AcyclicPackUsingEveryKindStillValidates guards against false
// positives: a reference walk that reads a shape operand, a field name, a cast
// target or a lambda body as a reference to a rule would report a cycle here.
func (suite *IndexTestSuite) TestValidate_AcyclicPackUsingEveryKindStillValidates() {
	src := `namespace com/ex
shape Shp { id: string }
policy q {
  fact f?: number as f default 0
  rule ok = f
  export decision of ok
}
policy pol {
  let rec = { "b": 1, "a": 2 }
  let lst = [1, 2]
  -- the fields b and a are names, not references to the rules b and a
  rule a = { yield rec.b + rec.a }
  rule b = { yield a }
  rule shape_operand = { yield { "id": "x" } is Shp }
  rule list_shape_operand = { yield [{ "id": "x" }] is list[Shp] }
  rule value_operand = { yield 1 is b }
  rule index_kinds = { yield lst[0] as number }
  rule map_kinds = { yield { [a]: transform "x" with "." } }
  rule comments = { yield b -- note
  }
  rule lambda_kinds = { yield any(lst, (n) => { yield n > b }) }
  rule imported = import decision ok from com/ex/q with f as b
  rule defined_empty = { yield (b is defined) and (lst is empty) }
  rule ternary_kinds = { yield b ? a : lst[1] }
  rule elvis = { yield b ?: a }
  export decision of b
}
`
	ctx := suite.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, "acyclic.sentrie").ParseProgram(ctx)
	suite.Require().NoError(err)
	suite.Require().NoError(idx.AddProgram(ctx, prog))
	suite.NoError(idx.Validate(ctx))
}

// TestForEachExprChildCoversEveryASTNode fails when the ast package gains an
// expression kind (or a block statement kind) that forEachExprChild has no case
// for, so the reference cycle check cannot silently skip a new kind.
func (suite *IndexTestSuite) TestForEachExprChildCoversEveryASTNode() {
	r := tokens.Range{}
	id := func(n string) *sast.Identifier { return sast.NewIdentifier(n, r) }
	samples := map[string]sast.Node{
		"Identifier":                 id("x"),
		"IntegerLiteral":             sast.NewIntegerLiteral(1, r),
		"FloatLiteral":               sast.NewFloatLiteral(1.5, r),
		"StringLiteral":              sast.NewStringLiteral("s", r),
		"NullLiteral":                sast.NewNullLiteral(r),
		"TrinaryLiteral":             sast.NewTrinaryLiteral(trinary.True, r),
		"PipelineHoleExpression":     sast.NewPipelineHoleExpression(r),
		"CallExpression":             sast.NewCallExpression(id("f"), []sast.Expression{id("a")}, false, nil, r),
		"InfixExpression":            sast.NewInfixExpression(id("l"), id("r"), "+", r),
		"UnaryExpression":            sast.NewUnaryExpression("!", id("u"), r),
		"TernaryExpression":          sast.NewTernaryExpression(id("c"), id("t"), id("e"), r),
		"BlockExpression":            sast.NewBlockExpression([]sast.Statement{sast.NewCommentStatement("c", r)}, id("y"), r),
		"LambdaExpression":           sast.NewLambdaExpression(nil, sast.NewBlockExpression(nil, id("b"), r), r),
		"FieldAccessExpression":      sast.NewFieldAccessExpression(id("l"), "f", r),
		"IndexAccessExpression":      sast.NewIndexAccessExpression(id("l"), id("i"), r),
		"ListLiteral":                sast.NewListLiteral([]sast.Expression{id("v")}, r),
		"MapLiteral":                 sast.NewMapLiteral([]sast.MapEntry{{Key: id("k"), Value: id("v")}}, r),
		"CastExpression":             sast.NewCastExpression(id("e"), sast.NewStringTypeRef(r), r),
		"TransformExpression":        sast.NewTransformExpression(id("a"), ".", r),
		"IsDefinedExpression":        sast.NewIsDefinedExpression(id("l"), r),
		"IsEmptyExpression":          sast.NewIsEmptyExpression(id("l"), r),
		"IsExpression":               sast.NewIsExpression(id("l"), id("r"), nil, r),
		"TrailingCommentExpression":  sast.NewTrailingCommentExpression("c", id("w"), r),
		"PrecedingCommentExpression": sast.NewPrecedingCommentExpression("c", id("w"), r),
		"ImportClause":               sast.NewImportClause("ok", nil, []*sast.WithClause{sast.NewWithClause("f", id("v"), r)}, r),
		"WithClause":                 sast.NewWithClause("f", id("v"), r),
		"AttachmentClause":           sast.NewAttachmentClause("what", id("a"), r),
		"VarDeclaration":             sast.NewVarDeclaration("v", nil, id("val"), r),
		"RuleStatement":              sast.NewRuleStatement("r", nil, nil, id("b"), r),
		"CommentStatement":           sast.NewCommentStatement("c", r),
	}

	// Every ast type that can sit in an expression position, or in a block as a
	// statement, must have a sample here.
	kinds := astNodeKinds(suite)
	suite.Greater(len(kinds), 20, "the ast package scan found too few node kinds")
	var missing []string
	for _, name := range kinds {
		if _, ok := samples[name]; !ok {
			missing = append(missing, name)
		}
	}
	suite.Empty(missing, "ast kinds with no forEachExprChild coverage: add a case to forEachExprChild and a sample here")

	for name, node := range samples {
		suite.Run(name, func() {
			suite.NoError(forEachExprChild(node, func(sast.Node) error { return nil }))
		})
	}

	suite.Run("unknown node", func() {
		suite.Error(forEachExprChild(sast.NewDeriveStatement("d", nil, r), func(sast.Node) error { return nil }))
	})
}

// astNodeKinds lists the ast types that implement expressionNode, plus the
// statement types a block can hold, by reading the ast package source.
func astNodeKinds(suite *IndexTestSuite) []string {
	fset := token.NewFileSet()
	entries, err := os.ReadDir("../ast")
	suite.Require().NoError(err)
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := goparser.ParseFile(fset, filepath.Join("../ast", e.Name()), nil, 0)
		suite.Require().NoError(err)
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != "expressionNode" {
				continue
			}
			t := fn.Recv.List[0].Type
			if star, ok := t.(*ast.StarExpr); ok {
				t = star.X
			}
			if id, ok := t.(*ast.Ident); ok {
				seen[id.Name] = true
			}
		}
	}
	for _, s := range []string{"VarDeclaration", "RuleStatement", "CommentStatement"} {
		seen[s] = true
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
