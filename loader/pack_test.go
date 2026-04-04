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

package loader

import "context"

func (s *LoaderTestSuite) TestLoadPack_ValidMinimal() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"
`)
	ctx := context.Background()
	p, err := LoadPack(ctx, dir)
	s.Require().NoError(err)
	s.NotNil(p)
	s.Equal(uint64(1), p.SchemaVersion.Version)
	s.Equal("test_pack", p.Pack.Name)
	s.Equal("0.1.0", p.Pack.Version.String())
}

func (s *LoaderTestSuite) TestLoadPack_ValidFull() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "example.pack"
version = "1.2.3-alpha.1"
description = "A test pack"
license = "MIT"
repository = "https://github.com/example/pack"

[pack.authors]
"John Doe" = "john@example.com"
"Jane Smith" = "jane@example.com"

[engine]
sentrie = ">=0.1.0 <2.0.0"

[permissions]
fs_read = ["./data/**", "/etc/ssl/certs/**"]
net = ["https://api.example.com", "https://sts.amazonaws.com"]
env = ["AWS_REGION", "AWS_PROFILE"]

[metadata]
category = "cloud"
maturity = "beta"
`)
	ctx := context.Background()
	p, err := LoadPack(ctx, dir)
	s.Require().NoError(err)
	s.NotNil(p)
	s.Equal("example.pack", p.Pack.Name)
	s.Equal("1.2.3-alpha.1", p.Pack.Version.String())
	s.Equal("A test pack", p.Pack.Description)
	s.Equal("MIT", p.Pack.License)
	s.Equal("https://github.com/example/pack", p.Pack.Repository)
	s.Len(p.Pack.Authors, 2)
	s.NotNil(p.Engine)
	s.NotNil(p.Permissions)
	s.Len(p.Permissions.FSRead, 2)
	s.Len(p.Permissions.Net, 2)
	s.Len(p.Permissions.Env, 2)
}

func (s *LoaderTestSuite) TestLoadPack_MissingSchema() {
	dir := s.writePackDir(`[pack]
name = "test_pack"
version = "0.1.0"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema version is required")
}

func (s *LoaderTestSuite) TestLoadPack_MissingPack() {
	dir := s.writePackDir(`[schema]
version = 1
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "name is required")
}

func (s *LoaderTestSuite) TestLoadPack_MissingPackVersion() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
	s.Contains(err.Error(), "version")
}

func (s *LoaderTestSuite) TestLoadPack_InvalidSchemaVersion() {
	dir := s.writePackDir(`[schema]
version = 2

[pack]
name = "test_pack"
version = "0.1.0"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
}

func (s *LoaderTestSuite) TestLoadPack_UnknownTopLevelTable() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[unknown_table]
field = "value"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "unknown top-level table '[unknown_table]'")
}

func (s *LoaderTestSuite) TestLoadPack_InvalidPackName() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "123invalid"
version = "0.1.0"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "name must be a valid identity")
}

func (s *LoaderTestSuite) TestLoadPack_InvalidVersionFormat() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "not-a-version"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.NotNil(err)
}

func (s *LoaderTestSuite) TestLoadPack_InvalidRepositoryFormat() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"
repository = "not-a-valid-uri"
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
}

func (s *LoaderTestSuite) TestLoadPack_InvalidEnvVarPattern() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[permissions]
env = ["invalid-env-var", "also_invalid"]
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "schema validation failed")
}

func (s *LoaderTestSuite) TestLoadPack_ValidEnvVars() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[permissions]
env = ["AWS_REGION", "AWS_PROFILE", "HOME"]
`)
	ctx := context.Background()
	p, err := LoadPack(ctx, dir)
	s.Require().NoError(err)
	s.NotNil(p)
	s.NotNil(p.Permissions)
	s.Len(p.Permissions.Env, 3)
}

func (s *LoaderTestSuite) TestLoadPack_EngineWithoutSentrie() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[engine]
`)
	ctx := context.Background()
	_, err := LoadPack(ctx, dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "sentrie")
}

func (s *LoaderTestSuite) TestLoadPack_MetadataAllowsArbitraryFields() {
	dir := s.writePackDir(`[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[metadata]
custom_field = "value"
nested = { key = "value" }
array = [1, 2, 3]
`)
	ctx := context.Background()
	p, err := LoadPack(ctx, dir)
	s.Require().NoError(err)
	s.NotNil(p)
	s.NotNil(p.Metadata)
}
