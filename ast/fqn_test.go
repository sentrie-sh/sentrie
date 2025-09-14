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

// TestFQNString tests the String() method of FQN
func (s *AstTestSuite) TestFQNString() {
	// Test empty FQN
	emptyFQN := FQN{}
	s.Equal("", emptyFQN.String())

	// Test single segment
	singleFQN := FQN{"com"}
	s.Equal("com", singleFQN.String())

	// Test multiple segments
	multiFQN := FQN{"com", "example", "foo"}
	s.Equal("com/example/foo", multiFQN.String())

	// Test with empty segments
	emptySegmentsFQN := FQN{"com", "", "foo"}
	s.Equal("com//foo", emptySegmentsFQN.String())
}

// TestCreateFQN tests the CreateFQN function
func (s *AstTestSuite) TestCreateFQN() {
	// Test with empty base
	base := FQN{}
	result := CreateFQN(base, "test")
	expected := FQN{"test"}
	s.Equal(expected, result)

	// Test with non-empty base
	base = FQN{"com", "example"}
	result = CreateFQN(base, "foo")
	expected = FQN{"com", "example", "foo"}
	s.Equal(expected, result)

	// Test with single segment base
	base = FQN{"com"}
	result = CreateFQN(base, "example")
	expected = FQN{"com", "example"}
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
		{
			desc:     "Empty FQN is not parent of anything",
			fqn:      FQN{},
			another:  FQN{"com", "example"},
			expected: false,
		},
		{
			desc:     "Single segment is parent of two segments",
			fqn:      FQN{"com"},
			another:  FQN{"com", "example"},
			expected: true,
		},
		{
			desc:     "Different root segments",
			fqn:      FQN{"com", "example"},
			another:  FQN{"org", "example", "foo"},
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
		{
			desc:     "Empty FQN is not child of anything",
			fqn:      FQN{},
			another:  FQN{"com", "example"},
			expected: false,
		},
		{
			desc:     "Two segments is child of single segment",
			fqn:      FQN{"com", "example"},
			another:  FQN{"com"},
			expected: true,
		},
		{
			desc:     "Different root segments",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"org", "example"},
			expected: false,
		},
		{
			desc:     "Same length but different segments",
			fqn:      FQN{"com", "example", "foo"},
			another:  FQN{"com", "example", "bar"},
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
	emptyFQN := FQN{}
	s.Equal("", emptyFQN.String())
	s.False(emptyFQN.IsParentOf(FQN{"test"}))
	s.False(emptyFQN.IsChildOf(FQN{"test"}))

	// Test with single empty string
	emptyStringFQN := FQN{""}
	s.Equal("", emptyStringFQN.String())

	// Test with multiple empty strings
	multipleEmptyFQN := FQN{"", "", ""}
	s.Equal("//", multipleEmptyFQN.String())

	// Test very long FQN
	longFQN := FQN{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	expectedLong := "a/b/c/d/e/f/g/h/i/j"
	s.Equal(expectedLong, longFQN.String())
}

// TestFQNConsistency tests that IsParentOf and IsChildOf are consistent
func (s *AstTestSuite) TestFQNConsistency() {
	parent := FQN{"com", "example"}
	child := FQN{"com", "example", "foo"}

	// If A is parent of B, then B should be child of A
	s.True(parent.IsParentOf(child))
	s.True(child.IsChildOf(parent))

	// If A is child of B, then B should be parent of A
	s.True(child.IsChildOf(parent))
	s.True(parent.IsParentOf(child))

	// Self should not be parent or child of self
	self := FQN{"com", "example", "foo"}
	s.False(self.IsParentOf(self))
	s.False(self.IsChildOf(self))
}

// TestFQNSeparator tests the FQNSeparator constant
func (s *AstTestSuite) TestFQNSeparator() {
	s.Equal("/", FQNSeparator)

	// Verify that the separator is used in string representation
	fqn := FQN{"a", "b", "c"}
	expected := "a" + FQNSeparator + "b" + FQNSeparator + "c"
	s.Equal(expected, fqn.String())
}

// TestFQNImplementationCorrectness verifies the implementation works correctly
func (s *AstTestSuite) TestFQNImplementationCorrectness() {
	parent := FQN{"com", "example"}
	child := FQN{"com", "example", "foo"}

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
