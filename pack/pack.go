// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"encoding/json"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
)

type Pack struct {
	Pack     *PackFile
	Programs []*ast.Program
}

func NewPackFile(name string) *PackFile {
	return &PackFile{
		SchemaVersion: &SentrieSchema{Version: 1},
		Pack: &PackInformation{
			Name:    name,
			Version: semver.MustParse("0.0.1"),
		},
	}
}

type PackFile struct {
	SchemaVersion *SentrieSchema   `toml:"schema" json:"schema"`
	Pack          *PackInformation `toml:"pack" json:"pack"`
	Permissions   *Permissions     `toml:"permissions,omitempty" json:"permissions"`
	Engine        *Engine          `toml:"engine,omitempty" json:"engine"`
	Metadata      map[string]any   `toml:"metadata,omitempty" json:"metadata"`
	// Data maps each static data name to a JSON file, relative to the pack
	// root (the `[data]` table). A policy reads an entry through a
	// `data <name>` declaration.
	Data     map[string]string `toml:"data,omitempty" json:"data"`
	Location string            `toml:"-" json:"-"`
	// DataValues holds each Data entry decoded and boxed, keyed by name. Like
	// Location it is not a manifest field: the loader fills it once at pack
	// load. The values are shared by every evaluation and must never be
	// mutated.
	DataValues map[string]box.Value `toml:"-" json:"-"`
}

type SentrieSchema struct {
	Version uint64 `toml:"version" json:"version"`
}

type PackInformation struct {
	Name        string            `toml:"name" json:"name"`
	Version     *semver.Version   `toml:"version" json:"version"`
	Description string            `toml:"description,omitempty" json:"description"`
	License     string            `toml:"license,omitempty" json:"license"`
	Repository  string            `toml:"repository,omitempty" json:"repository"`
	Authors     map[string]string `toml:"authors,omitempty" json:"authors"`
}

type Engine struct {
	Sentrie *semver.Constraints `toml:"sentrie" json:"sentrie"`
}

// MarshalJSON implements json.Marshaler for Engine to serialize semver.Constraints as string
// this is necessary because semver.Constraints does not implement json.Marshaler
func (e *Engine) MarshalJSON() ([]byte, error) {
	type Alias Engine
	aux := &struct {
		Sentrie string `json:"sentrie"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if e.Sentrie != nil {
		aux.Sentrie = e.Sentrie.String()
	}
	return json.Marshal(aux)
}

// UnmarshalJSON implements json.Unmarshaler for Engine to deserialize semver.Constraints from string
func (e *Engine) UnmarshalJSON(data []byte) error {
	type Alias Engine
	aux := &struct {
		Sentrie string `json:"sentrie"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Sentrie != "" {
		constraint, err := semver.NewConstraint(aux.Sentrie)
		if err != nil {
			return err
		}
		e.Sentrie = constraint
	}
	return nil
}

type Permissions struct {
	FSRead []string `toml:"fs_read,omitempty" json:"fs_read"`
	Net    []string `toml:"net,omitempty" json:"net"`
	Env    []string `toml:"env,omitempty" json:"env"`
}

func (p *Permissions) CheckEnvAccess(name string) bool {
	return slices.Contains(p.Env, name)
}
