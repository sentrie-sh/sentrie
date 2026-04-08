// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package loader

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/xeipuuv/gojsonschema"
)

func (s *LoaderTestSuite) TestLoadPackWrapsLocateError() {
	_, err := LoadPack(context.Background(), "/")
	s.Require().Error(err)
	s.Contains(err.Error(), "locate pack file")
	s.Contains(err.Error(), "cannot search from filesystem root")
}

func (s *LoaderTestSuite) TestLoadPackRejectsEmptyFile() {
	dir := s.T().TempDir()
	path := filepath.Join(dir, PackFileName)
	s.Require().NoError(os.WriteFile(path, nil, 0o644))

	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Equal("pack file is empty", err.Error())
}

func (s *LoaderTestSuite) TestLoadPackRejectsEmptyEngineSentrieString() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "ok.pack"
version = "0.1.0"

[engine]
sentrie = ""
`)
	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "engine table exists but 'sentrie' field is required")
}

func (s *LoaderTestSuite) TestLocatePackFileRejectsWhitespaceRoot() {
	_, err := locatePackFile(context.Background(), "   \t  ")
	s.Require().Error(err)
	s.Equal("root is empty", err.Error())
}

func (s *LoaderTestSuite) TestLoadPackWrapsReadError() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "ok.pack"
version = "0.1.0"
`)
	packPath := filepath.Join(dir, PackFileName)
	s.Require().NoError(os.Chmod(packPath, 0o000))
	defer func() { _ = os.Chmod(packPath, 0o600) }()

	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "read pack file")
}

func (s *LoaderTestSuite) TestLoadPackWrapsSecondTomlParseError() {
	dir := s.writePackDir(`[schema]
version = "oops"

[pack]
name = "ok.pack"
version = "0.1.0"
`)

	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to parse pack file")
}

func (s *LoaderTestSuite) TestLocatePackFileWrapsStatLookupError() {
	_, err := locatePackFile(context.Background(), filepath.Join(s.T().TempDir(), "missing"))
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to locate pack file")
}

func (s *LoaderTestSuite) TestValidatePackFileSchemaFailureWraps() {
	invalid := &pack.PackFile{
		SchemaVersion: &pack.SentrieSchema{Version: 1},
		Pack: &pack.PackInformation{
			Name: "ok.pack",
		},
	}
	err := ValidatePackFile(invalid)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
}

func (s *LoaderTestSuite) TestLoadPackWrapsStatPackFileError() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "ok.pack"
version = "0.1.0"
`)
	prev := statPackFile
	statPackFile = func(string) (os.FileInfo, error) {
		return nil, errors.New("stat failed")
	}
	defer func() { statPackFile = prev }()

	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "stat pack file")
	s.Contains(err.Error(), "stat failed")
}

func (s *LoaderTestSuite) TestLocatePackFileWrapsFilepathAbsError() {
	prev := filepathAbs
	filepathAbs = func(string) (string, error) {
		return "", errors.New("abs failed")
	}
	defer func() { filepathAbs = prev }()

	_, err := locatePackFile(context.Background(), s.T().TempDir())
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to get absolute path to root")
	s.Contains(err.Error(), "abs failed")
}

func (s *LoaderTestSuite) TestValidatePackFileWrapsSchemaValidateError() {
	prev := validatePackDocument
	validatePackDocument = func(gojsonschema.JSONLoader) (*gojsonschema.Result, error) {
		return nil, errors.New("validate failed")
	}
	defer func() { validatePackDocument = prev }()

	valid := &pack.PackFile{
		SchemaVersion: &pack.SentrieSchema{Version: 1},
		Pack: &pack.PackInformation{
			Name:    "ok.pack",
			Version: semver.MustParse("0.1.0"),
		},
	}
	err := ValidatePackFile(valid)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
	s.Contains(err.Error(), "validate failed")
}
