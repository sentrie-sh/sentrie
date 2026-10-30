// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/xerr"
)

// validateSources parses each (file, source) pair into one index and validates it.
func (s *IndexTestSuite) validateSources(files ...[2]string) error {
	ctx := s.T().Context()
	idx := CreateIndex()
	for _, f := range files {
		prog, err := parser.NewParserFromString(f[1], f[0]).ParseProgram(ctx)
		s.Require().NoError(err)
		s.Require().NoError(idx.AddProgram(ctx, prog))
	}
	return idx.Validate(ctx)
}

// unknownFactErrors returns every ImportUnknownFactError wrapped in err.
func unknownFactErrors(err error) []xerr.ImportUnknownFactError {
	var out []xerr.ImportUnknownFactError
	var walk func(error)
	walk = func(e error) {
		if e == nil {
			return
		}
		if u, ok := e.(xerr.ImportUnknownFactError); ok {
			out = append(out, u)
			return
		}
		if j, ok := e.(interface{ Unwrap() []error }); ok {
			for _, c := range j.Unwrap() {
				walk(c)
			}
			return
		}
		walk(errors.Unwrap(e))
	}
	walk(err)
	return out
}

// TestImportWithUnknownFactIsRejected is the reproduction from #151: the typo
// `usr` names no fact of the target, so the index must refuse it instead of
// letting the target run on its default.
func (s *IndexTestSuite) TestImportWithUnknownFactIsRejected() {
	const src = `namespace t
policy q {
  fact user?: document as user default {"role": "guest"}
  fact zone?: string as zone
  rule ok = user.role == "admin"
  export decision of ok
}
policy p {
  rule typo  = import decision ok from t/q with usr  as {"role": "admin"}
  rule right = import decision ok from t/q with user as {"role": "admin"}
  export decision of typo
  export decision of right
}
`
	err := s.validateSources([2]string{"f.sentrie", src})
	s.Require().Error(err)
	s.ErrorIs(err, xerr.ErrIndex)
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 1, "only the typo'd clause is rejected, not the valid one")
	e := errs[0]
	s.Equal("f.sentrie", e.At.File)
	// The span starts at the `with` keyword of the typo'd clause, on source
	// line 9 (lines and columns are 1-based).
	s.Equal(9, e.At.From.Line)
	s.Equal(44, e.At.From.Column)
	s.True(strings.HasPrefix(src[e.At.From.Offset:], "with usr  as"), "span must start at the with clause")
	s.Equal(`import binds unknown fact "usr": policy t/q accepts fact aliases: user, zone`, e.Message)
}

// TestImportWithUnknownFactListsDeclaredFactsSorted checks the diagnostic order
// does not depend on map iteration.
func (s *IndexTestSuite) TestImportWithUnknownFactListsDeclaredFactsSorted() {
	err := s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  fact zeta?: number as zeta
  fact alpha?: number as alpha
  fact mid?: number as mid
  rule ok = default true { yield true }
  export decision of ok
}
policy p {
  rule r = import decision ok from t/q with nope as 1
  export decision of r
}
`})
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 1)
	s.Contains(errs[0].Message, "accepts fact aliases: alpha, mid, zeta")
}

func (s *IndexTestSuite) TestImportWithOnTargetWithoutFactsIsRejected() {
	err := s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  rule ok = default true { yield true }
  export decision of ok
}
policy p {
  rule r = import decision ok from t/q with user as 1
  export decision of r
}
`})
	s.Require().Error(err)
	s.ErrorIs(err, xerr.ErrIndex)
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 1)
	s.Equal(`import binds unknown fact "user": policy t/q declares no facts`, errs[0].Message)
}

func (s *IndexTestSuite) TestImportWithKnownOptionalFactIsAccepted() {
	s.Require().NoError(s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  fact user?: document as user default {"role": "guest"}
  rule ok = user.role == "admin"
  export decision of ok
}
policy p {
  rule r = import decision ok from t/q with user as {"role": "admin"}
  export decision of r
}
`}))
}

func (s *IndexTestSuite) TestImportWithKnownRequiredFactIsAccepted() {
	s.Require().NoError(s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  fact level: number as level
  rule ok = default false { yield level > 1 }
  export decision of ok
}
policy p {
  rule r = import decision ok from t/q with level as 2
  export decision of r
}
`}))
}

// TestImportWithFactsMatchByAlias pins the key space: Policy.Facts is keyed by
// the fact alias, and the with name must match that key (the same one the
// runtime injects by). #147 may revisit this.
func (s *IndexTestSuite) TestImportWithFactsMatchByAlias() {
	src := `namespace t
policy q {
  fact subject?: string as who
  rule ok = default true { yield true }
  export decision of ok
}
policy p {
  rule viaAlias = import decision ok from t/q with %s as "x"
  export decision of viaAlias
}
`
	s.Require().NoError(s.validateSources([2]string{"f.sentrie", fmt.Sprintf(src, "who")}))
	err := s.validateSources([2]string{"f.sentrie", fmt.Sprintf(src, "subject")})
	s.Require().Error(err)
	s.Len(unknownFactErrors(err), 1)
}

// TestImportWithReportsOnlyTheBadClauses checks that a bad clause among good
// ones, in one import and across rules, is reported once each in source order,
// and the good ones are not.
func (s *IndexTestSuite) TestImportWithReportsOnlyTheBadClauses() {
	err := s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  fact a?: number as a
  fact b?: number as b
  rule ok = default true { yield true }
  export decision of ok
}
policy p {
  rule first = import decision ok from t/q with a as 1 with bad1 as 2 with b as 3
  rule second = import decision ok from t/q with a as 1
  rule third = import decision ok from t/q with b as 1 with bad2 as 2
  export decision of first
  export decision of second
  export decision of third
}
`})
	s.Require().Error(err)
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 2)
	s.Contains(errs[0].Message, `"bad1"`)
	s.Equal(9, errs[0].At.From.Line)
	s.Contains(errs[1].Message, `"bad2"`)
	s.Equal(11, errs[1].At.From.Line)
	s.Less(errs[0].At.From.Line, errs[1].At.From.Line)
}

// TestImportWithAcrossNamespacesChecksTheTargetsFacts resolves the target
// through its own namespace, not the importing policy's.
func (s *IndexTestSuite) TestImportWithAcrossNamespacesChecksTheTargetsFacts() {
	target := [2]string{"q.sentrie", `namespace com/other
policy q {
  fact level?: number as level
  rule ok = default true { yield level > 0 }
  export decision of ok
}
`}
	good := [2]string{"p.sentrie", `namespace com/ex
policy p {
  rule r = import decision ok from com/other/q with level as 2
  export decision of r
}
`}
	bad := [2]string{"p.sentrie", `namespace com/ex
policy p {
  rule r = import decision ok from com/other/q with lvl as 2
  export decision of r
}
`}
	s.Require().NoError(s.validateSources(target, good))

	err := s.validateSources(target, bad)
	s.Require().Error(err)
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 1)
	s.Equal("p.sentrie", errs[0].At.File)
	s.Equal(3, errs[0].At.From.Line)
	s.Equal(`import binds unknown fact "lvl": policy com/other/q accepts fact aliases: level`, errs[0].Message)
}

// TestImportWithByDeclaredNameOfAliasedFactHintsAlias: facts are bound by
// alias, so `with user` does not reach `fact user as u`. The message must say
// which names are accepted and point at the alias.
func (s *IndexTestSuite) TestImportWithByDeclaredNameOfAliasedFactHintsAlias() {
	err := s.validateSources([2]string{"f.sentrie", `namespace t
policy q {
  fact user?: document as u
  rule ok = default true { yield true }
  export decision of ok
}
policy p {
  rule r = import decision ok from t/q with user as {"role": "admin"}
  rule right = import decision ok from t/q with u as {"role": "admin"}
  export decision of r
  export decision of right
}
`})
	s.Require().Error(err)
	errs := unknownFactErrors(err)
	s.Require().Len(errs, 1, "only the declared-name clause is rejected")
	s.Equal(`import binds unknown fact "user": policy t/q accepts fact aliases: u; fact "user" is bound by its alias "u"`, errs[0].Message)
}
