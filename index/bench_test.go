// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"path/filepath"
	"testing"

	"github.com/sentrie-sh/sentrie/loader"
)

// BenchmarkBuildIndexExamplePack builds and validates the index of
// example_pack. Parsing happens outside the timed region, so this measures the
// index alone; each iteration gets freshly parsed programs so no iteration
// sees state left on the AST by an earlier one.
func BenchmarkBuildIndexExamplePack(b *testing.B) {
	ctx := b.Context()
	packFile, err := loader.LoadPack(ctx, filepath.Join("..", "example_pack"))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		programs, err := loader.LoadPrograms(ctx, packFile)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		idx := CreateIndex()
		if err := idx.SetPack(ctx, packFile); err != nil {
			b.Fatal(err)
		}
		for _, program := range programs {
			if err := idx.AddProgram(ctx, program); err != nil {
				b.Fatal(err)
			}
		}
		if err := idx.Validate(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
