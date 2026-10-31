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
	kindPath = []box.ValueKind{box.ValueString, box.ValueNumber, box.ValueList}

	declObjectGet = &Decl{
		Name: "object_get",
		Description: "Reads a value from a dict (or list) by path, or returns the default when the path does not exist. " +
			"The path is a list of segments (a string selects a dict key, a non-negative integer indexes a list) " +
			"or a single segment. An undefined or null object yields the default.",
		DeriveSafe: true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "object"},
				{Name: "path", Kinds: kindPath, KindError: "object_get: path must be a string, a number or a list of segments"},
				{Name: "default"},
			},
			TooFewError:  "object_get requires 3 arguments",
			TooManyError: "object_get requires 3 arguments",
			Result:       ParamSig{Name: "result"},
		},
		Impl: implObjectGet,
	}

	declObjectKeys = &Decl{
		Name:        "object_keys",
		Description: "Returns the keys of a dict as a list of strings in ascending order.",
		DeriveSafe:  true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "object", Kinds: kindDict, KindError: "object_keys: argument must be a dict"}},
			TooFewError:  "object_keys requires 1 argument",
			TooManyError: "object_keys requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implObjectKeys,
	}

	declObjectRemove = &Decl{
		Name:        "object_remove",
		Description: "Returns a copy of a dict without the listed top-level keys.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "object", Kinds: kindDict, KindError: "object_remove: first argument must be a dict"},
				{Name: "keys", Kinds: kindList, KindError: "object_remove: keys must be a list of strings"},
			},
			TooFewError:  "object_remove requires 2 arguments",
			TooManyError: "object_remove requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindDict},
		},
		Impl: implObjectRemove,
	}

	declObjectFilter = &Decl{
		Name:        "object_filter",
		Description: "Returns a copy of a dict holding only the listed top-level keys.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "object", Kinds: kindDict, KindError: "object_filter: first argument must be a dict"},
				{Name: "keys", Kinds: kindList, KindError: "object_filter: keys must be a list of strings"},
			},
			TooFewError:  "object_filter requires 2 arguments",
			TooManyError: "object_filter requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindDict},
		},
		Impl: implObjectFilter,
	}

	declObjectUnion = &Decl{
		Name:        "object_union",
		Description: "Deep-merges two dicts into a new dict; on a conflict the second wins, and non-dict values (lists included) are replaced, not combined.",
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "first", Kinds: kindDict, KindError: "object_union: first argument must be a dict"},
				{Name: "second", Kinds: kindDict, KindError: "object_union: second argument must be a dict"},
			},
			TooFewError:  "object_union requires 2 arguments",
			TooManyError: "object_union requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: kindDict},
		},
		Impl: implObjectUnion,
	}
)

// objectPathSegments normalises a path argument to its segments.
func objectPathSegments(v box.Value) ([]box.Value, error) {
	if list, ok := v.ListValue(); ok {
		for _, seg := range list {
			switch seg.Kind() {
			case box.ValueString, box.ValueNumber:
			default:
				return nil, fmt.Errorf("object_get: path segments must be strings or non-negative integers, got %s", seg.Kind())
			}
		}
		return list, nil
	}
	return []box.Value{v}, nil
}

func implObjectGet(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	if args[1].IsUndefined() {
		return box.Undefined(), nil
	}
	if args[1].IsNull() {
		return box.Undefined(), fmt.Errorf("object_get: path must be a string, a number or a list of segments")
	}
	segs, err := objectPathSegments(args[1])
	if err != nil {
		return box.Undefined(), err
	}
	cur := args[0]
	for _, seg := range segs {
		next, found := objectStep(cur, seg)
		if !found {
			return args[2], nil
		}
		cur = next
	}
	if cur.IsUndefined() || (len(segs) == 0 && cur.IsNull()) {
		// An absent object (undefined or null) yields the default even for an
		// empty path; a key that is present with a null value is returned.
		return args[2], nil
	}
	return cur, nil
}

// objectStep descends one path segment. It reports false when the segment does
// not exist (a missing key, an out-of-range or non-integer index, or a
// non-container value).
func objectStep(cur, seg box.Value) (box.Value, bool) {
	if m, ok := cur.DictValue(); ok {
		key, ok := seg.StringValue()
		if !ok {
			return box.Undefined(), false
		}
		v, ok := m[key]
		return v, ok
	}
	if xs, ok := cur.ListValue(); ok {
		n, ok := seg.NumberValue()
		if !ok || n < 0 || n != float64(int(n)) || int(n) >= len(xs) {
			return box.Undefined(), false
		}
		return xs[int(n)], true
	}
	return box.Undefined(), false
}

func implObjectKeys(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	if args[0].IsUndefined() {
		return box.Undefined(), nil
	}
	m, ok := args[0].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("object_keys: argument must be a dict")
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make([]box.Value, len(keys))
	for i, k := range keys {
		out[i] = box.String(k)
	}
	return box.List(out), nil
}

// objectKeyList reads a list of strings for object_remove and object_filter.
func objectKeyList(fn string, v box.Value) (map[string]struct{}, error) {
	list, ok := v.ListValue()
	if !ok {
		return nil, fmt.Errorf("%s: keys must be a list of strings", fn)
	}
	set := make(map[string]struct{}, len(list))
	for _, k := range list {
		s, ok := k.StringValue()
		if !ok {
			return nil, fmt.Errorf("%s: keys must be a list of strings", fn)
		}
		set[s] = struct{}{}
	}
	return set, nil
}

func objectSelect(fn string, keep bool, args []box.Value) (box.Value, error) {
	if args[0].IsUndefined() || args[1].IsUndefined() {
		return box.Undefined(), nil
	}
	m, ok := args[0].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("%s: first argument must be a dict", fn)
	}
	set, err := objectKeyList(fn, args[1])
	if err != nil {
		return box.Undefined(), err
	}
	out := make(map[string]box.Value, len(m))
	for k, v := range copyMapDeep(m) {
		if _, listed := set[k]; listed == keep {
			out[k] = v
		}
	}
	return box.Dict(out), nil
}

func implObjectRemove(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	return objectSelect("object_remove", false, args)
}

func implObjectFilter(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	return objectSelect("object_filter", true, args)
}

func implObjectUnion(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	if args[0].IsUndefined() || args[1].IsUndefined() {
		return box.Undefined(), nil
	}
	m1, ok := args[0].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("object_union: first argument must be a dict")
	}
	m2, ok := args[1].DictValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("object_union: second argument must be a dict")
	}
	return box.Dict(mergeValueDicts(m1, m2)), nil
}
