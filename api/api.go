// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/sentrie-sh/sentrie/runtime"
)

// API is the main entry point for the Sentrie API.
// It is used to execute policies and rules.
type API struct {
	Executor runtime.Executor
}

// NewAPI creates a new API instance.
func NewAPI(executor runtime.Executor) *API {
	return &API{Executor: executor}
}
