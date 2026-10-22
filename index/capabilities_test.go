// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"errors"
	"strings"

	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/xerr"
)

const capabilitiesTestPolicy = `namespace com/ex
policy pol {
  fact started?: number as started default now()
  use { v4 } from @sentrie/uuid as uuid
  use { sha256 } from @sentrie/hash as hash
  use { helper } from "./helper" as local
  use { other } from @local/lib/other as other
  rule shadowed = {
    let count = (x) => { yield 1 }
    yield count(1) > 0
  }
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}`

func (s *IndexTestSuite) mustParseCapabilities(src string) *capabilities.Capabilities {
	c, err := capabilities.Parse([]byte(src))
	s.Require().NoError(err)
	return c
}

func (s *IndexTestSuite) TestCheckCapabilitiesNilGrantsEverything() {
	idx := s.indexFromSource("caps.sentrie", capabilitiesTestPolicy)
	s.NoError(idx.CheckCapabilities(s.T().Context(), nil))
	s.NoError(idx.CheckCapabilities(s.T().Context(), s.mustParseCapabilities(`{"version": 1}`)))
}

func (s *IndexTestSuite) TestCheckCapabilitiesReportsEveryDenialWithSpan() {
	idx := s.indexFromSource("caps.sentrie", capabilitiesTestPolicy)
	grant := s.mustParseCapabilities(`{
  "version": 1,
  "builtins": [],
  "modules": ["@sentrie/hash"],
  "local_modules": false
}`)
	err := idx.CheckCapabilities(s.T().Context(), grant)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.True(errors.Is(err, xerr.ErrIndex))

	lines := strings.Split(err.Error(), "\n")
	s.Require().Len(lines, 6, err.Error())
	s.Contains(lines[0], `caps.sentrie:3:`)
	s.Contains(lines[0], `builtin "now": not granted by the capabilities file`)
	s.Contains(lines[1], `caps.sentrie:4:`)
	s.Contains(lines[1], `module "@sentrie/uuid"`)
	s.Contains(lines[2], `caps.sentrie:6:`)
	s.Contains(lines[2], `local module "./helper"`)
	s.Contains(lines[3], `caps.sentrie:7:`)
	s.Contains(lines[3], `local module "@local/lib/other"`)
	// count(1) in rule shadowed is denied even though a let named count is in
	// scope: the runtime's getTarget never resolves a call to a local binding
	// (#161), so it would call the builtin. The walk matches that (#200).
	s.Contains(lines[4], `caps.sentrie:10:`)
	s.Contains(lines[4], `builtin "count"`)
	s.Contains(lines[5], `caps.sentrie:12:`)
	s.Contains(lines[5], `builtin "count"`)
}

func (s *IndexTestSuite) TestCheckCapabilitiesAllowsGrantedUses() {
	idx := s.indexFromSource("caps.sentrie", capabilitiesTestPolicy)
	grant := s.mustParseCapabilities(`{
  "version": 1,
  "builtins": ["now", "count"],
  "modules": ["@sentrie/hash", "@sentrie/uuid"],
  "local_modules": true
}`)
	s.NoError(idx.CheckCapabilities(s.T().Context(), grant))
}

func (s *IndexTestSuite) TestCheckCapabilitiesCoversDerivesAndCancellation() {
	idx := s.indexFromSource("caps.sentrie", `namespace com/ex
derive size = (xs: list[number]): number => { yield count(xs) }
policy pol {
  rule ok = { yield size([1]) > 0 }
  export decision of ok
}`)
	err := idx.CheckCapabilities(s.T().Context(), s.mustParseCapabilities(`{"version": 1, "builtins": []}`))
	s.Require().Error(err)
	s.Contains(err.Error(), `caps.sentrie:2:`)
	s.Contains(err.Error(), `builtin "count"`)

	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()
	s.Error(idx.CheckCapabilities(ctx, s.mustParseCapabilities(`{"version": 1}`)))
}
