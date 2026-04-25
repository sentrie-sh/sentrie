// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/binaek/perch"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/runtime/js"
)

func testExecutorForModuleBinding() *executorImpl {
	idx := index.CreateIndex()
	idx.Pack = &pack.PackFile{Location: "."}
	exec := &executorImpl{
		index:              idx,
		jsRegistry:         js.NewRegistry("."),
		moduleBindingPerch: perch.New[*ModuleBinding](1 << 20),
		callMemoizePerch:   perch.New[any](1 << 20),
	}
	exec.jsRegistry.RegisterGoBuiltin("hash", js.BuiltinHashGo)
	exec.moduleBindingPerch.Reserve()
	exec.callMemoizePerch.Reserve()
	return exec
}

func (s *RuntimeTestSuite) TestGetModuleBindingCachesBindings() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"md5"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	ms, err := exec.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, ".")
	s.Require().NoError(err)

	first, loaded, err := exec.getModuleBinding(context.Background(), use, ms)
	s.Require().NoError(err)
	_ = loaded
	s.NotNil(first)
	s.Equal(ms.KeyOrPath(), first.CanonicalKey)
	s.Equal(use.As, first.Alias)

	second, loaded, err := exec.getModuleBinding(context.Background(), use, ms)
	s.Require().NoError(err)
	_ = loaded
	s.NotNil(second)
	s.Equal(ms.KeyOrPath(), second.CanonicalKey)
	s.Equal(use.As, second.Alias)
}

func (s *RuntimeTestSuite) TestJSBindingConstructorRejectsMissingRequestedExport() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"doesNotExist"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	ms, err := exec.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, ".")
	s.Require().NoError(err)

	_, err = exec.jsBindingConstructor(context.Background(), use, ms)
	s.Require().Error(err)
	s.Contains(err.Error(), "missing required export")
}

func (s *RuntimeTestSuite) TestBindUsesBindsPreparedModule() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"md5"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	policy := &index.Policy{
		FilePath: "policy.sentra",
		Uses:     map[string]*ast.UseStatement{"hash": use},
	}
	ec := NewExecutionContext(policy, exec)

	err := exec.bindUses(context.Background(), ec, policy)
	s.Require().NoError(err)
	binding, ok := ec.Module("hash")
	s.True(ok)
	s.NotNil(binding)
}
