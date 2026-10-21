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
