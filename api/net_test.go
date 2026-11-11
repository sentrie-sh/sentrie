// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

func (s *APITestSuite) TestResolveBindingsIPv6AndEmpty() {
	got, err := resolveBindings(7529, []string{"local6"})
	s.Require().NoError(err)
	s.Equal([]string{"[::1]:7529"}, got)

	got, err = resolveBindings(7529, []string{"network6"})
	s.Require().NoError(err)
	s.Equal([]string{"[::]:7529"}, got)

	_, err = resolveBindings(7529, nil)
	s.Error(err)
}
