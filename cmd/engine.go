// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"
	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/pack"
)

const skipEngineCheckFlagName = "skip-engine-check"

// engineVersion is the version of the running binary, set once by main from
// the build's version (#157). Empty means an unversioned or dev build.
//
//nolint:gochecknoglobals
var engineVersion atomic.Value

// SetEngineVersion records the running binary's version for the pack's
// `[engine]` constraint check. main calls it before Setup; an empty or
// unparseable version marks a dev build, which satisfies every constraint.
func SetEngineVersion(v string) {
	engineVersion.Store(v)
}

func runningEngineVersion() string {
	v, _ := engineVersion.Load().(string)
	return v
}

// skipEngineCheckFlag is the --skip-engine-check flag shared by exec, serve and validate.
func skipEngineCheckFlag() cling.CmdFlag {
	return cling.NewBoolCmdInput(skipEngineCheckFlagName).
		WithDefault(false).
		WithDescription("Do not check the pack's [engine] version constraint against this binary").
		AsFlag()
}

// checkEngine compares the pack's `[engine] sentrie` constraint with the
// running binary. Comparison is strict semver: a prerelease build such as
// 0.0.5-rc1 does not satisfy ">=0.0.5" but does satisfy ">=0.0.4". A mismatch is an error naming the
// constraint and the binary version. A dev or unversioned binary passes with a
// warning written to w. A pack without an `[engine]` table, and skip, pass
// silently.
func checkEngine(p *pack.PackFile, skip bool, w io.Writer) error {
	if skip || p == nil || p.Engine == nil || p.Engine.Sentrie == nil {
		return nil
	}
	constraint := p.Engine.Sentrie
	raw := strings.TrimSpace(runningEngineVersion())
	if raw == "" {
		_, err := fmt.Fprintf(w, "warning: pack requires sentrie %s; this is a dev or unversioned build, so the engine version is not checked\n", constraint)
		return err
	}
	running, err := semver.NewVersion(raw)
	if err != nil {
		_, werr := fmt.Fprintf(w, "warning: pack requires sentrie %s; the binary version %q is not a semantic version, so the engine version is not checked\n", constraint, raw)
		return werr
	}
	if isVCSBuildVersion(running) {
		_, werr := fmt.Fprintf(w, "warning: pack requires sentrie %s; this binary was built from source (%s), so the engine version is not checked\n", constraint, running)
		return werr
	}
	// Masterminds drops a prerelease unless the constraint names one; strict
	// semver orders 0.0.5-rc1 below 0.0.5 and above 0.0.4, so include them.
	strict := *constraint
	strict.IncludePrerelease = true
	if !strict.Check(running) {
		return fmt.Errorf("pack requires sentrie %s, running %s (use --%s to load it anyway)", constraint, running, skipEngineCheckFlagName)
	}
	return nil
}

// pseudoVersionPrerelease matches the prerelease of a Go pseudo-version, the
// version `go build` stamps on a binary built from a VCS checkout, with a
// reachable tag (vX.Y.Z-0.<time>-<hash>) or without one (v0.0.0-<time>-<hash>).
var pseudoVersionPrerelease = regexp.MustCompile(`(^|\.)(0\.)?\d{14}-[0-9a-f]{12}$`)

// isVCSBuildVersion reports whether v is a source build rather than a release:
// a Go pseudo-version, or a version marked "+dirty" by the Go toolchain.
func isVCSBuildVersion(v *semver.Version) bool {
	return pseudoVersionPrerelease.MatchString(v.Prerelease()) || strings.Contains(v.Metadata(), "dirty")
}
