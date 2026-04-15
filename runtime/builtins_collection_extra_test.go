// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package runtime

import (
	"context"
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
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
		{"any non-list", func() (box.Value, error) {
			return BuiltinAny(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"collect non-callable", func() (box.Value, error) { return BuiltinCollect(s.ctx, site, list, box.Number(9)) }, "expected callable"},
		{"reduce bad callable arity", func() (box.Value, error) {
			return BuiltinReduce(s.ctx, site, list, box.Number(0), box.Callable(stubCallable{arity: 1}))
		}, "arity 2 or 3"},
		{"distinct bad selector arity", func() (box.Value, error) {
			return BuiltinDistinct(s.ctx, site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"all non-list", func() (box.Value, error) {
			return BuiltinAll(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"first non-list", func() (box.Value, error) {
			return BuiltinFirst(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"filter non-list", func() (box.Value, error) {
			return BuiltinFilter(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"collect non-list", func() (box.Value, error) {
			return BuiltinCollect(s.ctx, site, box.Number(3), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"distinct direct non-list", func() (box.Value, error) {
			return BuiltinDistinct(s.ctx, site, box.Number(1))
		}, "first argument must be a list"},
		{"distinct selector non-list", func() (box.Value, error) {
			return BuiltinDistinct(s.ctx, site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"any bad callable arity", func() (box.Value, error) {
			return BuiltinAny(s.ctx, site, list, box.Callable(stubCallable{arity: 0}))
		}, "arity 1 or 2"},
		{"first bad callable arity", func() (box.Value, error) {
			return BuiltinFirst(s.ctx, site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"filter bad callable arity", func() (box.Value, error) {
			return BuiltinFilter(s.ctx, site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"collect bad callable arity", func() (box.Value, error) {
			return BuiltinCollect(s.ctx, site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"reduce non-list", func() (box.Value, error) {
			return BuiltinReduce(s.ctx, site, box.Number(9), box.Number(0), box.Callable(stubCallable{arity: 2}))
		}, "first argument must be a list"},
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

func (s *RuntimeTestSuite) TestBuiltinsCollection_UndefinedEmptyAndPredicates() {
	site := s.builtinSite()
	undef := box.Undefined()
	isEven := box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			n, ok := args[0].NumberValue()
			if !ok {
				return box.Bool(false), nil
			}
			return box.Bool(int(n)%2 == 0), nil
		},
	})

	out, err := BuiltinAny(s.ctx, site, undef, isEven)
	s.Require().NoError(err)
	s.False(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinAll(s.ctx, site, undef, isEven)
	s.Require().NoError(err)
	s.False(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinFirst(s.ctx, site, undef, isEven)
	s.Require().NoError(err)
	s.True(out.IsUndefined())

	out, err = BuiltinFilter(s.ctx, site, undef, isEven)
	s.Require().NoError(err)
	lst, ok := out.ListValue()
	s.Require().True(ok)
	s.Empty(lst)

	never := box.Callable(stubCallable{
		arity: 2,
		fn: func([]box.Value) (box.Value, error) {
			s.Fail("reducer must not run for undefined list")
			return box.Undefined(), nil
		},
	})
	out, err = BuiltinReduce(s.ctx, site, undef, box.Number(10), never)
	s.Require().NoError(err)
	s.True(out.IsUndefined())

	nums := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})
	out, err = BuiltinAny(s.ctx, site, nums, isEven)
	s.Require().NoError(err)
	s.True(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinAny(s.ctx, site, box.List([]box.Value{box.Number(1), box.Number(3)}), isEven)
	s.Require().NoError(err)
	s.False(box.TrinaryFrom(out).IsTrue())

	withIdx := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			n, _ := args[0].NumberValue()
			idx, _ := args[1].NumberValue()
			return box.Bool(int(n)%2 == 0 && idx == 1), nil
		},
	})
	out, err = BuiltinAny(s.ctx, site, nums, withIdx)
	s.Require().NoError(err)
	s.True(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinAll(s.ctx, site, box.List([]box.Value{box.Number(2), box.Number(4)}), isEven)
	s.Require().NoError(err)
	s.True(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinAll(s.ctx, site, nums, isEven)
	s.Require().NoError(err)
	s.False(box.TrinaryFrom(out).IsTrue())

	out, err = BuiltinFirst(s.ctx, site, nums, isEven)
	s.Require().NoError(err)
	n, ok := out.NumberValue()
	s.Require().True(ok)
	s.Equal(2.0, n)

	out, err = BuiltinFirst(s.ctx, site, box.List([]box.Value{box.Number(1)}), isEven)
	s.Require().NoError(err)
	s.True(out.IsUndefined())

	out, err = BuiltinFilter(s.ctx, site, nums, isEven)
	s.Require().NoError(err)
	lst, ok = out.ListValue()
	s.Require().True(ok)
	s.Len(lst, 1)
	v, _ := lst[0].NumberValue()
	s.Equal(2.0, v)
}

func (s *RuntimeTestSuite) TestBuiltinsCollection_CollectReduceAndDistinct() {
	site := s.builtinSite()
	nums := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})

	double := box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			n, _ := args[0].NumberValue()
			return box.Number(n * 2), nil
		},
	})
	out, err := BuiltinCollect(s.ctx, site, box.List(nil), double)
	s.Require().NoError(err)
	empty, ok := out.ListValue()
	s.Require().True(ok)
	s.Empty(empty)

	out, err = BuiltinCollect(s.ctx, site, nums, double)
	s.Require().NoError(err)
	doubled, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(doubled, 3)
	x2, _ := doubled[2].NumberValue()
	s.Equal(6.0, x2)

	sum2 := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			a, _ := args[0].NumberValue()
			b, _ := args[1].NumberValue()
			return box.Number(a + b), nil
		},
	})
	out, err = BuiltinReduce(s.ctx, site, box.List(nil), box.Number(5), sum2)
	s.Require().NoError(err)
	v, _ := out.NumberValue()
	s.Equal(5.0, v)

	out, err = BuiltinReduce(s.ctx, site, nums, box.Number(0), sum2)
	s.Require().NoError(err)
	v, _ = out.NumberValue()
	s.Equal(6.0, v)

	sumIdx := box.Callable(stubCallable{
		arity: 3,
		fn: func(args []box.Value) (box.Value, error) {
			acc, _ := args[0].NumberValue()
			el, _ := args[1].NumberValue()
			idx, _ := args[2].NumberValue()
			return box.Number(acc + el + idx), nil
		},
	})
	out, err = BuiltinReduce(s.ctx, site, box.List([]box.Value{box.Number(10), box.Number(20)}), box.Number(0), sumIdx)
	s.Require().NoError(err)
	v, _ = out.NumberValue()
	s.Equal(31.0, v)

	out, err = BuiltinDistinct(s.ctx, site, box.List([]box.Value{box.Number(7)}))
	s.Require().NoError(err)
	one, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(one, 1)

	mix := box.List([]box.Value{
		box.Null(),
		box.Undefined(),
		box.Bool(true),
		box.Bool(false),
		box.Number(1),
		box.Number(1),
		box.String("x"),
		box.Trinary(trinary.Unknown),
	})
	out, err = BuiltinDistinct(s.ctx, site, mix)
	s.Require().NoError(err)
	uniq, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(uniq, 7)

	_, err = BuiltinDistinct(s.ctx, site, box.List([]box.Value{
		box.Dict(map[string]box.Value{"k": box.Number(1)}),
		box.Dict(map[string]box.Value{"k": box.Number(2)}),
	}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "unsupported key kind")

	// Selector path: len < 2 returns without invoking callable.
	out, err = BuiltinDistinct(s.ctx, site, box.List([]box.Value{box.Number(42)}), box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			s.Fail("selector must not run when list has fewer than 2 elements")
			return box.Undefined(), nil
		},
	}))
	s.Require().NoError(err)
	short, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(short, 1)

	dupKeys := box.List([]box.Value{box.String("a"), box.String("b")})
	out, err = BuiltinDistinct(s.ctx, site, dupKeys, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.String("same"), nil
		},
	}))
	s.Require().NoError(err)
	kded, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(kded, 1)
}

func (s *RuntimeTestSuite) TestBuiltinsCollection_PredicateInvokeErrors() {
	site := s.builtinSite()
	nums := box.List([]box.Value{box.Number(1), box.Number(2)})
	boom := box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("predicate failed")
		},
	})

	_, err := BuiltinAny(s.ctx, site, nums, boom)
	s.Require().ErrorContains(err, "predicate failed")

	_, err = BuiltinAll(s.ctx, site, nums, boom)
	s.Require().ErrorContains(err, "predicate failed")

	_, err = BuiltinFirst(s.ctx, site, nums, boom)
	s.Require().ErrorContains(err, "predicate failed")

	_, err = BuiltinFilter(s.ctx, site, nums, boom)
	s.Require().ErrorContains(err, "predicate failed")

	_, err = BuiltinCollect(s.ctx, site, nums, boom)
	s.Require().ErrorContains(err, "predicate failed")

	_, err = BuiltinReduce(s.ctx, site, nums, box.Number(0), box.Callable(stubCallable{
		arity: 2,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("reduce failed")
		},
	}))
	s.Require().ErrorContains(err, "reduce failed")

	_, err = BuiltinDistinct(s.ctx, site, nums, box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("distinct key fn failed")
		},
	}))
	s.Require().ErrorContains(err, "distinct key fn failed")

	_, err = BuiltinDistinct(s.ctx, site, nums, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.Dict(map[string]box.Value{"k": args[0]}), nil
		},
	}))
	s.Require().ErrorContains(err, "distinct key:")
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
