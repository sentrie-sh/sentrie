// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
)

func invokeTestBuiltin(ctx context.Context, site *CallSite, name string, args ...box.Value) (box.Value, error) {
	decl := builtins.Table[name]
	if decl == nil {
		return box.Undefined(), nil
	}
	handled, v, err := decl.Precheck(site, args)
	if handled || err != nil {
		return v, err
	}
	return decl.Impl(ctx, site, args...)
}
