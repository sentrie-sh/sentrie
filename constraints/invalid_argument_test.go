// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"errors"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestInvalidArgumentIsDistinguishable checks that a failure caused by the
// constraint's own arguments matches ErrInvalidArgument, keeps its message,
// and that a value which simply fails the constraint does not match it.
func (s *ConstraintsTestSuite) TestInvalidArgumentIsDistinguishable() {
	ctx := s.T().Context()
	str := constraints.StringContraintCheckers
	num := constraints.NumberContraintCheckers
	tri := constraints.TrinaryConstraintCheckers

	invalid := map[string]error{
		"regexp invalid pattern": str["regexp"].Checker(ctx, box.String("ab"), []box.Value{box.String("[")}),
		"minlength wrong kind":   str["minlength"].Checker(ctx, box.String("ab"), []box.Value{box.String("q")}),
		"minlength arity":        str["minlength"].Checker(ctx, box.String("ab"), nil),
		"one_of wrong kind":      str["one_of"].Checker(ctx, box.String("ab"), []box.Value{box.Number(1)}),
		"in wrong set item":      num["in"].Checker(ctx, box.Number(1), []box.Value{box.List([]box.Value{box.String("x")})}),
		"multiple_of zero":       num["multiple_of"].Checker(ctx, box.Number(4), []box.Value{box.Number(0)}),
		"range wrong kind":       num["range"].Checker(ctx, box.Number(4), []box.Value{box.Number(0), box.String("x")}),
		"range min above max":    num["range"].Checker(ctx, box.Number(4), []box.Value{box.Number(5), box.Number(1)}),
		"trinary eq wrong kind":  tri["eq"].Checker(ctx, box.Trinary(trinary.True), []box.Value{box.Number(1)}),
	}
	for name, err := range invalid {
		s.Run(name, func() {
			s.Require().Error(err)
			s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
			s.NotContains(err.Error(), constraints.ErrInvalidArgument.Error())
		})
	}

	mismatch := map[string]error{
		"regexp no match":     str["regexp"].Checker(ctx, box.String("ab"), []box.Value{box.String("^z")}),
		"minlength too short": str["minlength"].Checker(ctx, box.String("a"), []box.Value{box.Number(2)}),
		"min below":           num["min"].Checker(ctx, box.Number(1), []box.Value{box.Number(2)}),
		"value wrong kind":    str["minlength"].Checker(ctx, box.Number(1), []box.Value{box.Number(2)}),
	}
	for name, err := range mismatch {
		s.Run(name, func() {
			s.Require().Error(err)
			s.False(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
		})
	}
}
