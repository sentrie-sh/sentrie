// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"errors"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

func testBuiltinValidateRange() tokens.Range {
	return tokens.Range{
		File: "p.sentrie",
		From: tokens.Pos{Line: 2, Column: 4, Offset: 4},
		To:   tokens.Pos{Line: 2, Column: 4, Offset: 8},
	}
}

func TestErrBuiltinArgKind(t *testing.T) {
	t.Parallel()
	r := testBuiltinValidateRange()
	err := ErrBuiltinArgKind(r, "filter: first argument must be a list")
	require := func(t *testing.T, cond bool, msg string) {
		t.Helper()
		if !cond {
			t.Fatal(msg)
		}
	}
	require(t, errors.Is(err, ErrIndex), "unwrap ErrIndex")
	require(t, err.Error() == "p.sentrie:3:4: filter: first argument must be a list: index error", err.Error())
}

func TestErrBuiltinCallableArity(t *testing.T) {
	t.Parallel()
	r := testBuiltinValidateRange()
	err := ErrBuiltinCallableArity(r, "reduce: reducer must have arity 2 or 3")
	if !errors.Is(err, ErrIndex) {
		t.Fatal("expected ErrIndex")
	}
	if err.Error() != "p.sentrie:3:4: reduce: reducer must have arity 2 or 3: index error" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestErrBuiltinCallArity(t *testing.T) {
	t.Parallel()
	r := testBuiltinValidateRange()
	err := ErrBuiltinCallArity(r, "now requires 0 arguments")
	if !errors.Is(err, ErrIndex) {
		t.Fatal("expected ErrIndex")
	}
	if err.Error() != "p.sentrie:3:4: now requires 0 arguments: index error" {
		t.Fatalf("got %q", err.Error())
	}
}
