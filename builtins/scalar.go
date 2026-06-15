// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/xerr"
)

var (
	declMerge = &Decl{
		Name:        "merge",
		Description: "Recursively merges two dict values into a new dict.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "first", Kinds: kindDict, KindError: "first argument is not a dict"},
				{Name: "second", Kinds: kindDict, KindError: "second argument is not a dict"},
			},
			TooFewError:  "merge requires 2 arguments",
			TooManyError: "merge requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindDict},
		},
		Impl: implMerge,
	}

	declCount = &Decl{
		Name:        "count",
		Description: "Returns the length of a list, string, or dict.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{{
				Name:       "collection",
				Kinds:      kindCountable,
				OnMismatch: MismatchUndefined,
			}},
			TooFewError:  "count requires 1 argument",
			TooManyError: "count requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindNumber},
		},
		Impl: implCount,
	}

	declError = &Decl{
		Name:        "error",
		Description: "Short-circuits execution with a formatted error.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{{Name: "format"}},
			Variadic: &ParamSig{Name: "args"},
			TooFewError: "error requires at least 1 argument",
			Result:      ParamSig{Name: "result"},
		},
		Impl: implError,
	}

	declFlatten = &Decl{
		Name:        "flatten",
		Description: "Flattens nested lists to a controlled depth.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "collection", Kinds: kindList, KindError: "flatten: first argument must be a list"},
				{Name: "depth", Optional: true},
			},
			TooFewError:  "flatten requires 1 or 2 arguments",
			TooManyError: "flatten requires 1 or 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implFlatten,
	}

	declFlattenDeep = &Decl{
		Name:        "flatten_deep",
		Description: "Recursively flattens nested lists.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{{
				Name:      "collection",
				Kinds:     kindList,
				KindError: "flatten_deep: argument must be a list",
			}},
			TooFewError:  "flatten_deep requires 1 argument",
			TooManyError: "flatten_deep requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implFlattenDeep,
	}

	declAsList = &Decl{
		Name:        "as_list",
		Description: "Normalizes one-or-many inputs to a list.",
		DeriveSafe:  true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "value"}},
			TooFewError:  "as_list requires 1 argument",
			TooManyError: "as_list requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implAsList,
	}

	declNormaliseList = &Decl{
		Name:        "normalise_list",
		Description: "Normalizes messy list inputs with one level of nesting.",
		DeriveSafe:  true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "value"}},
			TooFewError:  "normalise_list requires 1 argument",
			TooManyError: "normalise_list requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implNormaliseList,
	}

	declNow = &Decl{
		Name:        "now",
		Description: "Returns the policy execution start time as epoch milliseconds.",
		DeriveSafe:  true,
		Sig: Sig{
			TooFewError:  "now requires 0 arguments",
			TooManyError: "now requires 0 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindNumber},
		},
		Impl: implNow,
	}
)

func isUndefinedV(v box.Value) bool {
	return v.IsUndefined()
}

func toIntV(v box.Value) (int64, bool) {
	n, ok := v.NumberValue()
	if !ok {
		return 0, false
	}
	return int64(n), true
}

func copyMapDeep(m map[string]box.Value) map[string]box.Value {
	out := make(map[string]box.Value, len(m))
	for k, v := range m {
		if vm, ok := v.DictValue(); ok {
			out[k] = box.Dict(copyMapDeep(vm))
		} else {
			out[k] = v
		}
	}
	return out
}

func mergeValueDicts(map1, map2 map[string]box.Value) map[string]box.Value {
	result := copyMapDeep(map1)
	for key, value2 := range map2 {
		if existing, exists := result[key]; exists {
			m1, ok1 := existing.DictValue()
			m2, ok2 := value2.DictValue()
			if ok1 && ok2 {
				result[key] = box.Dict(mergeValueDicts(m1, m2))
				continue
			}
		}
		if nestedMap, ok := value2.DictValue(); ok {
			result[key] = box.Dict(copyMapDeep(nestedMap))
			continue
		}
		result[key] = value2
	}
	return result
}

func implMerge(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	m1, ok := args[0].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("first argument is not a dict")
	}
	m2, ok := args[1].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("second argument is not a dict")
	}
	return box.Dict(mergeValueDicts(m1, m2)), nil
}

func implCount(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if xs, ok := args[0].ListValue(); ok {
		return box.Number(len(xs)), nil
	}
	if s, ok := args[0].StringValue(); ok {
		return box.Number(len(s)), nil
	}
	if m, ok := args[0].DictValue(); ok {
		return box.Number(len(m)), nil
	}
	return box.Undefined(), nil
}

func implError(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	fa := args
	if len(fa) == 1 {
		fa = append([]box.Value{box.String("%v")}, fa...)
	}
	format, ok := fa[0].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("error: first argument must be a format string")
	}
	rest := make([]any, 0, len(fa)-1)
	for _, a := range fa[1:] {
		x, err := box.TryToBoundaryAny(a)
		if err != nil {
			return box.Undefined(), fmt.Errorf("error: %w", err)
		}
		rest = append(rest, x)
	}
	return box.Undefined(), xerr.ErrInjected(format, rest...)
}

func implFlatten(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	x, ok := args[0].ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("flatten: first argument must be a list")
	}
	var depth int64 = 1
	if len(args) == 2 {
		if isUndefinedV(args[1]) {
			return box.Undefined(), nil
		}
		n, ok := toIntV(args[1])
		if !ok {
			return box.Undefined(), fmt.Errorf("flatten: second argument must be a non-negative integer")
		}
		if n < 0 {
			return box.Undefined(), fmt.Errorf("flatten: depth must be a non-negative integer")
		}
		depth = n
	}
	if depth == 0 {
		return box.List(x), nil
	}
	return flattenListBox(x, depth)
}

func flattenListBox(x []box.Value, depth int64) (box.Value, error) {
	if depth == 0 {
		return box.List(x), nil
	}
	result := make([]box.Value, 0)
	for _, elem := range x {
		if isUndefinedV(elem) {
			return box.Undefined(), nil
		}
		if nestedList, ok := elem.ListValue(); ok {
			for _, nestedElem := range nestedList {
				if isUndefinedV(nestedElem) {
					return box.Undefined(), nil
				}
			}
			flattened, err := flattenListBox(nestedList, depth-1)
			if err != nil {
				return box.Undefined(), err
			}
			if flattened.IsUndefined() {
				return box.Undefined(), nil
			}
			sub, _ := flattened.ListValue()
			result = append(result, sub...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

func implFlattenDeep(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	x, ok := args[0].ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("flatten_deep: argument must be a list")
	}
	return flattenDeepBox(x)
}

func flattenDeepBox(x []box.Value) (box.Value, error) {
	result := make([]box.Value, 0)
	for _, elem := range x {
		if isUndefinedV(elem) {
			return box.Undefined(), nil
		}
		if nestedList, ok := elem.ListValue(); ok {
			flattened, err := flattenDeepBox(nestedList)
			if err != nil {
				return box.Undefined(), err
			}
			if flattened.IsUndefined() {
				return box.Undefined(), nil
			}
			sub, _ := flattened.ListValue()
			result = append(result, sub...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

func implAsList(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	v := args[0]
	if list, ok := v.ListValue(); ok {
		for _, elem := range list {
			if isUndefinedV(elem) {
				return box.Undefined(), nil
			}
		}
		return box.List(list), nil
	}
	return box.List([]box.Value{v}), nil
}

func implNormaliseList(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	v := args[0]
	var list []box.Value
	if l, ok := v.ListValue(); ok {
		list = l
	} else {
		list = []box.Value{v}
	}
	if slices.ContainsFunc(list, isUndefinedV) {
		return box.Undefined(), nil
	}
	for _, elem := range list {
		if nestedList, ok := elem.ListValue(); ok {
			for _, nestedElem := range nestedList {
				if isUndefinedV(nestedElem) {
					return box.Undefined(), nil
				}
				if _, ok := nestedElem.ListValue(); ok {
					return box.Undefined(), fmt.Errorf("normalise_list: input contains deeper than one level of nesting")
				}
			}
		}
	}
	result := make([]box.Value, 0)
	for _, elem := range list {
		if nestedList, ok := elem.ListValue(); ok {
			result = append(result, nestedList...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

func implNow(ctx context.Context, env Env, _ ...box.Value) (box.Value, error) {
	t := env.ExecutionStart()
	return box.Number(float64(t.UnixMilli())), nil
}
