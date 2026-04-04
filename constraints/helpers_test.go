// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package constraints_test

import (
	"context"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
)

func (s *ConstraintsTestSuite) runChecker(c constraints.ConstraintDefinition, val box.Value, args []box.Value, wantErr bool) {
	s.T().Helper()
	err := c.Checker(context.Background(), (*index.Policy)(nil), val, args)
	if wantErr {
		s.Error(err, "expected error, got nil")
	} else {
		s.NoError(err)
	}
}
