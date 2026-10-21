// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"time"
)

// TestNetCidrExpandRefusesLargeRanges pins #211: ranges above
// cidrExpandMaxHosts throw from the prefix length alone, before any
// enumeration, so they fail promptly instead of exhausting memory.
func (s *JSTestSuite) TestNetCidrExpandRefusesLargeRanges() {
	for _, cidr := range []string{"10.0.0.0/8", "10.0.0.0/15", "::/0", "2001:db8::/64", "2001:db8::/111"} {
		s.Run(cidr, func() {
			start := time.Now()
			_, err := s.runBuiltinScript(BuiltinNetGo, `m.cidrExpand("`+cidr+`")`)
			s.Require().Error(err)
			s.Contains(err.Error(), "more than the limit of 65536")
			s.Less(time.Since(start), time.Second, "refusal must not enumerate the range")
		})
	}
}

// TestNetCidrExpandSmallRanges checks ranges at and under the limit still
// enumerate every address, network and broadcast included.
func (s *JSTestSuite) TestNetCidrExpandSmallRanges() {
	cases := []struct {
		expr string
		want string
	}{
		{`m.cidrExpand("192.168.1.0/30").join(",")`, "192.168.1.0,192.168.1.1,192.168.1.2,192.168.1.3"},
		{`m.cidrExpand("192.168.1.7/32").join(",")`, "192.168.1.7"},
		{`m.cidrExpand("2001:db8::/127").join(",")`, "2001:db8::,2001:db8::1"},
		{`m.cidrExpand("10.0.0.0/16").length`, "65536"},
		{`m.cidrExpand("2001:db8::/112").length`, "65536"},
	}
	for _, tc := range cases {
		s.Run(tc.expr, func() {
			got, err := s.runBuiltinScript(BuiltinNetGo, tc.expr)
			s.Require().NoError(err)
			s.Equal(tc.want, got.String())
		})
	}
}
