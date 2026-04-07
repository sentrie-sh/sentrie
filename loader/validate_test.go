// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package loader

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/pack"
)

func TestValidatePackFileMarshalFailure(t *testing.T) {
	p := &pack.PackFile{
		SchemaVersion: &pack.SentrieSchema{Version: 1},
		Pack:          &pack.PackInformation{Name: "ok"},
		Metadata: map[string]any{
			"bad": make(chan int),
		},
	}

	err := ValidatePackFile(p)
	if err == nil {
		t.Fatal("expected marshal failure")
	}
	if !strings.Contains(err.Error(), "failed to marshal pack file to JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}
