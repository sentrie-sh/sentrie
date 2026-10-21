// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"errors"
	"strings"
	"testing"
)

func TestErrCapability(t *testing.T) {
	t.Parallel()
	err := ErrCapability(testBuiltinValidateRange(), `builtin "now" is not granted`)
	if !errors.Is(err, ErrNotGranted) {
		t.Fatal("expected ErrNotGranted")
	}
	if !errors.Is(err, ErrIndex) {
		t.Fatal("expected ErrIndex")
	}
	if !strings.HasPrefix(err.Error(), `p.sentrie:`) || !strings.HasSuffix(err.Error(), `builtin "now" is not granted: not granted by the capabilities file`) {
		t.Fatal(err.Error())
	}
}
