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
	"regexp"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/pack"
)

var (
	ErrPackFileNotFound   = errors.New("pack file not found")
	ErrPackFileLoadFailed = errors.New("pack file load failed")
)

var (
	PackFileName = (constants.APPNAME + "." + constants.PackFileExtension)
	NameRegex    = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]*)(\.[a-zA-Z][a-zA-Z0-9_-]*)*$`)
)

func IsValidPackName(name string) bool {
	return NameRegex.MatchString(name)
}

func LoadPack(ctx context.Context, root string) (_ *pack.PackFile, e error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	packPath, err := locatePackFile(ctx, root)
	if err != nil {
		return nil, errors.Wrap(err, "locate pack file")
	}

	stat, err := os.Stat(packPath)
	if err != nil {
		return nil, errors.Wrap(err, "stat pack file")
	}

	if stat.Size() == 0 {
		return nil, errors.New("pack file is empty")
	}

	// Read file content into memory
	fileContent, err := os.ReadFile(packPath)
	if err != nil {
		return nil, errors.Wrap(err, "read pack file")
	}

	// First decode into a map to check for unknown top-level keys
	var rawData map[string]interface{}
	if err := toml.Unmarshal(fileContent, &rawData); err != nil {
		return nil, errors.Wrap(err, "failed to parse pack file")
	}

	// Check for unknown top-level keys
	allowedKeys := map[string]bool{
		"schema":      true,
		"pack":        true,
		"engine":      true,
		"permissions": true,
		"metadata":    true,
	}
	for key := range rawData {
		if !allowedKeys[key] {
			return nil, errors.Errorf("unknown top-level table '[%s]'. Allowed tables are: schema, pack, engine, permissions, metadata", key)
		}
	}

	// Now decode into the struct
	var p pack.PackFile
	if err := toml.Unmarshal(fileContent, &p); err != nil {
		return nil, errors.Wrap(err, "failed to parse pack file")
	}

	if p.SchemaVersion == nil {
		return nil, errors.New("schema version is required")
	}

	if p.Pack == nil || p.Pack.Name == "" {
		return nil, errors.New("name is required")
	}

	// make sure that the name is an identity
	if !IsValidPackName(p.Pack.Name) {
		return nil, errors.New("name must be a valid identity")
	}

	// Check that if engine table exists, it must have sentrie field
	if engineData, exists := rawData["engine"]; exists {
		if engineMap, ok := engineData.(map[string]interface{}); ok {
			if _, hasSentrie := engineMap["sentrie"]; !hasSentrie {
				return nil, errors.New("engine table exists but 'sentrie' field is required")
			}
			c, ok := engineMap["sentrie"].(string)
			if ok && len(c) == 0 {
				return nil, errors.New("engine table exists but 'sentrie' field is required")
			}
		}
	}

	// Validate against JSON Schema
	if err := ValidatePackFile(&p); err != nil {
		return nil, errors.Wrap(err, "schema validation failed")
	}

	p.Location = filepath.Dir(packPath)

	return &p, nil
}

func locatePackFile(ctx context.Context, root string) (string, error) {
	if root == "/" {
		return "", errors.New("cannot search from filesystem root")
	}

	if len(strings.TrimSpace(root)) == 0 {
		return "", errors.New("root is empty")
	}

	// get the absolute path to the root
	root, err := filepath.Abs(root)
	if err != nil {
		return "", errors.Wrap(err, "failed to get absolute path to root")
	}

	// locate the pack file
	// if the root is a file, we take the containing directory of the file
	// then we check if the pack file exists in the root directory
	// if it does, we load it and return
	// if it doesn't, we walk up the directory tree
	// till we find one - if we reach the root and don't find it, we return an error
	info, err := os.Stat(root)
	if err != nil {
		return "", errors.Wrap(err, "failed to locate pack file")
	}

	// if the name is "sentrie.pack.toml", we use it
	if info.Name() == PackFileName {
		return root, nil
	}

	// the name is not "sentrie.pack.toml", we try to find it in the parent directory
	packFilePath := filepath.Join(root, PackFileName)

	// if we have a packfile here - we use it
	if _, err := os.Stat(packFilePath); err == nil {
		return packFilePath, nil
	}

	// otherwise, we walk up the directory tree till we find it or we reach root
	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		root = filepath.Dir(root)
		if root == "/" || (runtime.GOOS == "windows" && strings.HasSuffix(root, `:\` /* a drive letter */)) {
			break
		}
		if _, err := os.Stat(filepath.Join(root, PackFileName)); err == nil {
			return filepath.Join(root, PackFileName), nil
		}
	}

	return "", ErrPackFileNotFound
}
