// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package tokens

import "testing"

func TestIsKeywordRecognizesKnownAndUnknownValues(t *testing.T) {
	kind, ok := IsKeyword("transform")
	if !ok {
		t.Fatalf("expected transform to be recognized as keyword")
	}
	if kind != KeywordTransform {
		t.Fatalf("expected transform keyword kind, got %s", kind)
	}

	if _, ok = IsKeyword("transformer"); ok {
		t.Fatalf("did not expect non-keyword value to be recognized")
	}
}

func TestKindStringReturnsUnderlyingValue(t *testing.T) {
	if got := TokenPipeForward.String(); got != "PipeForward" {
		t.Fatalf("unexpected kind string for pipe forward: %q", got)
	}

	if got := KeywordContains.String(); got != "contains" {
		t.Fatalf("unexpected kind string for keyword: %q", got)
	}
}
