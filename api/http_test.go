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

	"github.com/sentrie-sh/sentrie/ast"
)

func (s *APITestSuite) TestPathParsing() {
	s.Run("valid deep path", func() {
		path := "/decision/sh/sentra/auth/v1/user/allow"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		s.Require().GreaterOrEqual(len(pathParts), 3, "expected valid parsing for %s", path)
		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)
		s.Equal("sh/sentra/auth/v1", namespace)
		s.Equal("user", policy)
		s.Equal("allow", rule)
	})
	s.Run("valid org/department/team path", func() {
		path := "/decision/org/department/team/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		s.Require().GreaterOrEqual(len(pathParts), 3, "expected valid parsing for %s", path)
		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)
		s.Equal("org/department/team", namespace)
		s.Equal("policy", policy)
		s.Equal("rule", rule)
	})
	s.Run("valid simple path", func() {
		path := "/decision/simple/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		s.Require().GreaterOrEqual(len(pathParts), 3, "expected valid parsing for %s", path)
		rule := pathParts[len(pathParts)-1]
		policy := pathParts[len(pathParts)-2]
		namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)
		s.Equal("simple", namespace)
		s.Equal("policy", policy)
		s.Equal("rule", rule)
	})
	s.Run("error - not enough segments (policy/rule)", func() {
		path := "/decision/policy/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		s.Require().Less(len(pathParts), 3, "expected error for path %s", path)
	})
	s.Run("error - not enough segments (rule)", func() {
		path := "/decision/rule"
		pathParts := strings.Split(strings.TrimPrefix(path, "/decision/"), ast.FQNSeparator)
		s.Require().Less(len(pathParts), 3, "expected error for path %s", path)
	})
}
