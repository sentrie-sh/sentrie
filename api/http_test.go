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

	"github.com/binaek/sentra/ast"
)

func TestPathParsing(t *testing.T) {
	tests := []struct {
		path           string
		expectedNS     string
		expectedPolicy string
		expectedRule   string
		shouldError    bool
	}{
		{
			path:           "/decision/sh/sentra/auth/v1/user/allow",
			expectedNS:     "sh/sentra/auth/v1",
			expectedPolicy: "user",
			expectedRule:   "allow",
			shouldError:    false,
		},
		{
			path:           "/decision/org/department/team/policy/rule",
			expectedNS:     "org/department/team",
			expectedPolicy: "policy",
			expectedRule:   "rule",
			shouldError:    false,
		},
		{
			path:           "/decision/simple/policy/rule",
			expectedNS:     "simple",
			expectedPolicy: "policy",
			expectedRule:   "rule",
			shouldError:    false,
		},
		{
			path:        "/decision/policy/rule",
			shouldError: true, // Not enough segments
		},
		{
			path:        "/decision/rule",
			shouldError: true, // Not enough segments
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Simulate the path parsing logic from handleDecision
			pathParts := strings.Split(strings.TrimPrefix(tt.path, "/decision/"), ast.FQNSeparator)

			if len(pathParts) < 3 {
				if !tt.shouldError {
					t.Errorf("Expected error for path %s, but got valid parsing", tt.path)
				}
				return
			}

			if tt.shouldError {
				t.Errorf("Expected error for path %s, but got valid parsing", tt.path)
				return
			}

			rule := pathParts[len(pathParts)-1]
			policy := pathParts[len(pathParts)-2]
			namespace := strings.Join(pathParts[:len(pathParts)-2], ast.FQNSeparator)

			if namespace != tt.expectedNS {
				t.Errorf("Expected namespace %s, got %s", tt.expectedNS, namespace)
			}
			if policy != tt.expectedPolicy {
				t.Errorf("Expected policy %s, got %s", tt.expectedPolicy, policy)
			}
			if rule != tt.expectedRule {
				t.Errorf("Expected rule %s, got %s", tt.expectedRule, rule)
			}
		})
	}
}
