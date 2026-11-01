// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
)

// loadExamplePackIndex runs the load, index and validate steps for
// example_pack.
func loadExamplePackIndex(ctx context.Context, tb testing.TB) *index.Index {
	tb.Helper()
	packFile, err := loader.LoadPack(ctx, examplePackDir())
	if err != nil {
		tb.Fatal(err)
	}
	programs, err := loader.LoadPrograms(ctx, packFile)
	if err != nil {
		tb.Fatal(err)
	}
	idx := index.CreateIndex()
	if err := idx.SetPack(ctx, packFile); err != nil {
		tb.Fatal(err)
	}
	for _, program := range programs {
		if err := idx.AddProgram(ctx, program); err != nil {
			tb.Fatal(err)
		}
	}
	if err := idx.Validate(ctx); err != nil {
		tb.Fatal(err)
	}
	return idx
}

// BenchmarkStartupExamplePack measures what a process pays before it can
// serve its first decision: load the pack, parse, index, validate and build
// the executor (#140).
func BenchmarkStartupExamplePack(b *testing.B) {
	ctx := b.Context()
	b.ReportAllocs()
	for b.Loop() {
		idx := loadExamplePackIndex(ctx, b)
		exec, err := NewExecutor(idx)
		if err != nil {
			b.Fatal(err)
		}
		if err := exec.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExecutorExamplePack evaluates example_pack decisions on an
// executor built once, outside the timed loop.
func BenchmarkExecutorExamplePack(b *testing.B) {
	ctx := b.Context()
	exec, err := NewExecutor(loadExamplePackIndex(ctx, b))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = exec.Close() })
	facts := map[string]any{"user": map[string]any{"role": "admin", "status": "active"}}
	// One untimed evaluation so that a single-iteration run of "rule" does not
	// carry the executor's first-call setup.
	if _, err := exec.ExecRule(ctx, "user_management", "user_access", "allow_user", facts); err != nil {
		b.Fatal(err)
	}

	b.Run("rule", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := exec.ExecRule(ctx, "user_management", "user_access", "allow_user", facts); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("policy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := exec.ExecPolicy(ctx, "user_management", "user_access", facts); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("pipeline", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := exec.ExecPolicy(ctx, "sh/sentrie/example/pipeline", "basics", nil); err != nil {
				b.Fatal(err)
			}
		}
	})
}
