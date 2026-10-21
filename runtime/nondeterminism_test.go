// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/runtime/js"
)

// Every module in the nondeterminism table must be one NewExecutor registers,
// and every listed export must exist, so the table cannot drift from the
// modules it describes.
func (s *RuntimeTestSuite) TestNondeterministicExportsExist() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  rule ok = { yield true }
  export decision of ok
}`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "nondet.sentrie", src: src})

	modules := js.NondeterministicModules()
	s.Require().NotEmpty(modules)
	for _, key := range modules {
		ar := js.NewAliasRuntime(exec.jsRegistry, ".")
		exports, err := ar.Require(ctx, ".", key)
		s.Require().NoError(err, "module %s", key)
		for _, name := range js.NondeterministicExports(key) {
			_, ok := goja.AssertFunction(exports.Get(name))
			s.True(ok, "%s.%s must be an exported function", key, name)
		}
	}
}
