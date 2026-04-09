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

import "github.com/sentrie-sh/sentrie/constraints"

func (s *ConstraintsTestSuite) TestEmptyCheckerMapsAreInitialized() {
	for name, m := range map[string]map[string]constraints.ConstraintDefinition{
		"map":      constraints.DictContraintCheckers,
		"record":   constraints.RecordContraintCheckers,
		"shape":    constraints.ShapeContraintCheckers,
		"document": constraints.DocumentContraintCheckers,
	} {
		s.NotNil(m, "%s: map is nil", name)
		s.Empty(m, "%s: expected empty map", name)
	}
}
