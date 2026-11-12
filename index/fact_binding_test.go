// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/xerr"
)

func factPolicy(facts string) string {
	return fmt.Sprintf("namespace t\npolicy p {\n%s\n  rule r = { yield true }\n  export decision of r\n}\n", facts)
}

// Two facts may not share a declared name: it is the key a caller supplies
// the fact under, even when their aliases differ (#147).
func (s *IndexTestSuite) TestDuplicateDeclaredFactNameIsAnIndexError() {
	err := s.buildSources([2]string{"f.sentrie", factPolicy("  fact user: string as a\n  fact user: string as b")})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact name 'user'")

	// the same name with the same (default) alias is the existing conflict
	s.Require().Error(s.buildSources([2]string{"f.sentrie", factPolicy("  fact user: string\n  fact user: number")}))

	// different names with different aliases are fine
	s.Require().NoError(s.buildSources([2]string{"f.sentrie", factPolicy("  fact user: string as a\n  fact other: string as b")}))
}

// A literal default of the wrong kind, and a null default on a type that does
// not accept null, fail while the index is built (so validate, exec, bench and
// serve all reject them); matching and nullable defaults pass.
func (s *IndexTestSuite) TestFactDefaultLiteralKindIsCheckedAtIndexBuild() {
	for _, tc := range []struct {
		name    string
		fact    string
		wantErr string
	}{
		{"string for number", `fact n?: number default "x"`, "default value of fact 'n'"},
		{"number for string", `fact n?: string default 1`, "default value of fact 'n'"},
		{"bool for number", `fact n?: number default true`, "default value of fact 'n'"},
		{"null for non-nullable", `fact n?: number default null`, "cannot have a null default"},
		{"null for string", `fact n?: string default null`, "cannot have a null default"},
		{"matching number", `fact n?: number default 1`, ""},
		{"matching string", `fact n?: string default "x"`, ""},
		{"null for nullable", `fact n?: number? default null`, ""},
		{"null for any", `fact n?: any default null`, ""},
		{"non-literal wrong default waits for runtime", `fact n?: number default ["x"][0]`, ""},
	} {
		s.Run(tc.name, func() {
			err := s.buildSources([2]string{"f.sentrie", factPolicy("  " + tc.fact)})
			if tc.wantErr == "" {
				s.NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), tc.wantErr)
			s.True(errors.Is(err, xerr.ErrIndex), "an index error: %v", err)
		})
	}
}

// The by-name lookup callers use, against the alias-keyed Policy.Facts.
func (s *IndexTestSuite) TestPolicyFactByNameAndAliases() {
	ctx := s.T().Context()
	idx := s.indexSources([2]string{"f.sentrie", factPolicy("  fact user: string as u\n  fact zone?: string")})
	s.Require().NoError(idx.Validate(ctx))
	pol := idx.Namespaces["t"].Policies["p"]
	s.Require().NotNil(pol)

	s.Same(pol.Facts["u"], pol.FactByName("user"))
	s.Same(pol.Facts["zone"], pol.FactByName("zone"))
	s.Nil(pol.FactByName("u"), "an alias is not a name")
	s.Equal(map[string]string{"user": "u", "zone": "zone"}, pol.FactAliasesByName())
}

func (s *IndexTestSuite) indexSources(files ...[2]string) *Index {
	ctx := s.T().Context()
	idx := CreateIndex()
	for _, f := range files {
		prog, err := parser.NewParserFromString(f[1], f[0]).ParseProgram(ctx)
		s.Require().NoError(err)
		s.Require().NoError(idx.AddProgram(ctx, prog))
	}
	return idx
}

// buildSources adds each source to one index and validates it, returning the
// first error: fact declarations are checked while the program is added.
func (s *IndexTestSuite) buildSources(files ...[2]string) error {
	ctx := s.T().Context()
	idx := CreateIndex()
	for _, f := range files {
		prog, err := parser.NewParserFromString(f[1], f[0]).ParseProgram(ctx)
		s.Require().NoError(err)
		if err := idx.AddProgram(ctx, prog); err != nil {
			return err
		}
	}
	return idx.Validate(ctx)
}
