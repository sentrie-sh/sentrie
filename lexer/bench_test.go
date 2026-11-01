// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

// examplePackSources reads every .sentrie file of example_pack once, outside
// any timed loop.
func examplePackSources(tb testing.TB) map[string][]byte {
	tb.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "example_pack", "*.sentrie"))
	if err != nil || len(paths) == 0 {
		tb.Fatalf("no example_pack sources found: %v", err)
	}
	sources := make(map[string][]byte, len(paths))
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			tb.Fatal(err)
		}
		sources[p] = content
	}
	return sources
}

// BenchmarkLexExamplePack lexes every example_pack source to EOF.
func BenchmarkLexExamplePack(b *testing.B) {
	sources := examplePackSources(b)
	var total int64
	for _, src := range sources {
		total += int64(len(src))
	}
	b.SetBytes(total)
	b.ReportAllocs()
	for b.Loop() {
		for name, src := range sources {
			l := NewLexer(bytes.NewReader(src), name)
			for {
				tok := l.NextToken()
				if tok.Kind == tokens.Error {
					b.Fatalf("lex error in %s", name)
				}
				if tok.Kind == tokens.EOF {
					break
				}
			}
		}
	}
}
