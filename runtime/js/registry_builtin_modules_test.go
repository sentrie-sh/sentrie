// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

func (s *JSTestSuite) TestBuiltinModules() {
	reg := NewRegistry(s.T().TempDir())
	s.Empty(reg.BuiltinModules())

	reg.RegisterGoBuiltin("zeta", nil)
	reg.RegisterTSBuiltin("alpha", "export {}")
	reg.RegisterGoBuiltin("mid", nil)
	s.Equal([]string{"@sentrie/alpha", "@sentrie/mid", "@sentrie/zeta"}, reg.BuiltinModules())
}
