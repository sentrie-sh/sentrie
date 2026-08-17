// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

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
