// SPDX-License-Identifier: Apache-2.0

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

import "github.com/sentrie-sh/sentrie/tokens"

// TestFQNString tests the String() method of FQN
func (s *AstTestSuite) TestFQNString() {
	// Test empty FQN
	emptyFQN := FQN{}
	s.Equal("", emptyFQN.String())

	// Test single segment
	singleFQN := NewFQN([]string{"com"}, tokens.Range{})
	s.Equal("com", singleFQN.String())

	// Test multiple segments
	multiFQN := NewFQN([]string{"com", "example", "foo"}, tokens.Range{})
	s.Equal("com/example/foo", multiFQN.String())

	// Test with empty segments
	emptySegmentsFQN := NewFQN([]string{"com", "", "foo"}, tokens.Range{})
	s.Equal("com//foo", emptySegmentsFQN.String())
}

// TestCreateFQN tests the CreateFQN function
func (s *AstTestSuite) TestCreateFQN() {
	// Test with empty base
	base := NewFQN([]string{}, tokens.Range{})
	result := CreateFQN(base, "test")
	expected := NewFQN([]string{"test"}, tokens.Range{})
	s.Equal(expected, result)

	// Test with non-empty base
	base = NewFQN([]string{"com", "example"}, tokens.Range{})
	result = CreateFQN(base, "foo")
	expected = NewFQN([]string{"com", "example", "foo"}, tokens.Range{})
	s.Equal(expected, result)

	// Test with single segment base
	base = NewFQN([]string{"com"}, tokens.Range{})
	result = CreateFQN(base, "example")
	expected = NewFQN([]string{"com", "example"}, tokens.Range{})
	s.Equal(expected, result)
}

// TestFQNIsParentOf tests the IsParentOf method
func (s *AstTestSuite) TestFQNIsParentOf() {
	testCases := []struct {
		desc     string
		fqn      FQN
		another  FQN
		expected bool
	}{
		{
			desc:     "com.example is parent of com.example.foo",
			fqn:      NewFQN([]string{"com", "example"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			expected: true,
		},
		{
			desc:     "com.example is not parent of com.example.foo.bar",
			fqn:      NewFQN([]string{"com", "example"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "foo", "bar"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "com.example is not parent of com.example2.foo",
			fqn:      NewFQN([]string{"com", "example"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example2", "foo"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Self is not a parent of self",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Empty FQN is not parent of anything",
			fqn:      NewFQN([]string{}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Single segment is parent of two segments",
			fqn:      NewFQN([]string{"com"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example"}, tokens.Range{}),
			expected: true,
		},
		{
			desc:     "Different root segments",
			fqn:      NewFQN([]string{"com", "example"}, tokens.Range{}),
			another:  NewFQN([]string{"org", "example", "foo"}, tokens.Range{}),
			expected: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.desc, func() {
			actual := tc.fqn.IsParentOf(tc.another)
			s.Equal(tc.expected, actual, "FQN %v.IsParentOf(%v) should be %v", tc.fqn, tc.another, tc.expected)
		})
	}
}

// TestFQNIsChildOf tests the IsChildOf method
func (s *AstTestSuite) TestFQNIsChildOf() {
	testCases := []struct {
		desc     string
		fqn      FQN
		another  FQN
		expected bool
	}{
		{
			desc:     "com.example.foo is child of com.example",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example"}, tokens.Range{}),
			expected: true,
		},
		{
			desc:     "com.example.foo is not child of com.example.bar",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "bar"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "com.example.foo is not child of com.example2.foo",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example2", "foo"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Self is not a child of self",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Empty FQN is not child of anything",
			fqn:      NewFQN([]string{}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Two segments is child of single segment",
			fqn:      NewFQN([]string{"com", "example"}, tokens.Range{}),
			another:  NewFQN([]string{"com"}, tokens.Range{}),
			expected: true,
		},
		{
			desc:     "Different root segments",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"org", "example"}, tokens.Range{}),
			expected: false,
		},
		{
			desc:     "Same length but different segments",
			fqn:      NewFQN([]string{"com", "example", "foo"}, tokens.Range{}),
			another:  NewFQN([]string{"com", "example", "bar"}, tokens.Range{}),
			expected: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.desc, func() {
			actual := tc.fqn.IsChildOf(tc.another)
			s.Equal(tc.expected, actual, "FQN %v.IsChildOf(%v) should be %v", tc.fqn, tc.another, tc.expected)
		})
	}
}

// TestFQNEdgeCases tests edge cases for FQN operations
func (s *AstTestSuite) TestFQNEdgeCases() {
	// Test with nil/empty slices
	emptyFQN := NewFQN([]string{}, tokens.Range{})
	s.Equal("", emptyFQN.String())
	s.False(emptyFQN.IsParentOf(NewFQN([]string{"test"}, tokens.Range{})))
	s.False(emptyFQN.IsChildOf(NewFQN([]string{"test"}, tokens.Range{})))

	// Test with single empty string
	emptyStringFQN := NewFQN([]string{""}, tokens.Range{})
	s.Equal("", emptyStringFQN.String())

	// Test with multiple empty strings
	multipleEmptyFQN := NewFQN([]string{"", "", ""}, tokens.Range{})
	s.Equal("//", multipleEmptyFQN.String())

	// Test very long FQN
	longFQN := NewFQN([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, tokens.Range{})
	expectedLong := "a/b/c/d/e/f/g/h/i/j"
	s.Equal(expectedLong, longFQN.String())
}

// TestFQNConsistency tests that IsParentOf and IsChildOf are consistent
func (s *AstTestSuite) TestFQNConsistency() {
	parent := NewFQN([]string{"com", "example"}, tokens.Range{})
	child := NewFQN([]string{"com", "example", "foo"}, tokens.Range{})

	// If A is parent of B, then B should be child of A
	s.True(parent.IsParentOf(child))
	s.True(child.IsChildOf(parent))

	// If A is child of B, then B should be parent of A
	s.True(child.IsChildOf(parent))
	s.True(parent.IsParentOf(child))

	// Self should not be parent or child of self
	self := NewFQN([]string{"com", "example", "foo"}, tokens.Range{})
	s.False(self.IsParentOf(self))
	s.False(self.IsChildOf(self))
}

// TestFQNSeparator tests the FQNSeparator constant
func (s *AstTestSuite) TestFQNSeparator() {
	s.Equal("/", FQNSeparator)

	// Verify that the separator is used in string representation
	fqn := NewFQN([]string{"a", "b", "c"}, tokens.Range{})
	expected := "a" + FQNSeparator + "b" + FQNSeparator + "c"
	s.Equal(expected, fqn.String())
}

// TestFQNImplementationCorrectness verifies the implementation works correctly
func (s *AstTestSuite) TestFQNImplementationCorrectness() {
	parent := FQN{Parts: []string{"com", "example"}}
	child := FQN{Parts: []string{"com", "example", "foo"}}

	// These should work correctly
	s.True(parent.IsParentOf(child), "Parent should be parent of child")
	s.True(child.IsChildOf(parent), "Child should be child of parent")

	// The string representations use consistent separators
	s.Equal("com/example", parent.String())
	s.Equal("com/example/foo", child.String())

	// Verify the prefix check works correctly
	expectedPrefix := parent.String() + FQNSeparator
	s.True(child.String()[:len(expectedPrefix)] == expectedPrefix, "Child should start with parent prefix")
}
