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

package ast

import "testing"

func TestFQNIsParentOf(t *testing.T) {
	testCases := []struct {
		desc     string
		fqn      FQN
		another  FQN
		expected bool
	}{
		{
			desc:     "com.example is parent of com.example.foo",
			fqn:      FQN{"com", "example"},
			another:  FQN{"com", "example", "foo"},
			expected: true,
		},
		{
			desc:     "com.example is not parent of com.example.foo.bar",
			fqn:      FQN{"com", "example"},
			another:  FQN{"com", "example", "foo", "bar"},
			expected: false,
		},
		{
			desc:     "com.example is not parent of com.example2.foo",
			fqn:      FQN{"com", "example"},
			another:  FQN{"com", "example2", "foo"},
			expected: false,
		},
		{
			desc:     "Self is not a parent of self",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example", "foo"},
			expected: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := tC.fqn.IsParentOf(tC.another)
			if actual != tC.expected {
				t.Errorf("expected %v, got %v", tC.expected, actual)
			}
		})
	}
}

func TestFQNIsChildOf(t *testing.T) {
	testCases := []struct {
		desc     string
		fqn      FQN
		another  FQN
		expected bool
	}{
		{
			desc:     "com.example.foo is child of com.example",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example"},
			expected: true,
		},
		{
			desc:     "com.example.foo is not child of com.example.bar",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example", "bar"},
			expected: false,
		},
		{
			desc:     "com.example.foo is not child of com.example2.foo",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example2", "foo"},
			expected: false,
		},
		{
			desc:     "Self is not a child of self",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example", "foo"},
			expected: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := tC.fqn.IsChildOf(tC.another)
			if actual != tC.expected {
				t.Errorf("expected %v, got %v", tC.expected, actual)
			}
		})
	}
}
