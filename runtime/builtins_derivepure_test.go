// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/runtime/derivepure"
)

func (s *RuntimeTestSuite) TestDerivePureBuiltinNamesAreRegistered() {
	for _, name := range derivepure.PureBuiltinNames() {
		_, ok := Builtins[name]
		s.True(ok, "pure builtin %q must exist in Builtins", name)
	}
}
