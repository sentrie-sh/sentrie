// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

func (s *ConstraintsTestSuite) runChecker(c constraints.ConstraintDefinition, val box.Value, args []box.Value, wantErr bool) {
	s.T().Helper()
	err := c.Checker(s.T().Context(), val, args)
	if wantErr {
		s.Error(err, "expected error, got nil")
	} else {
		s.NoError(err)
	}
}
