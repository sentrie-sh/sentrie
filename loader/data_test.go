// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sentrie-sh/sentrie/box"
)

const dataPackHeader = `[schema]
version = 1

[pack]
name = "data_pack"
version = "0.1.0"
`

// writeDataPack writes a manifest with the given [data] table body and the
// given files (path relative to the pack root -> content).
func (s *LoaderTestSuite) writeDataPack(dataTable string, files map[string]string) string {
	dir := s.writePackDir(dataPackHeader + "\n[data]\n" + dataTable)
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		s.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

func (s *LoaderTestSuite) TestLoadPackData_DecodesEachEntryOnce() {
	dir := s.writeDataPack(`roles = "data/roles.json"
allow = "allow.json"
`, map[string]string{
		"data/roles.json": `{"admin": ["read", "write"], "viewer": ["read"], "level": 3, "on": true, "none": null}`,
		"allow.json":      `["a", "b"]`,
	})

	p, err := LoadPack(s.T().Context(), dir)
	s.Require().NoError(err)
	s.Require().Len(p.DataValues, 2)

	roles := p.DataValues["roles"]
	s.Equal(box.ValueDict, roles.Kind())
	m, _ := roles.DictValue()
	admin, _ := m["admin"].ListValue()
	s.Require().Len(admin, 2)
	s.Equal("read", admin[0].Any())
	s.Equal(box.ValueNumber, m["level"].Kind())
	s.True(m["none"].IsNull())

	allow := p.DataValues["allow"]
	s.Equal(box.ValueList, allow.Kind())
}

func (s *LoaderTestSuite) TestLoadPackData_NoTableLeavesValuesNil() {
	dir := s.writePackDir(dataPackHeader)
	p, err := LoadPack(s.T().Context(), dir)
	s.Require().NoError(err)
	s.Nil(p.DataValues)
}

func (s *LoaderTestSuite) TestLoadPackData_Rejections() {
	cases := []struct {
		name    string
		table   string
		files   map[string]string
		wantErr []string
	}{
		{
			name:    "malformed JSON names the entry and file",
			table:   `roles = "data/roles.json"` + "\n",
			files:   map[string]string{"data/roles.json": `{"admin": [`},
			wantErr: []string{`data "roles" (data/roles.json)`, "invalid JSON"},
		},
		{
			name:    "trailing content after the document",
			table:   `roles = "roles.json"` + "\n",
			files:   map[string]string{"roles.json": `{"a": 1} {"b": 2}`},
			wantErr: []string{`data "roles" (roles.json)`, "unexpected content"},
		},
		{
			name:    "empty file",
			table:   `roles = "roles.json"` + "\n",
			files:   map[string]string{"roles.json": ``},
			wantErr: []string{`data "roles" (roles.json)`, "invalid JSON"},
		},
		{
			name:    "missing file",
			table:   `roles = "missing.json"` + "\n",
			wantErr: []string{`data "roles" (missing.json)`},
		},
		{
			name:    "absolute path",
			table:   `roles = "/etc/passwd"` + "\n",
			wantErr: []string{`data "roles" (/etc/passwd)`, "relative to the pack root"},
		},
		{
			name:    "parent escape",
			table:   `roles = "../outside.json"` + "\n",
			wantErr: []string{`data "roles" (../outside.json)`, "escapes the pack root"},
		},
		{
			name:    "nested parent escape",
			table:   `roles = "data/../../outside.json"` + "\n",
			wantErr: []string{"escapes the pack root"},
		},
		{
			name:    "directory",
			table:   `roles = "data"` + "\n",
			files:   map[string]string{"data/x.json": `{}`},
			wantErr: []string{`data "roles" (data)`, "not a regular file"},
		},
		{
			name:    "keyword name",
			table:   `string = "s.json"` + "\n",
			files:   map[string]string{"s.json": `"x"`},
			wantErr: []string{`data "string"`, "keyword"},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			dir := s.writeDataPack(tc.table, tc.files)
			_, err := LoadPack(s.T().Context(), dir)
			s.Require().Error(err)
			for _, want := range tc.wantErr {
				s.Contains(err.Error(), want)
			}
		})
	}
}

func (s *LoaderTestSuite) TestLoadPackData_SchemaRejectsBadNamesAndPaths() {
	for _, table := range []string{
		`"has-dash" = "a.json"` + "\n",
		`"1digit" = "a.json"` + "\n",
		`roles = ""` + "\n",
		`roles = 3` + "\n",
	} {
		dir := s.writeDataPack(table, nil)
		_, err := LoadPack(s.T().Context(), dir)
		s.Require().Error(err, table)
	}
}

func (s *LoaderTestSuite) TestLoadPackData_SymlinkEscapingRootRejected() {
	outside := filepath.Join(s.T().TempDir(), "secret.json")
	s.Require().NoError(os.WriteFile(outside, []byte(`{"secret": true}`), 0o644))

	dir := s.writeDataPack(`roles = "link.json"`+"\n", nil)
	if err := os.Symlink(outside, filepath.Join(dir, "link.json")); err != nil {
		s.T().Skipf("symlinks unavailable: %v", err)
	}

	_, err := LoadPack(s.T().Context(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), `data "roles" (link.json)`)
	s.Contains(err.Error(), "escapes")
}

func (s *LoaderTestSuite) TestLoadPackData_SymlinkInsideRootAllowed() {
	dir := s.writeDataPack(`roles = "link.json"`+"\n", map[string]string{"data/real.json": `{"ok": true}`})
	if err := os.Symlink(filepath.Join("data", "real.json"), filepath.Join(dir, "link.json")); err != nil {
		s.T().Skipf("symlinks unavailable: %v", err)
	}

	p, err := LoadPack(s.T().Context(), dir)
	s.Require().NoError(err)
	s.Equal(box.ValueDict, p.DataValues["roles"].Kind())
}

func (s *LoaderTestSuite) TestLoadPackData_OversizedFileRejected() {
	dir := s.writeDataPack(`big = "big.json"`+"\n", nil)
	// A sparse file over the limit: Stat reports the size without writing it.
	f, err := os.Create(filepath.Join(dir, "big.json"))
	s.Require().NoError(err)
	s.Require().NoError(f.Truncate(MaxDataFileSize + 1))
	s.Require().NoError(f.Close())

	_, err = LoadPack(s.T().Context(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), `data "big" (big.json)`)
	s.True(strings.Contains(err.Error(), "byte limit"), err.Error())
}

func (s *LoaderTestSuite) TestDecodeDataJSON_MatchesFactBoxing() {
	v, err := decodeDataJSON([]byte(`{"n": 1.5, "s": "x", "l": [1, {"k": false}]}`))
	s.Require().NoError(err)
	want := box.FromBoundaryAny(map[string]any{
		"n": 1.5, "s": "x", "l": []any{float64(1), map[string]any{"k": false}},
	})
	s.Equal(want.Any(), v.Any())
}
