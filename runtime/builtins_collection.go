// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package runtime

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
)

// BuiltinAny reports whether any element satisfies the predicate callable.
func BuiltinAny(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("any requires 2 arguments")
	}
	col := args[0]
	if col.IsUndefined() {
		return box.Bool(false), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("any: first argument must be a list")
	}
	fn := args[1]
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("any: callable must have arity 1 or 2")
	}
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		res, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			return box.Bool(true), nil
		}
	}
	return box.Bool(false), nil
}

// BuiltinAll reports whether every element satisfies the predicate callable.
func BuiltinAll(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("all requires 2 arguments")
	}
	col := args[0]
	if col.IsUndefined() {
		return box.Bool(false), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("all: first argument must be a list")
	}
	fn := args[1]
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("all: callable must have arity 1 or 2")
	}
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		res, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		if !box.TrinaryFrom(res).IsTrue() {
			return box.Bool(false), nil
		}
	}
	return box.Bool(true), nil
}

// BuiltinFirst returns the first item satisfying the predicate, or undefined.
func BuiltinFirst(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("first requires 2 arguments")
	}
	col := args[0]
	if col.IsUndefined() {
		return box.Undefined(), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("first: first argument must be a list")
	}
	fn := args[1]
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("first: callable must have arity 1 or 2")
	}
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		res, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			return item, nil
		}
	}
	return box.Undefined(), nil
}

// BuiltinFilter returns items for which the predicate is true.
func BuiltinFilter(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("filter requires 2 arguments")
	}
	col := args[0]
	if col.IsUndefined() {
		return box.List(nil), nil
	}
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("filter: first argument must be a list")
	}
	fn := args[1]
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("filter: callable must have arity 1 or 2")
	}
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		res, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		if box.TrinaryFrom(res).IsTrue() {
			out = append(out, item)
		}
	}
	return box.List(out), nil
}

// BuiltinMap maps each element through the callable.
func BuiltinMap(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("map requires 2 arguments")
	}
	col := args[0]
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("map: first argument must be a list")
	}
	fn := args[1]
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("map: callable must have arity 1 or 2")
	}
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		res, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		out = append(out, res)
	}
	return box.List(out), nil
}

// BuiltinReduce folds the list with an initial accumulator using the reducer callable.
func BuiltinReduce(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 3 {
		return box.Undefined(), fmt.Errorf("reduce requires 3 arguments")
	}
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
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 2 && c.Arity() != 3 {
		return box.Undefined(), fmt.Errorf("reduce: reducer must have arity 2 or 3")
	}
	for idx, item := range list {
		callArgs, err := reduceArgs(site, c, acc, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		next, err := invokeCallable(ctx, site, c, callArgs)
		if err != nil {
			return box.Undefined(), err
		}
		acc = next
	}
	return acc, nil
}

// BuiltinDistinct removes duplicates: either by scalar identity of elements, or by a key selector.
func BuiltinDistinct(ctx context.Context, site *CallSite, args ...box.Value) (box.Value, error) {
	switch len(args) {
	case 1:
		return builtinDistinctDirect(args[0])
	case 2:
		return builtinDistinctSelector(ctx, site, args[0], args[1])
	default:
		return box.Undefined(), fmt.Errorf("distinct requires 1 or 2 arguments")
	}
}

func builtinDistinctDirect(col box.Value) (box.Value, error) {
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

func builtinDistinctSelector(ctx context.Context, site *CallSite, col, fn box.Value) (box.Value, error) {
	list, ok := col.ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("distinct: first argument must be a list")
	}
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	if c.Arity() != 1 && c.Arity() != 2 {
		return box.Undefined(), fmt.Errorf("distinct: selector must have arity 1 or 2")
	}
	if len(list) < 2 {
		return box.List(slices.Clone(list)), nil
	}
	seen := make(map[string]struct{}, len(list))
	out := make([]box.Value, 0, len(list))
	for idx, item := range list {
		callArgs, err := iterArgs(site, c, item, idx)
		if err != nil {
			return box.Undefined(), err
		}
		keyVal, err := invokeCallable(ctx, site, c, callArgs)
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

// scalarFingerprint builds a stable dedupe key for supported scalar kinds.
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
