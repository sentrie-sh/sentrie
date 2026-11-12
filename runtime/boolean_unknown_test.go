// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

const booleanUnknownSrc = `namespace com/ex

policy pol {
  fact flag: boolean
  fact tri: trinary
  fact maybe?: boolean

  let b: boolean = flag
  let t: trinary = tri

  rule boolFact = { yield flag }
  rule boolLetFromFact: boolean = flag
  rule triLetFromFact: trinary = tri
  rule unknownBoolRule: boolean = unknown
  rule unknownTriRule: trinary = unknown
  rule trueBoolRule: boolean = true
  rule falseBoolRule: boolean = false
  rule castBool = { yield (tri as boolean) }
  rule castTri = { yield (tri as trinary) }

  -- PROVISIONAL: skipped base, no true contribution, so the check is skipped
  rule skipped: boolean = when false { yield true }
  contribute skipped { yield false }

  -- a contribution yielding unknown to a boolean rule errors
  rule badContribution: boolean = false
  contribute badContribution { yield tri }

  -- a trinary rule still accepts an unknown contribution
  rule triContribution: trinary = false
  contribute triContribution { yield tri }

  -- a true contribution rescues a skipped base
  rule rescued: boolean = when false { yield false }
  contribute rescued { yield true }

  export decision of boolFact
  export decision of boolLetFromFact
  export decision of triLetFromFact
  export decision of unknownBoolRule
  export decision of unknownTriRule
  export decision of trueBoolRule
  export decision of falseBoolRule
  export decision of castBool
  export decision of castTri
  export decision of skipped
  export decision of badContribution
  export decision of triContribution
  export decision of rescued
}
`

func (s *RuntimeTestSuite) booleanUnknownExec() *executorImpl {
	_, exec := s.mustBuildDeriveExecutor(s.T().Context(), deriveTestProgram{name: "boolean_unknown.sentrie", src: booleanUnknownSrc})
	return exec
}

// TestBooleanRejectsUnknown: `boolean` accepts true and false and rejects
// unknown on facts and rule return types, while `trinary` still accepts it
// (#144).
func (s *RuntimeTestSuite) TestBooleanRejectsUnknown() {
	exec := s.booleanUnknownExec()
	ctx := s.T().Context()
	facts := func(flag, tri any) map[string]any { return map[string]any{"flag": flag, "tri": tri} }

	for name, c := range map[string]struct {
		rule string
		in   map[string]any
		want trinary.Value
	}{
		"boolean fact true":            {"boolFact", facts(true, true), trinary.True},
		"boolean fact false":           {"boolFact", facts(false, true), trinary.False},
		"trinary fact unknown":         {"triLetFromFact", facts(true, trinary.Unknown), trinary.Unknown},
		"boolean rule true":            {"trueBoolRule", facts(true, true), trinary.True},
		"boolean rule false":           {"falseBoolRule", facts(true, true), trinary.False},
		"trinary rule unknown":         {"unknownTriRule", facts(true, true), trinary.Unknown},
		"cast unknown to trinary":      {"castTri", facts(true, trinary.Unknown), trinary.Unknown},
		"cast true to boolean":         {"castBool", facts(true, true), trinary.True},
		"trinary contribution unknown": {"triContribution", facts(true, trinary.Unknown), trinary.Unknown},
		"true contribution rescues":    {"rescued", facts(true, true), trinary.True},
	} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", c.rule, c.in)
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State)
		})
	}

	for name, c := range map[string]struct {
		rule string
		in   map[string]any
		msg  string
	}{
		"boolean fact unknown":            {"boolFact", facts(trinary.Unknown, true), "not a boolean"},
		"boolean rule from unknown fact":  {"boolLetFromFact", facts(trinary.Unknown, true), "not a boolean"},
		"boolean rule returning unknown":  {"unknownBoolRule", facts(true, true), "use trinary to allow unknown"},
		"cast unknown to boolean":         {"castBool", facts(true, trinary.Unknown), "not a boolean"},
		"contribution unknown to boolean": {"badContribution", facts(true, trinary.Unknown), "not a boolean"},
	} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", c.rule, c.in)
			s.Require().Error(err)
			s.Contains(err.Error(), c.msg)
			s.Equal(trinary.Unknown, out.Decision.State)
		})
	}
}

// TestContributedBooleanSkippedBase: a boolean rule whose base is skipped and
// that has no true contribution reports unknown instead of failing its own
// type (PROVISIONAL, #144).
func (s *RuntimeTestSuite) TestContributedBooleanSkippedBase() {
	out, err := s.booleanUnknownExec().ExecRule(s.T().Context(), "com/ex", "pol", "skipped",
		map[string]any{"flag": true, "tri": true})
	s.Require().NoError(err)
	s.Equal(trinary.Unknown, out.Decision.State)
}
