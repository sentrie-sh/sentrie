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

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPack_ValidMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	p, err := LoadPack(ctx, tmpDir)

	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, uint64(1), p.SchemaVersion.Version)
	assert.Equal(t, "test_pack", p.Pack.Name)
	assert.Equal(t, "0.1.0", p.Pack.Version.String())
}

func TestLoadPack_ValidFull(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
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
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	p, err := LoadPack(ctx, tmpDir)

	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "example.pack", p.Pack.Name)
	assert.Equal(t, "1.2.3-alpha.1", p.Pack.Version.String())
	assert.Equal(t, "A test pack", p.Pack.Description)
	assert.Equal(t, "MIT", p.Pack.License)
	assert.Equal(t, "https://github.com/example/pack", p.Pack.Repository)
	assert.Len(t, p.Pack.Authors, 2)
	assert.NotNil(t, p.Engine)
	assert.NotNil(t, p.Permissions)
	assert.Len(t, p.Permissions.FSRead, 2)
	assert.Len(t, p.Permissions.Net, 2)
	assert.Len(t, p.Permissions.Env, 2)
}

func TestLoadPack_MissingSchema(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[pack]
name = "test_pack"
version = "0.1.0"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema version is required")
}

func TestLoadPack_MissingPack(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestLoadPack_MissingPackVersion(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
	assert.Contains(t, err.Error(), "version")
}

func TestLoadPack_InvalidSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 2

[pack]
name = "test_pack"
version = "0.1.0"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

func TestLoadPack_UnknownTopLevelTable(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[unknown_table]
field = "value"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown top-level table '[unknown_table]'")
}

func TestLoadPack_InvalidPackName(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "123invalid"
version = "0.1.0"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name must be a valid identity")
}

func TestLoadPack_InvalidVersionFormat(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "not-a-version"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	// Should fail during TOML parsing or schema validation
	assert.NotNil(t, err)
}

func TestLoadPack_InvalidRepositoryFormat(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"
repository = "not-a-valid-uri"
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

func TestLoadPack_InvalidEnvVarPattern(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[permissions]
env = ["invalid-env-var", "also_invalid"]
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

func TestLoadPack_ValidEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[permissions]
env = ["AWS_REGION", "AWS_PROFILE", "HOME"]
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	p, err := LoadPack(ctx, tmpDir)

	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.Permissions)
	assert.Len(t, p.Permissions.Env, 3)
}

func TestLoadPack_EngineWithoutSentrie(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[engine]
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	_, err := LoadPack(ctx, tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sentrie")
}

func TestLoadPack_MetadataAllowsArbitraryFields(t *testing.T) {
	tmpDir := t.TempDir()
	packFile := filepath.Join(tmpDir, PackFileName)

	content := `[schema]
version = 1

[pack]
name = "test_pack"
version = "0.1.0"

[metadata]
custom_field = "value"
nested = { key = "value" }
array = [1, 2, 3]
`
	require.NoError(t, os.WriteFile(packFile, []byte(content), 0644))

	ctx := context.Background()
	p, err := LoadPack(ctx, tmpDir)

	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.Metadata)
}
