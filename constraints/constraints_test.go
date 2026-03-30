// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
)

func runChecker(t *testing.T, c constraints.ConstraintDefinition, val box.Value, args []box.Value, wantErr bool) {
	t.Helper()
	err := c.Checker(context.Background(), (*index.Policy)(nil), val, args)
	if wantErr && err == nil {
		t.Fatal("expected error, got nil")
	}
	if !wantErr && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmptyCheckerMapsAreInitialized(t *testing.T) {
	for name, m := range map[string]map[string]constraints.ConstraintDefinition{
		"map":      constraints.MapContraintCheckers,
		"record":   constraints.RecordContraintCheckers,
		"shape":    constraints.ShapeContraintCheckers,
		"document": constraints.DocumentContraintCheckers,
	} {
		if m == nil {
			t.Fatalf("%s: map is nil", name)
		}
		if len(m) != 0 {
			t.Fatalf("%s: expected empty map", name)
		}
	}
}
