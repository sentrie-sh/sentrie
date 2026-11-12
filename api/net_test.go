// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import "net"

func (s *APITestSuite) TestResolveBindingsIPv6HostsAreBracketedOnce() {
	for _, name := range []string{"local", "local4", "local6", "network", "network4", "network6"} {
		got, err := resolveBindings(8080, []string{name})
		s.Require().NoError(err, name)
		s.Require().Len(got, 1, name)
		s.NotContains(got[0], "[[", name)
		s.NotContains(got[0], "]]", name)
		host, port, err := net.SplitHostPort(got[0])
		s.Require().NoError(err, got[0])
		s.Equal("8080", port)
		s.NotContains(host, "[", name)
		if name != "local" && name != "network" {
			s.NotNil(net.ParseIP(host), name)
		}
	}
}

func (s *APITestSuite) TestResolveBindingsIPv6Exact() {
	got, err := resolveBindings(7529, []string{"local6"})
	s.Require().NoError(err)
	s.Equal([]string{"[::1]:7529"}, got)

	got, err = resolveBindings(7529, []string{"network6"})
	s.Require().NoError(err)
	s.Equal([]string{"[::]:7529"}, got)
}

func (s *APITestSuite) TestResolveBindingsMixedLists() {
	got, err := resolveBindings(9000, []string{"127.0.0.1", "::1", "example.com"})
	s.Require().NoError(err)
	s.Equal([]string{"127.0.0.1:9000", "[::1]:9000", "example.com:9000"}, got)

	_, err = resolveBindings(9000, []string{"local6", "127.0.0.1"})
	s.Require().Error(err)
	s.Contains(err.Error(), "exactly one address")
}

func (s *APITestSuite) TestResolveBindingsEmptyListIsAnError() {
	for _, listen := range [][]string{nil, {}} {
		var (
			got []string
			err error
		)
		s.NotPanics(func() { got, err = resolveBindings(7529, listen) })
		s.Require().Error(err)
		s.Equal("at least one listen address is required", err.Error())
		s.Nil(got)
	}
}

func (s *APITestSuite) TestResolveBindingsRejectsEmptyAddress() {
	for _, listen := range [][]string{{""}, {"", ""}, {"127.0.0.1", ""}, {"  "}} {
		got, err := resolveBindings(7529, listen)
		s.Require().Error(err, "%q", listen)
		s.Equal("listen address must not be empty; use network to listen on all interfaces", err.Error())
		s.Nil(got)
	}
}

func (s *APITestSuite) TestSetupRejectsEmptyListenList() {
	api := NewHTTPAPI(nil)
	var err error
	s.NotPanics(func() { err = api.Setup(s.T().Context(), 0, nil) })
	s.Require().Error(err)
	s.Contains(err.Error(), "at least one listen address is required")
}
