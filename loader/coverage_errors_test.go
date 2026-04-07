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
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/pack"
)

func (s *LoaderTestSuite) TestLoadPackWrapsLocateError() {
	_, err := LoadPack(context.Background(), "/")
	s.Require().Error(err)
	s.Contains(err.Error(), "locate pack file")
	s.Contains(err.Error(), "cannot search from filesystem root")
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
