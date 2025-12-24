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

package api

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
)

func TestPathParsing(t *testing.T) {
	// "/decision/sh/sentra/auth/v1/user/allow"
	t.Run("valid deep path", func(t *testing.T) {
		path := "/decision/sh/sentra/auth/v1/user/allow"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)

		if len(pathParts) < 3 {
			t.Fatalf("Expected valid parsing for %s", path)
		}

		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)

		if namespace != "sh/sentra/auth/v1" {
			t.Errorf("Expected namespace %s, got %s", "sh/sentra/auth/v1", namespace)
		}
		if policy != "user" {
			t.Errorf("Expected policy %s, got %s", "user", policy)
		}
		if rule != "allow" {
			t.Errorf("Expected rule %s, got %s", "allow", rule)
		}
	})

	// "/decision/org/department/team/policy/rule"
	t.Run("valid org/department/team path", func(t *testing.T) {
		path := "/decision/org/department/team/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)

		if len(pathParts) < 3 {
			t.Fatalf("Expected valid parsing for %s", path)
		}

		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)

		if namespace != "org/department/team" {
			t.Errorf("Expected namespace %s, got %s", "org/department/team", namespace)
		}
		if policy != "policy" {
			t.Errorf("Expected policy %s, got %s", "policy", policy)
		}
		if rule != "rule" {
			t.Errorf("Expected rule %s, got %s", "rule", rule)
		}
	})

	// "/decision/simple/policy/rule"
	t.Run("valid simple path", func(t *testing.T) {
		path := "/decision/simple/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)

		if len(pathParts) < 3 {
			t.Fatalf("Expected valid parsing for %s", path)
		}

		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)

		if namespace != "simple" {
			t.Errorf("Expected namespace %s, got %s", "simple", namespace)
		}
		if policy != "policy" {
			t.Errorf("Expected policy %s, got %s", "policy", policy)
		}
		if rule != "rule" {
			t.Errorf("Expected rule %s, got %s", "rule", rule)
		}
	})

	t.Run("error - not enough segments (policy/rule)", func(t *testing.T) {
		path := "/decision/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		if len(pathParts) >= 3 {
			t.Fatalf("Expected error for path %s, but got valid parsing", path)
		}
	})

	t.Run("error - not enough segments (rule)", func(t *testing.T) {
		path := "/decision/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		if len(pathParts) >= 3 {
			t.Fatalf("Expected error for path %s, but got valid parsing", path)
		}
	})
}
