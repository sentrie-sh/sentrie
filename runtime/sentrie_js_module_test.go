// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/runtime/js"
)

// TestSentrieJSNowExportRemoved guards #207: `use { now } from @sentrie/js`
// fails loudly at bind time instead of handing the policy an unpinned
// millisecond clock. Other exports of the module still bind.
func (s *RuntimeTestSuite) TestSentrieJSNowExportRemoved() {
	exec := testExecutorForModuleBinding()
	exec.jsRegistry.RegisterTSBuiltin("js", string(js.BuiltinJSTS))

	for _, tc := range []struct {
		name    string
		export  string
		wantErr bool
	}{
		{name: "now is gone", export: "now", wantErr: true},
		{name: "round still binds", export: "round"},
	} {
		s.Run(tc.name, func() {
			use := ast.NewUseStatement([]string{tc.export}, "", []string{constants.APPNAME, "js"}, "js", stubRange())
			ms, err := exec.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, ".")
			s.Require().NoError(err)

			_, err = exec.jsBindingConstructor(s.T().Context(), use, ms)
			if tc.wantErr {
				s.Require().Error(err)
				s.Contains(err.Error(), "missing required export")
				s.Contains(err.Error(), `"now"`)
				return
			}
			s.Require().NoError(err)
		})
	}
}
