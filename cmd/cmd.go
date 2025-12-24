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

package cmd

import (
	"context"
	"log/slog"

	"github.com/binaek/cling"
)

func Setup(ctx context.Context, version string) *cling.CLI {
	cli := cling.NewCLI("sentrie", version).
		WithDescription("Sentrie is the policy enforcement engine").
		WithPreRun(func(ctx context.Context, args []string) error {
			slog.DebugContext(ctx, "==> Starting Sentrie", slog.String("version", version))
			return nil
		}).
		WithPostRun(func(ctx context.Context, args []string) error {
			slog.DebugContext(ctx, "==> Exiting Sentrie")
			return nil
		})

	addServeCmd(cli)
	addInitCmd(cli)
	addExecCmd(cli)
	addValidateCmd(cli)

	return cli
}

func Execute(ctx context.Context, cli *cling.CLI, args []string) error {
	if cli == nil {
		panic("CLI cannot be NIL")
	}
	return cli.Run(ctx, args)
}
