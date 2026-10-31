// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"sync"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

const contributeGapsSrc = `namespace com/ex

policy src {
  fact user: document as user

  rule deny: list[string] = []
  contribute deny when user.disabled { yield "user disabled" }
  contribute deny { yield "always" }
  export decision of deny

  rule ok: boolean = false
  contribute ok when user.admin { yield true }
  export decision of ok
}

policy gaps {
  fact user: document as user

  -- a computed contribution that fails at runtime
  rule boom: list[string] = []
  contribute boom { yield "first" }
  contribute boom { yield 4 % user.zero }
  export decision of boom

  -- a failing guard
  rule boomwhen: list[string] = []
  contribute boomwhen when 4 % user.zero > 0 { yield "x" }
  export decision of boomwhen

  -- a null base of a nullable list rule starts the list empty
  rule nullbase: list[string]? = null
  contribute nullbase { yield "x" }
  export decision of nullbase

  -- a list of records takes a record literal
  rule pairs: list[record[string, number]] = []
  contribute pairs { yield ["a", 1] }
  export decision of pairs

  -- a typed rule whose own default has the wrong type
  rule baddefault: list[string] = default "nope" when false { yield [] }
  export decision of baddefault

  -- a contribution that reads another rule, declared before and after it
  rule before = "b"
  rule uses: list[string] = []
  contribute uses { yield before }
  contribute uses { yield after }
  rule after = "a"
  export decision of uses

  -- boolean OR over true, false and unknown, with a skipped base
  rule falseBase: boolean = false
  contribute falseBase { yield unknown }
  rule falseBaseTrue: boolean = false
  contribute falseBaseTrue { yield unknown }
  contribute falseBaseTrue { yield true }
  rule unknownBase: boolean = unknown
  contribute unknownBase { yield false }
  rule trueBase: boolean = true
  contribute trueBase { yield unknown }
  rule skippedBase: boolean = when user.never { yield true }
  contribute skippedBase { yield false }
  rule skippedBaseTrue: boolean = when user.never { yield true }
  contribute skippedBaseTrue { yield true }
  rule undefinedContrib: boolean = false
  contribute undefinedContrib { yield user.missing }
  rule trinaryRule: trinary = false
  contribute trinaryRule { yield unknown }
  export decision of falseBase
  export decision of falseBaseTrue
  export decision of unknownBase
  export decision of trueBase
  export decision of skippedBase
  export decision of skippedBaseTrue
  export decision of undefinedContrib
  export decision of trinaryRule

  -- a nullable element accepts a null contribution
  rule maybe: list[string?] = []
  contribute maybe { yield "a" }
  contribute maybe { yield user.nothing }
  contribute maybe { yield "b" }
  export decision of maybe

  -- a list constraint applies to the combined list, not to the skipped base
  rule nonempty: list[string]@not_empty() = when user.never { yield ["x"] }
  contribute nonempty when user.admin { yield "b" }
  export decision of nonempty
}

policy consumer {
  fact user: document as user
  rule deny = import decision deny from com/ex/src with user as user
  rule ok = import decision ok from com/ex/src with user as user
  rule viaRef = count(deny.value)
  export decision of deny
  export decision of ok
  export decision of viaRef
}
`

func (s *RuntimeTestSuite) contributeGapsExec() *executorImpl {
	_, exec := s.mustBuildDeriveExecutor(s.T().Context(), deriveTestProgram{name: "contribute_gaps.sentrie", src: contributeGapsSrc})
	return exec
}

func gapsFacts(extra map[string]any) map[string]any {
	user := map[string]any{"disabled": true, "admin": false, "never": false, "zero": float64(0), "nothing": nil}
	for k, v := range extra {
		user[k] = v
	}
	return map[string]any{"user": user}
}

func (s *RuntimeTestSuite) TestContributeOrderIsStableAcrossRepeatedAndConcurrentRuns() {
	exec := s.contributeGapsExec()
	ctx := s.T().Context()
	want := []any{"user disabled", "always"}

	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for range 16 {
		wg.Go(func() {
			for range 8 {
				out, err := exec.ExecRule(ctx, "com/ex", "src", "deny", gapsFacts(nil))
				if err != nil {
					errs <- err.Error()
					return
				}
				got := box.ToBoundaryAny(out.Decision.Value)
				if len(got.([]any)) != 2 || got.([]any)[0] != want[0] || got.([]any)[1] != want[1] {
					errs <- "unexpected order"
					return
				}
			}
		})
	}
	// exported rules of the policy are evaluated concurrently by ExecPolicy
	for range 8 {
		wg.Go(func() {
			outs, err := exec.ExecPolicy(ctx, "com/ex", "src", gapsFacts(nil))
			if err != nil {
				errs <- err.Error()
				return
			}
			for _, o := range outs {
				if o.RuleName == "deny" && len(box.ToBoundaryAny(o.Decision.Value).([]any)) != 2 {
					errs <- "unexpected policy output"
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		s.Fail(e)
	}
}

func (s *RuntimeTestSuite) TestContributeRuntimeErrorFailsTheRuleAndIsReported() {
	exec := s.contributeGapsExec()
	ctx := s.T().Context()

	for _, rule := range []string{"boom", "boomwhen"} {
		out, err := exec.ExecRule(ctx, "com/ex", "gaps", rule, gapsFacts(nil))
		s.Require().Error(err, rule)
		s.Contains(err.Error(), "divide by zero", rule)
		// a failed contribution reports Unknown, never the base or a partial list
		s.Require().NotNil(out, rule)
		s.Equal(trinary.Unknown, out.Decision.State, rule)
		s.Equal(box.Trinary(trinary.Unknown), out.Decision.Value, rule)
	}

	// the failing rule is reported by ExecPolicy, the other rules still run
	outs, err := exec.ExecPolicy(ctx, "com/ex", "gaps", gapsFacts(nil))
	s.Require().Error(err)
	s.Contains(err.Error(), "divide by zero")
	names := map[string]bool{}
	for _, o := range outs {
		names[o.RuleName] = true
	}
	s.False(names["boom"])
	s.False(names["boomwhen"])
	s.True(names["uses"])
}

func (s *RuntimeTestSuite) TestContributeTypeErrorsReportUnknown() {
	exec := s.contributeGapsExec()
	for _, rule := range []string{"baddefault"} {
		out, err := exec.ExecRule(s.T().Context(), "com/ex", "gaps", rule, gapsFacts(nil))
		s.Require().Error(err, rule)
		s.Require().NotNil(out, rule)
		s.Equal(trinary.Unknown, out.Decision.State, rule)
		s.Equal(box.Trinary(trinary.Unknown), out.Decision.Value, rule)
	}
}

func (s *RuntimeTestSuite) TestContributeNullBaseStartsEmpty() {
	out, err := s.contributeGapsExec().ExecRule(s.T().Context(), "com/ex", "gaps", "nullbase", gapsFacts(nil))
	s.Require().NoError(err)
	s.Equal([]any{"x"}, box.ToBoundaryAny(out.Decision.Value))
}

func (s *RuntimeTestSuite) TestContributeListOfRecords() {
	out, err := s.contributeGapsExec().ExecRule(s.T().Context(), "com/ex", "gaps", "pairs", gapsFacts(nil))
	s.Require().NoError(err)
	s.Equal([]any{[]any{"a", float64(1)}}, box.ToBoundaryAny(out.Decision.Value))
}

func (s *RuntimeTestSuite) TestContributionReadsRulesDeclaredBeforeAndAfterIt() {
	out, err := s.contributeGapsExec().ExecRule(s.T().Context(), "com/ex", "gaps", "uses", gapsFacts(nil))
	s.Require().NoError(err)
	s.Equal([]any{"b", "a"}, box.ToBoundaryAny(out.Decision.Value))
}

func (s *RuntimeTestSuite) TestContributeBooleanThreeValuedOr() {
	exec := s.contributeGapsExec()
	ctx := s.T().Context()
	for rule, want := range map[string]trinary.Value{
		"falseBase":        trinary.Unknown, // false OR unknown
		"falseBaseTrue":    trinary.True,    // unknown OR true wins
		"unknownBase":      trinary.Unknown, // unknown OR false
		"trueBase":         trinary.True,    // true OR unknown
		"skippedBase":      trinary.Unknown, // a skipped base is unknown, OR false
		"skippedBaseTrue":  trinary.True,
		"undefinedContrib": trinary.False, // an undefined contribution is skipped
		"trinaryRule":      trinary.Unknown,
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "gaps", rule, gapsFacts(nil))
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}
}

func (s *RuntimeTestSuite) TestContributeNullableElementAcceptsNull() {
	out, err := s.contributeGapsExec().ExecRule(s.T().Context(), "com/ex", "gaps", "maybe", gapsFacts(nil))
	s.Require().NoError(err)
	// an undefined `user.nothing`-style yield is skipped, a null one is kept
	got, ok := box.ToBoundaryAny(out.Decision.Value).([]any)
	s.Require().True(ok)
	s.Equal([]any{"a", nil, "b"}, got)
}

func (s *RuntimeTestSuite) TestContributeListConstraintAppliesToCombinedList() {
	exec := s.contributeGapsExec()
	ctx := s.T().Context()

	_, err := exec.ExecRule(ctx, "com/ex", "gaps", "nonempty", gapsFacts(nil))
	s.Require().Error(err)

	out, err := exec.ExecRule(ctx, "com/ex", "gaps", "nonempty", gapsFacts(map[string]any{"admin": true}))
	s.Require().NoError(err)
	s.Equal([]any{"b"}, box.ToBoundaryAny(out.Decision.Value))
}

func (s *RuntimeTestSuite) TestContributedValueIsSeenByAnImportingPolicy() {
	exec := s.contributeGapsExec()
	ctx := s.T().Context()

	out, err := exec.ExecRule(ctx, "com/ex", "consumer", "deny", gapsFacts(nil))
	s.Require().NoError(err)
	s.Equal([]any{"user disabled", "always"}, box.ToBoundaryAny(out.Decision.Value))

	out, err = exec.ExecRule(ctx, "com/ex", "consumer", "deny", gapsFacts(map[string]any{"disabled": false}))
	s.Require().NoError(err)
	s.Equal([]any{"always"}, box.ToBoundaryAny(out.Decision.Value))

	out, err = exec.ExecRule(ctx, "com/ex", "consumer", "ok", gapsFacts(map[string]any{"admin": true}))
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
	out, err = exec.ExecRule(ctx, "com/ex", "consumer", "ok", gapsFacts(nil))
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
}
