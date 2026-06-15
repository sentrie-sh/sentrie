// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
)

var (
	declAll = &Decl{
		Name:        "all",
		Description: "Reports whether every list element satisfies the predicate callable.",
		DeriveSafe:  true,
		Sig:         hofSig("all", "all: callable must have arity 1 or 2"),
		Impl:        implAll,
	}

	declAny = &Decl{
		Name:        "any",
		Description: "Reports whether any list element satisfies the predicate callable.",
		DeriveSafe:  true,
		Sig:         hofSig("any", "any: callable must have arity 1 or 2"),
		Impl:        implAny,
	}

	declFirst = &Decl{
		Name:        "first",
		Description: "Returns the first list element satisfying the predicate callable.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "first: first argument must be a list"},
				{
					Name:               "predicate",
					Kinds:              kindCallable,
					CallableArities:    []int{1, 2},
					CallableArityError: "first: callable must have arity 1 or 2",
				},
			},
			TooFewError:  "first requires 2 arguments",
			TooManyError: "first requires 2 arguments",
			Result:       ParamSig{Name: "result"},
		},
		Impl: implFirst,
	}

	declFilter = &Decl{
		Name:        "filter",
		Description: "Returns list elements for which the predicate callable is true.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "filter: first argument must be a list"},
				{
					Name:               "predicate",
					Kinds:              kindCallable,
					CallableArities:    []int{1, 2},
					CallableArityError: "filter: callable must have arity 1 or 2",
				},
			},
			TooFewError:  "filter requires 2 arguments",
			TooManyError: "filter requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implFilter,
	}

	declCollect = &Decl{
		Name:        "collect",
		Description: "Maps each list element through the transform callable.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "collect: first argument must be a list"},
				{
					Name:               "transform",
					Kinds:              kindCallable,
					CallableArities:    []int{1, 2},
					CallableArityError: "collect: callable must have arity 1 or 2",
				},
			},
			TooFewError:  "collect requires 2 arguments",
			TooManyError: "collect requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implCollect,
	}

	declReduce = &Decl{
		Name:        "reduce",
		Description: "Folds a list with an initial accumulator using the reducer callable.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "reduce: first argument must be a list"},
				{Name: "initial"},
				{
					Name:               "reducer",
					Kinds:              kindCallable,
					CallableArities:    []int{2, 3},
					CallableArityError: "reduce: reducer must have arity 2 or 3",
				},
			},
			TooFewError:  "reduce requires 3 arguments",
			TooManyError: "reduce requires 3 arguments",
			Result:       ParamSig{Name: "result"},
		},
		Impl: implReduce,
	}

	declDistinct = &Decl{
		Name:        "distinct",
		Description: "Removes duplicate list elements, optionally by a key selector callable.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "distinct: first argument must be a list"},
				{
					Name:               "keyFn",
					Kinds:              kindCallable,
					Optional:           true,
					CallableArities:    []int{1, 2},
					CallableArityError: "distinct: selector must have arity 1 or 2",
				},
			},
			TooFewError:  "distinct requires 1 or 2 arguments",
			TooManyError: "distinct requires 1 or 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implDistinct,
	}
)

func iterArgs(arity int, item box.Value, idx int) []box.Value {
	if arity == 2 {
		return []box.Value{item, box.Number(idx)}
	}
	return []box.Value{item}
}

func reduceArgs(arity int, acc, item box.Value, idx int) ([]box.Value, error) {
	switch arity {
	case 2:
		return []box.Value{acc, item}, nil
	case 3:
		return []box.Value{acc, item, box.Number(idx)}, nil
	default:
		return nil, fmt.Errorf("reducer callable must have arity 2 or 3, got %d", arity)
	}
}

func implAny(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	if col.IsUndefined() {
		return box.Bool(false), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("any: first argument must be a list")
	}
	fn := args[1]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	for idx, item := range list {
		res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			return box.Bool(true), nil
		}
	}
	return box.Bool(false), nil
}

func implAll(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	if col.IsUndefined() {
		return box.Bool(false), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("all: first argument must be a list")
	}
	fn := args[1]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	for idx, item := range list {
		res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		if !box.TrinaryFrom(res).IsTrue() {
			return box.Bool(false), nil
		}
	}
	return box.Bool(true), nil
}

func implFirst(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	if col.IsUndefined() {
		return box.Undefined(), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("first: first argument must be a list")
	}
	fn := args[1]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	for idx, item := range list {
		res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			return item, nil
		}
	}
	return box.Undefined(), nil
}

func implFilter(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	if col.IsUndefined() {
		return box.List(nil), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("filter: first argument must be a list")
	}
	fn := args[1]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			out = append(out, item)
		}
	}
	return box.List(out), nil
}

func implCollect(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("collect: first argument must be a list")
	}
	fn := args[1]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		out = append(out, res)
	}
	return box.List(out), nil
}

func implReduce(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	col := args[0]
	if col.IsUndefined() {
		return box.Undefined(), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("reduce: first argument must be a list")
	}
	acc := args[1]
	fn := args[2]
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	for idx, item := range list {
		callArgs, err := reduceArgs(arity, acc, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		next, err := env.Call(ctx, fn, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		acc = next
	}
	return acc, nil
}

func implDistinct(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if len(args) == 1 {
		return distinctDirect(args[0])
	}
	return distinctSelector(ctx, env, args[0], args[1])
}

func distinctDirect(col box.Value) (box.Value, error) {
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("distinct: first argument must be a list")
	}
	if len(list) < 2 {
		return box.List(slices.Clone(list)), nil
	}
	seen := make(map[string]struct{}, len(list))
	out := make([]box.Value, 0, len(list))
	for _, item := range list {
		k, err := scalarFingerprint(item)
		if err != nil {
			return box.Undefined(), err
		}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, item)
	}
	return box.List(out), nil
}

func distinctSelector(ctx context.Context, env Env, col, fn box.Value) (box.Value, error) {
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("distinct: first argument must be a list")
	}
	arity, err := env.CallableArity(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if len(list) < 2 {
		return box.List(slices.Clone(list)), nil
	}
	seen := make(map[string]struct{}, len(list))
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		keyVal, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
		if err != nil {
			return box.Undefined(), err
		}
		k, err := scalarFingerprint(keyVal)
		if err != nil {
			return box.Undefined(), fmt.Errorf("distinct key: %w", err)
		}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, item)
	}
	return box.List(out), nil
}

func scalarFingerprint(v box.Value) (string, error) {
	switch v.Kind() {
	case box.ValueUndefined:
		return "undef:", nil
	case box.ValueNull:
		return "null:", nil
	case box.ValueBool:
		b, _ := v.BoolValue()
		return fmt.Sprintf("bool:%v", b), nil
	case box.ValueNumber:
		n, _ := v.NumberValue()
		return fmt.Sprintf("num:%.17g", n), nil
	case box.ValueString:
		s, _ := v.StringValue()
		return "str:" + s, nil
	case box.ValueTrinary:
		t, _ := v.TrinaryValue()
		return fmt.Sprintf("tri:%d", t), nil
	default:
		return "", fmt.Errorf("unsupported key kind %s for distinct (expected string, number, bool, trinary, null, or undefined)", v.Kind())
	}
}
