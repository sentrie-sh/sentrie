// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/sentrie-sh/sentrie/runtime"
)

func (s *CmdTestSuite) TestParseEvalTimeout() {
	d, err := parseEvalTimeout(runtime.DefaultEvalTimeout.String())
	s.Require().NoError(err)
	s.Equal(runtime.DefaultEvalTimeout, d)

	d, err = parseEvalTimeout("250ms")
	s.Require().NoError(err)
	s.Equal(250*time.Millisecond, d)

	d, err = parseEvalTimeout("0")
	s.Require().NoError(err)
	s.Zero(d, "0 disables the deadline")

	_, err = parseEvalTimeout("-1s")
	s.Require().ErrorContains(err, "must not be negative")
	_, err = parseEvalTimeout("soon")
	s.Require().ErrorContains(err, "invalid --eval-timeout")
}

// TestParseEvalMaxTimeout walks #243's flag table. The deadlines a request
// then gets under each row are covered by api's TestRequestDeadlineCombinations.
func (s *CmdTestSuite) TestParseEvalMaxTimeout() {
	const d = 10 * time.Second
	cases := []struct {
		name        string
		evalTimeout time.Duration
		max         string
		want        time.Duration
		errPart     string
	}{
		{name: "D, unset: requests may only shorten", evalTimeout: d, max: "", want: 0},
		{name: "D, M above D", evalTimeout: d, max: "1m", want: time.Minute},
		{name: "D, M equal to D", evalTimeout: d, max: "10s", want: d},
		{name: "D, M below D is a startup error", evalTimeout: d, max: "5s", errPart: "must not be below --eval-timeout 10s"},
		{name: "0, unset: unbounded", evalTimeout: 0, max: "", want: 0},
		{name: "0, M: the default becomes M", evalTimeout: 0, max: "30s", want: 30 * time.Second},
		{name: "explicit zero", evalTimeout: d, max: "0", errPart: "must be positive"},
		{name: "negative", evalTimeout: d, max: "-1m", errPart: "must be positive"},
		{name: "garbage", evalTimeout: d, max: "later", errPart: "invalid --eval-max-timeout"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := parseEvalMaxTimeout(tc.max, tc.evalTimeout)
			if tc.errPart != "" {
				s.Require().ErrorContains(err, tc.errPart)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}
