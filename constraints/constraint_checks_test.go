// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"errors"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

// TestStringInAndNotIn covers @in / @not_in on strings (#146): the value must
// equal one of the listed strings, each given as its own argument.
func (s *ConstraintsTestSuite) TestStringInAndNotIn() {
	in := constraints.StringContraintCheckers["in"]
	notIn := constraints.StringContraintCheckers["not_in"]
	abc := []box.Value{box.String("a"), box.String("b"), box.String("c")}
	s.runChecker(in, box.String("b"), abc, false)
	s.runChecker(in, box.String("z"), abc, true)
	s.runChecker(in, box.String("z"), []box.Value{box.String("z")}, false)
	s.runChecker(notIn, box.String("z"), abc, false)
	s.runChecker(notIn, box.String("b"), abc, true)
	s.runChecker(in, box.Number(1), abc, true)
	s.runChecker(notIn, box.Number(1), abc, true)

	list := []box.Value{box.List(abc)}
	for name, c := range map[string]constraints.ConstraintDefinition{"in": in, "not_in": notIn} {
		err := c.Checker(s.T().Context(), box.String("a"), list)
		s.Require().Error(err, name)
		s.True(errors.Is(err, constraints.ErrInvalidArgument), name)
		s.Contains(err.Error(), "takes each value as its own argument", name)
	}
}

// TestNumberInListFormIsInvalidArgument pins the message for the old list form.
func (s *ConstraintsTestSuite) TestNumberInListFormIsInvalidArgument() {
	list := []box.Value{box.List([]box.Value{box.Number(1), box.Number(2)})}
	for _, name := range []string{"in", "not_in"} {
		err := constraints.NumberContraintCheckers[name].Checker(s.T().Context(), box.Number(1), list)
		s.Require().Error(err, name)
		s.True(errors.Is(err, constraints.ErrInvalidArgument), name)
		s.Contains(err.Error(), name+" constraint takes each value as its own argument, as in @"+name+"(1, 2), not a list", name)
	}
}

// TestStringURLSchemesAndHosts covers @url (#146).
func (s *ConstraintsTestSuite) TestStringURLSchemesAndHosts() {
	c := constraints.StringContraintCheckers["url"]
	s.Run("default schemes", func() {
		for _, ok := range []string{
			"http://example.com", "https://example.com/a?b=c#d", "HTTP://EXAMPLE.COM",
			"http://localhost", "http://localhost:8080/x", "https://example.com:443",
			"http://127.0.0.1", "http://127.0.0.1:3000/p", "http://[::1]", "http://[::1]:8080/p",
			"http://[2001:db8::1]/x",
		} {
			s.runChecker(c, box.String(ok), nil, false)
		}
		for _, bad := range []string{
			"localhost:8080", "http://:80", "/path", "example.com", "http://", "https:///path",
			"ftp://example.com", "", "mailto:a@b.co",
		} {
			s.runChecker(c, box.String(bad), nil, true)
		}
	})
	s.Run("schemes replace the default", func() {
		ftp := []box.Value{box.String("ftp")}
		s.runChecker(c, box.String("ftp://example.com/f"), ftp, false)
		s.runChecker(c, box.String("http://example.com"), ftp, true)
		s.runChecker(c, box.String("https://example.com"), ftp, true)
		three := []box.Value{box.String("http"), box.String("https"), box.String("ftp")}
		for _, u := range []string{"http://a.co", "https://a.co", "ftp://a.co"} {
			s.runChecker(c, box.String(u), three, false)
		}
		s.runChecker(c, box.String("s3://bucket"), three, true)
		s.runChecker(c, box.String("ftp://example.com"), []box.Value{box.String("FTP")}, false)
		s.runChecker(c, box.String("FTP://example.com"), ftp, false)
	})
	s.Run("invalid scheme argument", func() {
		for _, arg := range []box.Value{box.String("ftp://"), box.String(""), box.String("1ftp"), box.String("a b"), box.Number(1)} {
			err := c.Checker(s.T().Context(), box.String("ftp://example.com"), []box.Value{arg})
			s.Require().Error(err)
			s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
		}
	})
	s.Run("non-string value", func() {
		s.runChecker(c, box.Number(1), nil, true)
	})
}

// TestLengthCountsCodePoints checks @length family counts runes, not bytes (#146).
func (s *ConstraintsTestSuite) TestLengthCountsCodePoints() {
	one := []box.Value{box.Number(1)}
	two := []box.Value{box.Number(2)}
	pre := "é"  // precomposed e-acute: 2 bytes, 1 code point
	dec := "é" // decomposed: 2 code points
	emoji := "\U0001F600"
	s.runChecker(constraints.StringContraintCheckers["length"], box.String(pre), one, false)
	s.runChecker(constraints.StringContraintCheckers["length"], box.String(pre), two, true)
	s.runChecker(constraints.StringContraintCheckers["length"], box.String(dec), two, false)
	s.runChecker(constraints.StringContraintCheckers["length"], box.String(emoji), one, false)
	s.runChecker(constraints.StringContraintCheckers["length"], box.String("日本"), two, false)
	s.runChecker(constraints.StringContraintCheckers["minlength"], box.String(pre), two, true)
	s.runChecker(constraints.StringContraintCheckers["minlength"], box.String(pre), one, false)
	s.runChecker(constraints.StringContraintCheckers["maxlength"], box.String(pre), one, false)
	s.runChecker(constraints.StringContraintCheckers["maxlength"], box.String(emoji), one, false)
	s.runChecker(constraints.StringContraintCheckers["maxlength"], box.String(dec), one, true)
}

// TestMultipleOfNegativeValues covers @multiple_of with negative values (#146).
func (s *ConstraintsTestSuite) TestMultipleOfNegativeValues() {
	c := constraints.NumberContraintCheckers["multiple_of"]
	d := func(n float64) []box.Value { return []box.Value{box.Number(n)} }
	s.runChecker(c, box.Number(-6), d(3), false)
	s.runChecker(c, box.Number(-6), d(-3), false)
	s.runChecker(c, box.Number(6), d(-3), false)
	s.runChecker(c, box.Number(-7), d(3), true)
	s.runChecker(c, box.Number(-3.5), d(2), true)
	s.runChecker(c, box.Number(-0.5), d(0.25), false)
	s.runChecker(c, box.Number(-7), d(-3), true)
}
