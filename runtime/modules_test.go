// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/jackc/puddle/v2"
	"github.com/sentrie-sh/sentrie/box"
)

func testModuleBindingWithExports(rt *goja.Runtime, exports map[string]goja.Value) ModuleBinding {
	instance := &JSInstance{rt: rt, exports: exports}
	pool, err := puddle.NewPool(&puddle.Config[*JSInstance]{
		Constructor: func(ctx context.Context) (*JSInstance, error) {
			return instance, nil
		},
		Destructor: func(res *JSInstance) {
			res.rt.ClearInterrupt()
		},
		MaxSize: 10,
	})
	if err != nil {
		panic(err)
	}
	return ModuleBinding{
		Alias:        "mod",
		instancePool: pool,
	}
}

func (s *RuntimeTestSuite) TestModuleBindingCallHandlesNullAndUndefinedReturns() {
	ctx := s.T().Context()
	rt := goja.New()
	exportsVal, err := rt.RunString(`({
		returnNull: function() { return null; },
		returnUndefined: function() { return undefined; },
	})`)
	s.Require().NoError(err)
	obj := exportsVal.ToObject(rt)
	exports := map[string]goja.Value{
		"returnNull":      obj.Get("returnNull"),
		"returnUndefined": obj.Get("returnUndefined"),
	}
	binding := testModuleBindingWithExports(rt, exports)
	ec := NewExecutionContext(newEvalTestPolicy(), &executorImpl{})

	nullOut, err := binding.Call(ctx, ec, "returnNull")
	s.Require().NoError(err)
	s.Require().Nil(nullOut)
	nullVal := box.FromBoundaryAny(nullOut)
	s.Require().True(nullVal.IsNull())

	undefOut, err := binding.Call(ctx, ec, "returnUndefined")
	s.Require().NoError(err)
	undefVal := box.FromBoundaryAny(undefOut)
	s.Require().True(undefVal.IsUndefined())
}

func (s *RuntimeTestSuite) TestModuleBindingCallMissingExportMessage() {
	ctx := s.T().Context()
	rt := goja.New()
	binding := testModuleBindingWithExports(rt, map[string]goja.Value{})
	ec := NewExecutionContext(newEvalTestPolicy(), &executorImpl{})

	_, err := binding.Call(ctx, ec, "missing")
	s.Require().Error(err)
	s.Require().Equal("function missing not found in module \"mod\"", err.Error())
}

func (s *RuntimeTestSuite) TestModuleBindingCallSuccessAndErrorPaths() {
	ctx := s.T().Context()
	ec := NewExecutionContext(newEvalTestPolicy(), &executorImpl{})

	nilPool := ModuleBinding{Alias: "mod"}
	_, err := nilPool.Call(ctx, ec, "fn")
	s.Require().Error(err)
	s.Require().Equal("module has no JS binding", err.Error())

	failPool, err := puddle.NewPool(&puddle.Config[*JSInstance]{
		Constructor: func(ctx context.Context) (*JSInstance, error) {
			return nil, fmt.Errorf("pool construct failed")
		},
		MaxSize: 1,
	})
	s.Require().NoError(err)
	failBinding := ModuleBinding{Alias: "mod", instancePool: failPool}
	_, err = failBinding.Call(ctx, ec, "fn")
	s.Require().Error(err)
	s.Require().Equal("pool construct failed", err.Error())

	rt := goja.New()
	exportsVal, err := rt.RunString(`({
		returnString: function() { return "ok"; },
		returnBool: function() { return true; },
		returnMap: function() { return { a: 1, b: "two" }; },
		returnSlice: function() { return [1, 2, 3]; },
		returnNumber: function() { return 42; },
		returnFn: function() { return function() { return 1; }; },
		throwError: function() { throw new Error("boom"); },
		withArg: function(x) { return x; },
	})`)
	s.Require().NoError(err)
	obj := exportsVal.ToObject(rt)
	exports := map[string]goja.Value{
		"returnString": obj.Get("returnString"),
		"returnBool":   obj.Get("returnBool"),
		"returnMap":    obj.Get("returnMap"),
		"returnSlice":  obj.Get("returnSlice"),
		"returnNumber": obj.Get("returnNumber"),
		"returnFn":     obj.Get("returnFn"),
		"throwError":   obj.Get("throwError"),
		"withArg":      obj.Get("withArg"),
	}
	binding := testModuleBindingWithExports(rt, exports)

	strOut, err := binding.Call(ctx, ec, "returnString")
	s.Require().NoError(err)
	s.Equal("ok", strOut)

	boolOut, err := binding.Call(ctx, ec, "returnBool")
	s.Require().NoError(err)
	s.Equal(true, boolOut)

	mapOut, err := binding.Call(ctx, ec, "returnMap")
	s.Require().NoError(err)
	mapTyped, ok := mapOut.(map[string]any)
	s.Require().True(ok)
	s.Equal(int64(1), mapTyped["a"])
	s.Equal("two", mapTyped["b"])

	sliceOut, err := binding.Call(ctx, ec, "returnSlice")
	s.Require().NoError(err)
	sliceTyped, ok := sliceOut.([]any)
	s.Require().True(ok)
	s.Len(sliceTyped, 3)
	s.Equal(int64(1), sliceTyped[0])
	s.Equal(int64(2), sliceTyped[1])
	s.Equal(int64(3), sliceTyped[2])

	numOut, err := binding.Call(ctx, ec, "returnNumber")
	s.Require().NoError(err)
	s.Equal(int64(42), numOut)

	undefArgOut, err := binding.Call(ctx, ec, "withArg", box.ToBoundaryAny(box.Undefined()))
	s.Require().NoError(err)
	s.True(box.FromBoundaryAny(undefArgOut).IsUndefined())

	nonCallable := testModuleBindingWithExports(rt, map[string]goja.Value{
		"notFn": rt.ToValue("not callable"),
	})
	_, err = nonCallable.Call(ctx, ec, "notFn")
	s.Require().Error(err)
	s.Require().Equal("export '\"notFn\"' is not callable", err.Error())

	_, err = binding.Call(ctx, ec, "returnFn")
	s.Require().Error(err)
	s.Contains(err.Error(), "unexpected return type")

	_, err = binding.Call(ctx, ec, "throwError")
	s.Require().Error(err)
	s.Contains(err.Error(), "boom")

	type moduleStruct struct {
		Field string
	}
	structFn := func() moduleStruct {
		return moduleStruct{Field: "struct"}
	}
	s.Require().NoError(rt.Set("returnStruct", structFn))
	structBinding := testModuleBindingWithExports(rt, map[string]goja.Value{
		"returnStruct": rt.Get("returnStruct"),
	})
	structOut, err := structBinding.Call(ctx, ec, "returnStruct")
	s.Require().NoError(err)
	s.Equal(map[string]any{"Field": "struct"}, structOut)
}
