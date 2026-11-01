// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkParseExamplePack parses every example_pack source into a program.
func BenchmarkParseExamplePack(b *testing.B) {
	paths, err := filepath.Glob(filepath.Join("..", "example_pack", "*.sentrie"))
	if err != nil || len(paths) == 0 {
		b.Fatalf("no example_pack sources found: %v", err)
	}
	sources := make(map[string][]byte, len(paths))
	var total int64
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			b.Fatal(err)
		}
		sources[p] = content
		total += int64(len(content))
	}
	b.SetBytes(total)
	b.ReportAllocs()
	for b.Loop() {
		for name, src := range sources {
			if _, err := NewParser(bytes.NewReader(src), name).ParseProgram(b.Context()); err != nil {
				b.Fatalf("parse %s: %v", name, err)
			}
		}
	}
}
