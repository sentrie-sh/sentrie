// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package capabilities

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CapabilitiesTestSuite struct {
	suite.Suite
}

func TestCapabilitiesTestSuite(t *testing.T) {
	suite.Run(t, new(CapabilitiesTestSuite))
}
