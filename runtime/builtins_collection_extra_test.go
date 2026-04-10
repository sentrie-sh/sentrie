// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
)

type stubCallable struct {
	arity int
	fn    func(args []box.Value) (box.Value, error)
}

func (s stubCallable) Arity() int { return s.arity }
func (s stubCallable) Invoke(_ context.Context, _ *CallSite, args []box.Value) (box.Value, error) {
	if s.fn == nil {
		return box.Undefined(), nil
	}
	return s.fn(args)
}

func (s *RuntimeTestSuite) TestBuiltinsCollection_ArityAndTypeErrors() {
	site := s.builtinSite()
	list := box.List([]box.Value{box.Number(1)})

	cases := []struct {
		name string
		fn   func() (box.Value, error)
		msg  string
	}{
		{"any wrong count", func() (box.Value, error) { return BuiltinAny(s.ctx, site, list) }, "requires 2 arguments"},
		{"all wrong count", func() (box.Value, error) { return BuiltinAll(s.ctx, site, list) }, "requires 2 arguments"},
		{"first wrong count", func() (box.Value, error) { return BuiltinFirst(s.ctx, site, list) }, "requires 2 arguments"},
		{"filter wrong count", func() (box.Value, error) { return BuiltinFilter(s.ctx, site, list) }, "requires 2 arguments"},
		{"collect wrong count", func() (box.Value, error) { return BuiltinCollect(s.ctx, site, list) }, "requires 2 arguments"},
		{"reduce wrong count", func() (box.Value, error) { return BuiltinReduce(s.ctx, site, list, box.Number(0)) }, "requires 3 arguments"},
		{"distinct wrong count", func() (box.Value, error) { return BuiltinDistinct(s.ctx, site, list, box.Number(1), box.Number(2)) }, "requires 1 or 2 arguments"},
		{"any non-list", func() (box.Value, error) { return BuiltinAny(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1})) }, "first argument must be a list"},
		{"collect non-callable", func() (box.Value, error) { return BuiltinCollect(s.ctx, site, list, box.Number(9)) }, "expected callable"},
		{"reduce bad callable arity", func() (box.Value, error) { return BuiltinReduce(s.ctx, site, list, box.Number(0), box.Callable(stubCallable{arity: 1})) }, "arity 2 or 3"},
		{"distinct bad selector arity", func() (box.Value, error) { return BuiltinDistinct(s.ctx, site, list, box.Callable(stubCallable{arity: 3})) }, "arity 1 or 2"},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := tc.fn()
			s.Require().Error(err)
			s.Require().ErrorContains(err, tc.msg)
		})
	}
}

func (s *RuntimeTestSuite) TestBuiltinsCollection_DistinctBranches() {
	site := s.builtinSite()
	list := box.List([]box.Value{
		box.String("a"), box.String("a"), box.String("b"),
	})

	// selector with index branch (arity 2)
	selector := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			_, ok := args[1].NumberValue()
			s.Require().True(ok)
			return args[0], nil
		},
	})
	out, err := BuiltinDistinct(s.ctx, site, list, selector)
	s.Require().NoError(err)
	vals, ok := out.ListValue()
	s.Require().True(ok)
	s.Require().Len(vals, 2)

	// unsupported key kind branch
	_, err = BuiltinDistinct(s.ctx, site, list, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.List([]box.Value{args[0]}), nil
		},
	}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "unsupported key kind")
}

func (s *RuntimeTestSuite) TestCallableHelpers_ErrorBranches() {
	site := s.builtinSite()
	ctx := context.Background()

	// callable payload not implementing Callable
	_, err := callableFromValue(box.Callable(struct{}{}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "internal error")

	// direct invocation now takes an already-unwrapped callable.
	_, err = invokeCallable(ctx, site, stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			s.Require().Len(args, 1)
			return box.Number(1), nil
		},
	}, []box.Value{box.Number(1)})
	s.Require().NoError(err)

	_, err = iterArgs(site, stubCallable{arity: 9}, box.Number(1), 0)
	s.Require().Error(err)
	_, err = reduceArgs(site, stubCallable{arity: 9}, box.Number(0), box.Number(1), 0)
	s.Require().Error(err)
}

func (s *RuntimeTestSuite) TestEvalCallBoundaryBranches() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	ec.BindModule("mod", &ModuleBinding{Alias: "mod"})

	// module callable arg should be rejected before boundary conversion
	_, err := getTarget(ctx, ec, &executorImpl{}, p, ast.NewCallExpression(
		ast.NewIdentifier("mod.fn", stubRange()),
		nil, false, nil, stubRange(),
	))
	s.Require().NoError(err)

	target, _ := getTarget(ctx, ec, &executorImpl{}, p, ast.NewCallExpression(
		ast.NewIdentifier("mod.fn", stubRange()),
		nil, false, nil, stubRange(),
	))
	_, err = target(ctx, box.Callable(stubCallable{arity: 0}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "cannot pass callable value")

	// memoized-call hash path returns empty for callable boundary failure.
	hash := calculateHashKey(&ast.CallExpression{}, []box.Value{box.Callable(stubCallable{arity: 0})})
	s.Equal("", hash)
}
