// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package builtins

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

type arityFailAfterPrecheckEnv struct {
	*testEnv
	calls int
}

func (e *arityFailAfterPrecheckEnv) CallableArity(fn box.Value) (int, error) {
	e.calls++
	if e.calls > 1 {
		return 0, fmt.Errorf("callable arity unavailable")
	}
	return e.testEnv.CallableArity(fn)
}

func TestBuiltinsCollection_ArityAndTypeErrors(t *testing.T) {
	t.Parallel()
	site := noopEnv()
	list := box.List([]box.Value{box.Number(1)})

	cases := []struct {
		name string
		fn   func() (box.Value, error)
		msg  string
	}{
		{"any wrong count", func() (box.Value, error) { return invoke(t, "any", site, list) }, "requires 2 arguments"},
		{"all wrong count", func() (box.Value, error) { return invoke(t, "all", site, list) }, "requires 2 arguments"},
		{"first wrong count", func() (box.Value, error) { return invoke(t, "first", site, list) }, "requires 2 arguments"},
		{"filter wrong count", func() (box.Value, error) { return invoke(t, "filter", site, list) }, "requires 2 arguments"},
		{"collect wrong count", func() (box.Value, error) { return invoke(t, "collect", site, list) }, "requires 2 arguments"},
		{"reduce wrong count", func() (box.Value, error) { return invoke(t, "reduce", site, list, box.Number(0)) }, "requires 3 arguments"},
		{"distinct wrong count", func() (box.Value, error) {
			return invoke(t, "distinct", site, list, box.Number(1), box.Number(2))
		}, "requires 1 or 2 arguments"},
		{"any non-list", func() (box.Value, error) {
			return invoke(t, "any", site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"collect non-callable", func() (box.Value, error) { return invoke(t, "collect", site, list, box.Number(9)) }, "expected callable"},
		{"reduce bad callable arity", func() (box.Value, error) {
			return invoke(t, "reduce", site, list, box.Number(0), box.Callable(stubCallable{arity: 1}))
		}, "arity 2 or 3"},
		{"distinct bad selector arity", func() (box.Value, error) {
			return invoke(t, "distinct", site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"all non-list", func() (box.Value, error) {
			return invoke(t, "all", site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"first non-list", func() (box.Value, error) {
			return invoke(t, "first", site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"filter non-list", func() (box.Value, error) {
			return invoke(t, "filter", site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"collect non-list", func() (box.Value, error) {
			return invoke(t, "collect", site, box.Number(3), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"distinct direct non-list", func() (box.Value, error) {
			return invoke(t, "distinct", site, box.Number(1))
		}, "first argument must be a list"},
		{"distinct selector non-list", func() (box.Value, error) {
			return invoke(t, "distinct", site, box.Number(1), box.Callable(stubCallable{arity: 1}))
		}, "first argument must be a list"},
		{"any bad callable arity", func() (box.Value, error) {
			return invoke(t, "any", site, list, box.Callable(stubCallable{arity: 0}))
		}, "arity 1 or 2"},
		{"first bad callable arity", func() (box.Value, error) {
			return invoke(t, "first", site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"filter bad callable arity", func() (box.Value, error) {
			return invoke(t, "filter", site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"collect bad callable arity", func() (box.Value, error) {
			return invoke(t, "collect", site, list, box.Callable(stubCallable{arity: 3}))
		}, "arity 1 or 2"},
		{"reduce non-list", func() (box.Value, error) {
			return invoke(t, "reduce", site, box.Number(9), box.Number(0), box.Callable(stubCallable{arity: 2}))
		}, "first argument must be a list"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.fn()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.msg)
		})
	}
}

func TestBuiltinsCollection_NullCollectionImplErrors(t *testing.T) {
	t.Parallel()
	site := noopEnv()
	fn := box.Callable(stubCallable{arity: 1})
	null := box.Null()

	cases := []struct {
		name string
		run  func() (box.Value, error)
		msg  string
	}{
		{"any", func() (box.Value, error) { return invoke(t, "any", site, null, fn) }, "first argument must be a list"},
		{"all", func() (box.Value, error) { return invoke(t, "all", site, null, fn) }, "first argument must be a list"},
		{"first", func() (box.Value, error) { return invoke(t, "first", site, null, fn) }, "first argument must be a list"},
		{"filter", func() (box.Value, error) { return invoke(t, "filter", site, null, fn) }, "first argument must be a list"},
		{"collect", func() (box.Value, error) { return invoke(t, "collect", site, null, fn) }, "first argument must be a list"},
		{"reduce", func() (box.Value, error) {
			return invoke(t, "reduce", site, null, box.Number(0), box.Callable(stubCallable{arity: 2}))
		}, "first argument must be a list"},
		{"distinct selector", func() (box.Value, error) {
			return invoke(t, "distinct", site, null, fn)
		}, "first argument must be a list"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := tc.run()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.msg)
		})
	}
}

func TestReduceArgsInvalidArity(t *testing.T) {
	t.Parallel()
	_, err := reduceArgs(1, box.Number(0), box.Number(1), 0)
	require.Error(t, err)
	require.ErrorContains(t, err, "arity 2 or 3")
}

func TestBuiltinsCollection_ImplCallableArityFailsAfterPrecheck(t *testing.T) {
	t.Parallel()
	list := box.List([]box.Value{box.Number(1)})
	fn := box.Callable(stubCallable{arity: 1, fn: func([]box.Value) (box.Value, error) {
		return box.Bool(true), nil
	}})

	hofs := []string{"any", "all", "first", "filter", "collect", "reduce", "distinct"}
	for _, name := range hofs {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			site := &arityFailAfterPrecheckEnv{testEnv: noopEnv()}
			var err error
			switch name {
			case "reduce":
				_, err = invoke(t, name, site, list, box.Number(0), box.Callable(stubCallable{arity: 2}))
			case "distinct":
				_, err = invoke(t, name, site, list, fn)
			default:
				_, err = invoke(t, name, site, list, fn)
			}
			require.Error(t, err)
			require.ErrorContains(t, err, "callable arity unavailable")
		})
	}
}

type flipArityEnv struct {
	*testEnv
	calls int
}

func (e *flipArityEnv) CallableArity(_ box.Value) (int, error) {
	e.calls++
	if e.calls == 1 {
		return 2, nil
	}
	return 1, nil
}

func TestBuiltinsCollection_ReduceReduceArgsErrorAfterPrecheck(t *testing.T) {
	t.Parallel()
	site := &flipArityEnv{testEnv: noopEnv()}
	list := box.List([]box.Value{box.Number(1), box.Number(2)})
	_, err := invoke(t, "reduce", site, list, box.Number(0), box.Callable(stubCallable{arity: 2}))
	require.Error(t, err)
	require.ErrorContains(t, err, "arity 2 or 3")
}

func TestBuiltinsCollection_DistinctBranches(t *testing.T) {
	t.Parallel()
	site := noopEnv()
	list := box.List([]box.Value{
		box.String("a"), box.String("a"), box.String("b"),
	})

	// selector with index branch (arity 2)
	selector := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			_, ok := args[1].NumberValue()
			require.True(t, ok)
			return args[0], nil
		},
	})
	out, err := invoke(t, "distinct", site, list, selector)
	require.NoError(t, err)
	vals, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, vals, 2)

	// unsupported key kind branch
	_, err = invoke(t, "distinct", site, list, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.List([]box.Value{args[0]}), nil
		},
	}))
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported key kind")
}

func TestBuiltinsCollection_UndefinedEmptyAndPredicates(t *testing.T) {
	t.Parallel()
	site := noopEnv()
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

	out, err := invoke(t, "any", site, undef, isEven)
	require.NoError(t, err)
	require.False(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "all", site, undef, isEven)
	require.NoError(t, err)
	require.False(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "first", site, undef, isEven)
	require.NoError(t, err)
	require.True(t, out.IsUndefined())

	out, err = invoke(t, "filter", site, undef, isEven)
	require.NoError(t, err)
	lst, ok := out.ListValue()
	require.True(t, ok)
	require.Empty(t, lst)

	never := box.Callable(stubCallable{
		arity: 2,
		fn: func([]box.Value) (box.Value, error) {
			panic("reducer must not run for undefined list")
		},
	})
	out, err = invoke(t, "reduce", site, undef, box.Number(10), never)
	require.NoError(t, err)
	require.True(t, out.IsUndefined())

	nums := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})
	out, err = invoke(t, "any", site, nums, isEven)
	require.NoError(t, err)
	require.True(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "any", site, box.List([]box.Value{box.Number(1), box.Number(3)}), isEven)
	require.NoError(t, err)
	require.False(t, box.TrinaryFrom(out).IsTrue())

	withIdx := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			n, _ := args[0].NumberValue()
			idx, _ := args[1].NumberValue()
			return box.Bool(int(n)%2 == 0 && idx == 1), nil
		},
	})
	out, err = invoke(t, "any", site, nums, withIdx)
	require.NoError(t, err)
	require.True(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "all", site, box.List([]box.Value{box.Number(2), box.Number(4)}), isEven)
	require.NoError(t, err)
	require.True(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "all", site, nums, isEven)
	require.NoError(t, err)
	require.False(t, box.TrinaryFrom(out).IsTrue())

	out, err = invoke(t, "first", site, nums, isEven)
	require.NoError(t, err)
	n, ok := out.NumberValue()
	require.True(t, ok)
	require.Equal(t, 2.0, n)

	out, err = invoke(t, "first", site, box.List([]box.Value{box.Number(1)}), isEven)
	require.NoError(t, err)
	require.True(t, out.IsUndefined())

	out, err = invoke(t, "filter", site, nums, isEven)
	require.NoError(t, err)
	lst, ok = out.ListValue()
	require.True(t, ok)
	require.Len(t, lst, 1)
	v, _ := lst[0].NumberValue()
	require.Equal(t, 2.0, v)
}

func TestBuiltinsCollection_CollectReduceAndDistinct(t *testing.T) {
	t.Parallel()
	site := noopEnv()
	nums := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})

	double := box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			n, _ := args[0].NumberValue()
			return box.Number(n * 2), nil
		},
	})
	out, err := invoke(t, "collect", site, box.List(nil), double)
	require.NoError(t, err)
	empty, ok := out.ListValue()
	require.True(t, ok)
	require.Empty(t, empty)

	out, err = invoke(t, "collect", site, nums, double)
	require.NoError(t, err)
	doubled, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, doubled, 3)
	x2, _ := doubled[2].NumberValue()
	require.Equal(t, 6.0, x2)

	sum2 := box.Callable(stubCallable{
		arity: 2,
		fn: func(args []box.Value) (box.Value, error) {
			a, _ := args[0].NumberValue()
			b, _ := args[1].NumberValue()
			return box.Number(a + b), nil
		},
	})
	out, err = invoke(t, "reduce", site, box.List(nil), box.Number(5), sum2)
	require.NoError(t, err)
	v, _ := out.NumberValue()
	require.Equal(t, 5.0, v)

	out, err = invoke(t, "reduce", site, nums, box.Number(0), sum2)
	require.NoError(t, err)
	v, _ = out.NumberValue()
	require.Equal(t, 6.0, v)

	sumIdx := box.Callable(stubCallable{
		arity: 3,
		fn: func(args []box.Value) (box.Value, error) {
			acc, _ := args[0].NumberValue()
			el, _ := args[1].NumberValue()
			idx, _ := args[2].NumberValue()
			return box.Number(acc + el + idx), nil
		},
	})
	out, err = invoke(t, "reduce", site, box.List([]box.Value{box.Number(10), box.Number(20)}), box.Number(0), sumIdx)
	require.NoError(t, err)
	v, _ = out.NumberValue()
	require.Equal(t, 31.0, v)

	out, err = invoke(t, "distinct", site, box.List([]box.Value{box.Number(7)}))
	require.NoError(t, err)
	one, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, one, 1)

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
	out, err = invoke(t, "distinct", site, mix)
	require.NoError(t, err)
	uniq, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, uniq, 7)

	_, err = invoke(t, "distinct", site, box.List([]box.Value{
		box.Dict(map[string]box.Value{"k": box.Number(1)}),
		box.Dict(map[string]box.Value{"k": box.Number(2)}),
	}))
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported key kind")

	// Selector path: len < 2 returns without invoking callable.
	out, err = invoke(t, "distinct", site, box.List([]box.Value{box.Number(42)}), box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			panic("selector must not run when list has fewer than 2 elements")
		},
	}))
	require.NoError(t, err)
	short, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, short, 1)

	dupKeys := box.List([]box.Value{box.String("a"), box.String("b")})
	out, err = invoke(t, "distinct", site, dupKeys, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.String("same"), nil
		},
	}))
	require.NoError(t, err)
	kded, ok := out.ListValue()
	require.True(t, ok)
	require.Len(t, kded, 1)
}

func TestBuiltinsCollection_PredicateInvokeErrors(t *testing.T) {
	t.Parallel()
	site := noopEnv()
	nums := box.List([]box.Value{box.Number(1), box.Number(2)})
	boom := box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("predicate failed")
		},
	})

	_, err := invoke(t, "any", site, nums, boom)
	require.ErrorContains(t, err, "predicate failed")

	_, err = invoke(t, "all", site, nums, boom)
	require.ErrorContains(t, err, "predicate failed")

	_, err = invoke(t, "first", site, nums, boom)
	require.ErrorContains(t, err, "predicate failed")

	_, err = invoke(t, "filter", site, nums, boom)
	require.ErrorContains(t, err, "predicate failed")

	_, err = invoke(t, "collect", site, nums, boom)
	require.ErrorContains(t, err, "predicate failed")

	_, err = invoke(t, "reduce", site, nums, box.Number(0), box.Callable(stubCallable{
		arity: 2,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("reduce failed")
		},
	}))
	require.ErrorContains(t, err, "reduce failed")

	_, err = invoke(t, "distinct", site, nums, box.Callable(stubCallable{
		arity: 1,
		fn: func([]box.Value) (box.Value, error) {
			return box.Undefined(), errors.New("distinct key fn failed")
		},
	}))
	require.ErrorContains(t, err, "distinct key fn failed")

	_, err = invoke(t, "distinct", site, nums, box.Callable(stubCallable{
		arity: 1,
		fn: func(args []box.Value) (box.Value, error) {
			return box.Dict(map[string]box.Value{"k": args[0]}), nil
		},
	}))
	require.ErrorContains(t, err, "distinct key:")
	require.ErrorContains(t, err, "unsupported key kind")
}
