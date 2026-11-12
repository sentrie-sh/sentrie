// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"

	"github.com/sentrie-sh/sentrie/xerr"
)

const checksHeader = "namespace com/ex\npolicy p {\n"
const checksFooter = "}\n"

// checksSource wraps body (policy members) in a namespace and policy.
func checksSource(body string) string {
	return checksHeader + body + checksFooter
}

// TestValidateReportsSelfReference is #155: a rule or let whose own value
// reads itself is a validation error, where it used to pass until exec.
func (s *IndexTestSuite) TestValidateReportsSelfReference() {
	cases := map[string]struct {
		body string
		msg  string
	}{
		"rule reads itself": {
			body: "  rule r = r\n  export decision of r\n",
			msg:  "rule 'com/ex/p/r' references itself",
		},
		"rule reads itself through a block": {
			body: "  rule r = { let a = 1\n yield r + a }\n  export decision of r\n",
			msg:  "rule 'com/ex/p/r' references itself",
		},
		"rule reads itself in its when": {
			body: "  rule r = default false when r { yield true }\n  export decision of r\n",
			msg:  "rule 'com/ex/p/r' references itself",
		},
		"rule reads itself in its default": {
			body: "  rule r = default r { yield true }\n  export decision of r\n",
			msg:  "rule 'com/ex/p/r' references itself",
		},
		"policy let reads itself": {
			body: "  let x = x + 1\n  rule r = x > 0\n  export decision of r\n",
			msg:  "let 'x' references itself",
		},
		"block let reads itself": {
			body: "  rule r = { let a = a\n yield a }\n  export decision of r\n",
			msg:  "let 'a' references itself",
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			err := s.parseAndValidate(checksSource(tc.body), "c.sentrie")
			s.Require().Error(err)
			s.True(errors.Is(err, xerr.ErrIndex), err.Error())
			s.Contains(err.Error(), tc.msg)
			s.Contains(err.Error(), "c.sentrie:")
		})
	}
}

// TestValidateAttachmentReadingItsRuleIsSelfReference: execRule evaluates an
// export's attachments with the rule still on the reference stack, so
// `attach note as r` on `export decision of r` fails at exec with "references
// itself" and is a validation error. An attachment reading another rule is
// not.
func (s *IndexTestSuite) TestValidateAttachmentReadingItsRuleIsSelfReference() {
	err := s.parseAndValidate(checksSource("  rule r = true\n  export decision of r attach note as r\n"), "c.sentrie")
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrIndex), err.Error())
	s.Contains(err.Error(), "rule 'com/ex/p/r' references itself")
	s.Contains(err.Error(), "c.sentrie:")

	s.NoError(s.parseAndValidate(checksSource("  rule q = 1\n  rule r = true\n  export decision of r attach note as q\n"), "c.sentrie"))
}

// TestValidateSelfReferenceNoFalsePositives pins the fail-safe: names bound
// by a parameter, block let or fact, and recursion inside a lambda body, are
// never reported as a self-reference.
func (s *IndexTestSuite) TestValidateSelfReferenceNoFalsePositives() {
	cases := map[string]string{
		"block let shadows the rule":     "  rule x = { let x = 1\n yield x }\n  export decision of x\n",
		"lambda param shadows the rule":  "  rule x = count(filter([1], (x) => { yield x })) > 0\n  export decision of x\n",
		"block let over a fact":          "  fact user: document as user\n  rule r = { let user = user.name\n yield user == \"a\" }\n  export decision of r\n",
		"rule read inside a lambda body": "  rule r = count(filter([1], (v) => { yield r })) > 0\n  export decision of r\n",
		"let lambda call":                "  let ident = (n) => { yield n }\n  rule r = ident(1) == 1\n  export decision of r\n",
	}
	for name, body := range cases {
		s.Run(name, func() {
			s.NoError(s.parseAndValidate(checksSource(body), "c.sentrie"))
		})
	}
}

// TestValidateShadowedLetSelfReferenceIsSkipped is the PR review case: a let
// reading its own name while an outer let, rule, fact or lambda parameter
// binds that name too may read the outer value at runtime (evalIdent checks
// cached locals across ancestor contexts before lets), so it is skipped as
// ambiguous rather than reported.
func (s *IndexTestSuite) TestValidateShadowedLetSelfReferenceIsSkipped() {
	cases := map[string]string{
		"policy let":       "  let x = 1\n  rule r = (x + { let x = x + 1\n yield x }) == 2\n  export decision of r\n",
		"rule":             "  rule x = 1\n  rule r = (x + { let x = x + 1\n yield x }) == 2\n  export decision of r\n",
		"outer block let":  "  rule r = { let x = 1\n yield x + { let x = x + 1\n yield x } } == 2\n  export decision of r\n",
		"fact alias":       "  fact x?: number default 1\n  rule r = { let x = x + 1\n yield x } == 1\n  export decision of r\n",
		"lambda parameter": "  rule r = count(filter([1], (x) => { let x = x + 1\n yield x == 1 })) == 1\n  export decision of r\n",
	}
	for name, body := range cases {
		s.Run(name, func() {
			s.NoError(s.parseAndValidate(checksSource(body), "c.sentrie"))
		})
	}

	// A nested let reading its own name inside a lambda body, with an outer
	// policy let of that name, is skipped too.
	s.NoError(s.parseAndValidate(checksSource("  let x = 1\n  rule r = count(filter([1], (v) => { let x = x + v\n yield x == 2 })) == 1\n  export decision of r\n"), "c.sentrie"))

	// With no outer x of any kind the self-reference is still reported.
	err := s.parseAndValidate(checksSource("  rule r = { let x = x + 1\n yield x } == 2\n  export decision of r\n"), "c.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "let 'x' references itself")
}

// TestValidateReportsUndefinedIdentifier is #155: a name that nothing binds is
// a validation error at validate time, with a hint where it is a near miss.
func (s *IndexTestSuite) TestValidateReportsUndefinedIdentifier() {
	cases := map[string]struct {
		body string
		msg  string
	}{
		"plain undefined in a rule": {
			body: "  rule r = { yield undefined_thing }\n  export decision of r\n",
			msg:  "identifier 'undefined_thing' not found at c.sentrie:",
		},
		"undefined in a let": {
			body: "  let x = nothing_here\n  rule r = x\n  export decision of r\n",
			msg:  "identifier 'nothing_here' not found",
		},
		"undefined in a when": {
			body: "  rule r = default false when ghost { yield true }\n  export decision of r\n",
			msg:  "identifier 'ghost' not found",
		},
		"undefined in an attachment": {
			body: "  rule r = true\n  export decision of r attach note as ghost\n",
			msg:  "identifier 'ghost' not found",
		},
		"undefined in a lambda body": {
			body: "  rule r = count(filter([1], (v) => { yield ghost })) > 0\n  export decision of r\n",
			msg:  "identifier 'ghost' not found",
		},
		"module alias used as a value": {
			body: "  use { sqrt } from @sentrie/math as math\n  rule r = { yield math }\n  export decision of r\n",
			msg:  "'math' is a module alias, not a value",
		},
		"module name imported by use used as a value": {
			body: "  use { sqrt } from @sentrie/math as m\n  rule r = { yield sqrt }\n  export decision of r\n",
			msg:  "imported from module alias 'm' and must be called as m.sqrt(...)",
		},
		"builtin used as a value": {
			body: "  rule r = { yield count }\n  export decision of r\n",
			msg:  "'count' is a builtin, not a value; call it as count(...)",
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			err := s.parseAndValidate(checksSource(tc.body), "c.sentrie")
			s.Require().Error(err)
			s.True(errors.Is(err, xerr.ErrIndex), err.Error())
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// TestValidateIdentifiersNoFalsePositives pins that every kind of binding
// still resolves: params, block lets, facts, lets, rules, derives and a use
// alias used as a callee.
func (s *IndexTestSuite) TestValidateIdentifiersNoFalsePositives() {
	cases := map[string]string{
		"lambda param":                          "  rule r = count(filter([1], (v) => { yield v + 1 })) > 0\n  export decision of r\n",
		"block let":                             "  rule r = { let a = 1\n yield a > 0 }\n  export decision of r\n",
		"fact by alias":                         "  fact user: document as u\n  rule r = u.name == \"a\"\n  export decision of r\n",
		"policy let and rule":                   "  let x = 1\n  rule a = x > 0\n  rule r = a\n  export decision of r\n",
		"derive":                                "  derive d = (n: number): number => { yield n }\n  rule r = d(1) == 1\n  export decision of r\n",
		"use alias callee":                      "  use { sqrt } from @sentrie/math as math\n  rule r = math.sqrt(4) == 2\n  export decision of r\n",
		"dynamically scoped let read in lambda": "  let y = x + 1\n  rule r = count(filter([1], (x) => { yield y })) > 0\n  export decision of r\n",
	}
	for name, body := range cases {
		s.Run(name, func() {
			s.NoError(s.parseAndValidate(checksSource(body), "c.sentrie"))
		})
	}
}

// TestValidateStillAcceptsStaticLiteralTypeError pins the deferred part of
// #155: `1 + true` is not a static error yet, so the follow-up issue shows the
// gap. When static literal type checks land, this test is flipped.
func (s *IndexTestSuite) TestValidateStillAcceptsStaticLiteralTypeError() {
	s.NoError(s.parseAndValidate(checksSource("  rule r = { yield 1 + true }\n  export decision of r\n"), "c.sentrie"))
}

// TestValidateMutualLetRecursionOutcome pins two policy lets that read each
// other with no shadowing.
func (s *IndexTestSuite) TestValidateMutualLetRecursionOutcome() {
	err := s.parseAndValidate(checksSource("  let a = b\n  let b = a\n  rule r = a\n  export decision of r\n"), "c.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "infinite recursion")
}
