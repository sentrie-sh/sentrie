// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"

	"github.com/sentrie-sh/sentrie/pack"
)

func (s *CmdTestSuite) TestUnenforcedPermissionWarnings() {
	cases := []struct {
		name string
		p    *pack.PackFile
		want []string
		not  []string
	}{
		{"nil pack", nil, nil, []string{"warning"}},
		{"no permissions", &pack.PackFile{}, nil, []string{"warning"}},
		{"empty lists", &pack.PackFile{Permissions: &pack.Permissions{FSRead: []string{}, Net: []string{}}}, nil, []string{"warning"}},
		{"env only", &pack.PackFile{Permissions: &pack.Permissions{Env: []string{"HOME"}}}, nil, []string{"warning"}},
		{"fs_read", &pack.PackFile{Permissions: &pack.Permissions{FSRead: []string{"/data"}}}, []string{`"fs_read"`, "not enforced"}, []string{`"net"`}},
		{"net", &pack.PackFile{Permissions: &pack.Permissions{Net: []string{"example.com"}}}, []string{`"net"`, "not enforced"}, []string{`"fs_read"`}},
		{"both", &pack.PackFile{Permissions: &pack.Permissions{FSRead: []string{"/data"}, Net: []string{"example.com"}}}, []string{`"fs_read"`, `"net"`}, nil},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			var buf bytes.Buffer
			s.Require().NoError(printUnenforcedPermissionWarnings(&buf, tc.p))
			for _, w := range tc.want {
				s.Contains(buf.String(), w)
			}
			for _, n := range tc.not {
				s.NotContains(buf.String(), n)
			}
		})
	}
}
