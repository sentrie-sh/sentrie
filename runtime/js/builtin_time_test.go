// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import "github.com/dop251/goja"

// TestTimeNowIsPinned checks that @sentrie/time.now() reports the execution
// start in Unix seconds and throws, rather than reading the wall clock, when
// the VM was not given one (#153).
func (s *JSTestSuite) TestTimeNowIsPinned() {
	got, err := s.runBuiltinScript(BuiltinTimeGo, `[m.now(), m.now()]`)
	s.Require().NoError(err)
	s.Equal([]any{testExecutionStartUnix, testExecutionStartUnix}, got.Export())

	vm := goja.New()
	ex, err := BuiltinTimeGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("m", ex))
	_, err = vm.RunString(`m.now()`)
	var exc *goja.Exception
	s.Require().ErrorAs(err, &exc)
	s.Contains(err.Error(), "execution start time is not set")
}
