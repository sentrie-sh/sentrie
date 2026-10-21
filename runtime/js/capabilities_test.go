// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *JSTestSuite) mustGrant(src string) *capabilities.Capabilities {
	c, err := capabilities.Parse([]byte(src))
	s.Require().NoError(err)
	return c
}

func (s *JSTestSuite) newGrantTestRegistry() (*Registry, string) {
	packRoot := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(packRoot, "plain.ts"), []byte("export const x = 1;"), 0o644))
	s.Require().NoError(os.WriteFile(filepath.Join(packRoot, "needs_uuid.ts"),
		[]byte("const u = require('@sentrie/uuid');\nexport const id = () => u.v4();"), 0o644))
	reg := NewRegistry(packRoot)
	reg.RegisterGoBuiltin("uuid", BuiltinUuidGo)
	reg.RegisterGoBuiltin("hash", BuiltinHashGo)
	return reg, packRoot
}

func (s *JSTestSuite) TestRegistryNilGrantAllowsEverything() {
	reg, packRoot := s.newGrantTestRegistry()
	_, err := reg.LoadRequire(packRoot, "@sentrie/uuid")
	s.Require().NoError(err)
	_, err = reg.LoadRequire(packRoot, "./plain")
	s.Require().NoError(err)
	_, err = reg.PrepareUse("", []string{"sentrie", "uuid"}, packRoot)
	s.Require().NoError(err)
}

func (s *JSTestSuite) TestRegistryRefusesUngrantedBuiltinModule() {
	reg, packRoot := s.newGrantTestRegistry()
	reg.SetCapabilities(s.mustGrant(`{"version": 1, "modules": ["@sentrie/hash"]}`))

	_, err := reg.LoadRequire(packRoot, "@sentrie/uuid")
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.Contains(err.Error(), `module "@sentrie/uuid"`)

	_, err = reg.PrepareUse("", []string{"sentrie", "uuid"}, packRoot)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))

	_, err = reg.LoadRequire(packRoot, "@sentrie/hash")
	s.Require().NoError(err)
	_, err = reg.LoadRequire(packRoot, "./plain")
	s.Require().NoError(err, "local modules stay granted when local_modules is absent")
}

func (s *JSTestSuite) TestRegistryRefusesLocalModulesWhenNotGranted() {
	reg, packRoot := s.newGrantTestRegistry()
	reg.SetCapabilities(s.mustGrant(`{"version": 1, "local_modules": false}`))

	for _, spec := range []string{"./plain", "@local/plain.ts", "/plain.ts"} {
		_, err := reg.LoadRequire(packRoot, spec)
		s.Require().Error(err, spec)
		s.True(errors.Is(err, xerr.ErrNotGranted), spec)
		s.Contains(err.Error(), "local module", spec)
	}
	_, err := reg.PrepareUse("./plain", nil, packRoot)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	_, err = reg.PrepareUse("", []string{"local", "plain"}, packRoot)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))

	_, err = reg.LoadRequire(packRoot, "@sentrie/uuid")
	s.Require().NoError(err)
}

// require() inside a granted local module is held to the same grant.
func (s *JSTestSuite) TestRequireInsideTypeScriptIsHeldToTheGrant() {
	reg, packRoot := s.newGrantTestRegistry()
	reg.SetCapabilities(s.mustGrant(`{"version": 1, "modules": ["@sentrie/hash"], "local_modules": true}`))

	ar := NewAliasRuntime(reg, packRoot)
	_, err := ar.Require(s.T().Context(), packRoot, "./needs_uuid")
	s.Require().Error(err)
	s.Contains(err.Error(), `module "@sentrie/uuid"`)
	s.Contains(err.Error(), xerr.ErrNotGranted.Error())

	ar = NewAliasRuntime(reg, packRoot)
	_, err = ar.Require(s.T().Context(), packRoot, "./plain")
	s.Require().NoError(err)
}
