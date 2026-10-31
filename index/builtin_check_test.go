// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/require"
)

func (s *IndexTestSuite) parseAndValidate(src, file string) error {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, file).ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	return idx.Validate(ctx)
}

func (s *IndexTestSuite) TestBuiltinCheckFilterStringLiteralReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact items: list[number] as items
  rule allow = { yield filter("abc", (x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}

func (s *IndexTestSuite) TestBuiltinCheckFilterListLiteralAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = { yield filter([1, 2], (x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckLetStringAnnotationReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = {
    let s: string = "abc"
    yield filter(s, (x) => { yield x })
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}

func (s *IndexTestSuite) TestBuiltinCheckMergeAfterFilterReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact ys: list[number] as ys
  fact d: document as d
  rule allow = {
    let xs = filter(ys, (x) => { yield x })
    yield merge(xs, d)
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "first argument is not a dict")
}

func (s *IndexTestSuite) TestBuiltinCheckPipelineFilterSameAsDirect() {
	direct := `namespace com/ex
policy p {
  rule allow = { yield filter("abc", (x) => { yield x }) }
  export decision of allow
}
`
	pipeline := `namespace com/ex
policy p {
  rule allow = { yield "abc" |> filter((x) => { yield x }) }
  export decision of allow
}
`
	errDirect := s.parseAndValidate(direct, "direct.sentrie")
	errPipeline := s.parseAndValidate(pipeline, "pipe.sentrie")
	s.Require().Error(errDirect)
	s.Require().Error(errPipeline)
	s.Contains(errDirect.Error(), "filter: first argument must be a list")
	s.Contains(errPipeline.Error(), "filter: first argument must be a list")
}

func (s *IndexTestSuite) TestBuiltinCheckPipelineLoweringIsCallExpression() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy p {
  fact ys: list[number] as ys
  rule allow = { yield ys |> filter((x) => { yield x }) }
  export decision of allow
}
`
	prog, err := parser.NewParserFromString(src, "pipe.sentrie").ParseProgram(ctx)
	s.Require().NoError(err)
	pol := prog.Statements[1].(*ast.PolicyStatement)
	rule := pol.Statements[1].(*ast.RuleStatement)
	block := rule.Body.(*ast.BlockExpression)
	call, ok := block.Yield.(*ast.CallExpression)
	s.Require().True(ok, "pipeline must lower to CallExpression")
	_, ok = call.Callee.(*ast.Identifier)
	s.Require().True(ok)
	s.Require().Len(call.Arguments, 2)
}

func (s *IndexTestSuite) TestBuiltinCheckReduceCallableArityReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield reduce(xs, 0, (a) => { yield a }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "reduce: reducer must have arity 2 or 3")
}

func (s *IndexTestSuite) TestBuiltinCheckReduceCallableArityAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield reduce(xs, 0, (a, b, i) => { yield a }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckDistinctOptionalCallableReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield distinct(xs, (a, b, c) => { yield a }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "distinct: selector must have arity 1 or 2")
}

func (s *IndexTestSuite) TestBuiltinCheckDistinctOptionalCallableAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield distinct(xs) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)

	err = s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield distinct(xs, (x) => { yield x }) }
  export decision of allow
}
`, "f2.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckCountMismatchUndefinedAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = { yield count(5) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckBuiltinShadowedByLetAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = {
    let filter = (x) => { yield x }
    yield filter("abc")
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckCallArityTooMany() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield flatten(xs, 1, 2) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "flatten requires 1 or 2 arguments")
}

func (s *IndexTestSuite) TestBuiltinCheckCallArityTooFew() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact d: document as d
  rule allow = { yield merge(d) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "merge requires 2 arguments")
}

func (s *IndexTestSuite) TestBuiltinCheckCallArityNowTooMany() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = { yield now(1) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "now requires 0 arguments")
}

func (s *IndexTestSuite) TestBuiltinCheckCastListPipelineAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact doc: document as doc
  rule allow = {
    yield (doc.items as list[number]) |> filter((x) => { yield x })
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckUntypedDocumentFieldSilentAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact doc: document as doc
  rule allow = { yield doc.x |> filter((x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckPropagationChainReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact items: list[number] as items
  rule allow = {
    yield items |> filter((x) => { yield x }) |> count() |> filter((x) => { yield x })
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}

func (s *IndexTestSuite) TestBuiltinCheckOptionalLambdaArityAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield filter(xs, (a: number, b?: number) => { yield a }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckExcessRequiredLambdaArityReject() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact xs: list[number] as xs
  rule allow = { yield filter(xs, (a, b, c) => { yield a }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: callable must have arity 1 or 2")
}

func (s *IndexTestSuite) TestBuiltinCheckShapeListFieldAccept() {
	err := s.parseAndValidate(`namespace com/ex
shape Row {
  items: list[number]
}
policy p {
  fact row: Row as row
  rule allow = { yield filter(row.items, (x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckShapeStringFieldReject() {
	err := s.parseAndValidate(`namespace com/ex
shape Row {
  label: string
}
policy p {
  fact row: Row as row
  rule allow = { yield filter(row.label, (x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}

func (s *IndexTestSuite) TestBuiltinCheckShapeAliasStringNotDict() {
	err := s.parseAndValidate(`namespace com/ex
shape LabelAlias string
policy p {
  fact label: LabelAlias as label
  rule allow = { yield count(label) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckListFieldHopSilentAccept() {
	err := s.parseAndValidate(`namespace com/ex
shape Row {
  items: list[number]
}
policy p {
  fact row: Row as row
  rule allow = { yield row.items.name |> filter((x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckErrorsJoinMultiple() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule allow = {
    yield filter("a", (x) => { yield x })
  }
  export decision of allow
}
policy q {
  rule allow = { yield now(1) }
  export decision of allow
}
`, "multi.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}

func TestRequiredLambdaArityMatchesAST(t *testing.T) {
	t.Parallel()
	lam := ast.NewLambdaExpressionFull(
		[]string{"a", "b", "c"},
		nil,
		[]bool{false, true, false},
		nil,
		ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, tokens.Range{}), tokens.Range{}),
		tokens.Range{},
	)
	require.Equal(t, 2, ast.RequiredLambdaArity(lam))
	require.Equal(t, 0, ast.RequiredLambdaArity(nil))
}

func TestLookupShapeKindAliasVsComplex(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := tokens.Range{File: "t.sentrie"}
	nsFQN := ast.NewFQN([]string{"com", "ex"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	aliasShape, err := createShape(ns, nil, ast.NewShapeStatement("LabelAlias", ast.NewStringTypeRef(rng), nil, rng))
	require.NoError(t, err)
	ns.Shapes["LabelAlias"] = aliasShape

	kind, ok := lookupShapeKind(idx, nil, ast.NewShapeTypeRef(ast.NewFQN([]string{"com", "ex", "LabelAlias"}, rng).Ptr(), rng))
	require.True(t, ok)
	require.Equal(t, box.ValueString, kind)
}

func TestValidateAllRepoPacksUnchanged(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..")
	dirs := []string{
		filepath.Join(root, "example_pack"),
	}
	ctx := t.Context()

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		for _, ent := range entries {
			if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".sentrie") {
				continue
			}
			path := filepath.Join(dir, ent.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			src := string(data)
			if !strings.Contains(src, "export decision") && !strings.Contains(src, "export ") {
				continue
			}
			idx := CreateIndex()
			prog, err := parser.NewParserFromString(src, ent.Name()).ParseProgram(ctx)
			if err != nil {
				continue
			}
			if err := idx.AddProgram(ctx, prog); err != nil {
				continue
			}
			err = idx.Validate(ctx)
			require.NoError(t, err, "unexpected validate failure for %s: %v", path, err)
		}
	}
}

func TestCheckBuiltinCallsLetCycleDoesNotStackOverflow(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	src := `namespace com/ex
policy p {
  rule allow = {
    let a = b
    let b = a
    yield filter(a, (x) => { yield x })
  }
  export decision of allow
}
`
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, "cycle.sentrie").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))

	// Call checkBuiltinCalls directly - no detectReferenceCycle - must not stack-overflow.
	err = idx.checkBuiltinCalls(ctx)
	require.NoError(t, err)
}

func TestErrBuiltinKindsWrapIndex(t *testing.T) {
	t.Parallel()
	r := tokens.Range{File: "f.sentrie", From: tokens.Pos{Line: 0, Column: 0}, To: tokens.Pos{Line: 0, Column: 1}}
	require.True(t, errors.Is(xerr.ErrBuiltinArgKind(r, "msg"), xerr.ErrIndex))
	require.True(t, errors.Is(xerr.ErrBuiltinCallableArity(r, "msg"), xerr.ErrIndex))
	require.True(t, errors.Is(xerr.ErrBuiltinCallArity(r, "msg"), xerr.ErrIndex))
}

// TestBuiltinCheckEveryPolicyExpressionSite checks that a bad builtin call
// fails validation at its span wherever a policy holds an expression, not only
// in rules and derives (#218): policy lets, fact defaults, export attachments
// and import `with` clauses.
func (s *IndexTestSuite) TestBuiltinCheckEveryPolicyExpressionSite() {
	const target = `namespace com/ex
policy q {
  fact f: number as f
  rule allow = default true { yield f > 0 }
  export decision of allow
}
`
	testCases := []struct {
		name string
		src  string
		line int // 1-based line of the offending argument
	}{
		{"policy let", `namespace com/ex
policy p {
  let n = count([1], [2])
  rule allow = default true { yield n > 0 }
  export decision of allow
}
`, 3},
		{"fact default", `namespace com/ex
policy p {
  fact n?: number as n default count([1], [2])
  rule allow = default true { yield n > 0 }
  export decision of allow
}
`, 3},
		{"export attachment", `namespace com/ex
policy p {
  rule allow = default true { yield true }
  export decision of allow attach n as count([1], [2])
}
`, 4},
		{"import with clause", target + `policy p {
  rule allow = import decision allow from com/ex/q with f as count([1], [2])
  export decision of allow
}
`, 8},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.parseAndValidate(tc.src, "f.sentrie")
			s.Require().Error(err)
			var arity xerr.BuiltinCallArityError
			s.Require().ErrorAs(err, &arity)
			s.Equal("count requires 1 argument", arity.Message)
			s.Equal("f.sentrie", arity.At.File)
			s.Equal(tc.line, arity.At.From.Line)
			s.Equal("[2]", tc.src[arity.At.From.Offset:arity.At.To.Offset], "span must cover the first excess argument")
		})
	}
}

// TestBuiltinCheckKindAtPolicyLet checks the argument-kind check, not only
// arity, at a site outside rules and derives.
func (s *IndexTestSuite) TestBuiltinCheckKindAtPolicyLet() {
	src := `namespace com/ex
policy p {
  let xs = filter("abc", (x) => { yield x })
  rule allow = default true { yield count(xs) > 0 }
  export decision of allow
}
`
	err := s.parseAndValidate(src, "f.sentrie")
	s.Require().Error(err)
	var kind xerr.BuiltinArgKindError
	s.Require().ErrorAs(err, &kind)
	s.Contains(kind.Message, "filter: first argument must be a list")
	s.Equal(3, kind.At.From.Line)
	s.Equal(`"abc"`, src[kind.At.From.Offset:kind.At.To.Offset])
}

// TestBuiltinCheckPolicySitesRespectShadowing checks that the wider walk keeps
// the runtime's precedence: a policy let that shadows a builtin is not checked
// as that builtin anywhere in the policy.
func (s *IndexTestSuite) TestBuiltinCheckPolicySitesRespectShadowing() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  let count = (a: number, b: number) => { yield a + b }
  let n = count(1, 2)
  rule allow = default true { yield n > 0 }
  export decision of allow attach m as count(3, 4)
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestBuiltinCheckStdlibArityAndKind() {
	cases := []struct {
		name string
		expr string
		msg  string
	}{
		{"glob_match too few", `glob_match("a", ["/"])`, "glob_match requires 3 arguments"},
		{"glob_match delimiters string", `glob_match("a", "/", "b")`, "glob_match: delimiters must be a list of single-character strings"},
		{"glob_match pattern number", `glob_match(1, ["/"], "b")`, "glob_match: pattern must be a string"},
		{"object_keys too many", `object_keys({}, {})`, "object_keys requires 1 argument"},
		{"object_keys string", `object_keys("a")`, "object_keys: argument must be a dict"},
		{"object_get too few", `object_get({}, "a")`, "object_get requires 3 arguments"},
		{"object_remove keys string", `object_remove({}, "a")`, "object_remove: keys must be a list of strings"},
		{"object_filter dict first", `object_filter(1, ["a"])`, "object_filter: first argument must be a dict"},
		{"object_union too few", `object_union({})`, "object_union requires 2 arguments"},
		{"parse_quantity number", `parse_quantity(5)`, "parse_quantity: argument must be a string"},
		{"parse_duration too many", `parse_duration("1h", "2h")`, "parse_duration requires 1 argument"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := s.parseAndValidate("namespace com/ex\npolicy p {\n  rule allow = { yield "+tc.expr+" }\n  export decision of allow\n}\n", "f.sentrie")
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

func (s *IndexTestSuite) TestBuiltinCheckStdlibWellFormedAccept() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact d: document as d
  rule allow = {
    let g = glob_match("/a/**", ["/"], "/a/b")
    let k = object_keys({ "a": 1 })
    let v = object_get(d, ["a", "b"], 0)
    let r = object_remove({ "a": 1 }, ["a"])
    let f = object_filter({ "a": 1 }, ["a"])
    let u = object_union({ "a": 1 }, { "b": 2 })
    yield g and parse_quantity("10Gi") > 0 and parse_duration("2h") > 0
  }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

// A document is always a dict at runtime, so the static kind check must accept
// a document fact wherever a dict is required (this also covers merge; the count case is a regression guard, since count is never kind-checked).
func (s *IndexTestSuite) TestBuiltinCheckDocumentFactAcceptedAsDict() {
	for _, expr := range []string{
		`object_keys(d)`,
		`object_union(d, {})`,
		`object_remove(d, ["a"])`,
		`object_filter(d, ["a"])`,
		`merge(d, {})`,
		`count(d)`,
	} {
		s.Run(expr, func() {
			err := s.parseAndValidate("namespace com/ex\npolicy p {\n  fact d: document as d\n  rule allow = { yield "+expr+" }\n  export decision of allow\n}\n", "f.sentrie")
			s.Require().NoError(err)
		})
	}
}

func (s *IndexTestSuite) TestBuiltinCheckDocumentFactStillRejectedAsList() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact d: document as d
  rule allow = { yield filter(d, (x) => { yield x }) }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "filter: first argument must be a list")
}
