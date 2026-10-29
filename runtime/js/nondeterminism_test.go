// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

func (s *JSTestSuite) TestNondeterministicExportLookup() {
	s.Equal("@sentrie/uuid", BuiltinModuleKey("uuid"))
	s.True(IsBuiltinModuleKey("@sentrie/uuid"))
	s.False(IsBuiltinModuleKey("@local/uuid"))

	s.True(IsNondeterministicExport("@sentrie/uuid", "v4"))
	s.True(IsNondeterministicExport("@sentrie/time", "now"))
	s.True(IsNondeterministicExport("@sentrie/js", "random"))
	s.False(IsNondeterministicExport("@sentrie/js", "now"), "removed in #207")
	s.False(IsNondeterministicExport("@sentrie/time", "parse"))
	s.False(IsNondeterministicExport("@sentrie/hash", "sha256"))
	s.False(IsNondeterministicExport("@local/uuid", "v4"))

	s.Equal([]string{"v4", "v6", "v7"}, NondeterministicExports("@sentrie/uuid"))
	s.Nil(NondeterministicExports("@sentrie/hash"))
	s.Equal([]string{"@sentrie/js", "@sentrie/time", "@sentrie/uuid"}, NondeterministicModules())

	// The returned slice is a copy.
	got := NondeterministicExports("@sentrie/uuid")
	got[0] = "mutated"
	s.True(IsNondeterministicExport("@sentrie/uuid", "v4"))
}
