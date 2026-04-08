// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package lexer

import (
	"errors"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func TestLexerErrorHelpersWrapLexerError(t *testing.T) {
	pos := tokens.Pos{Line: 4, Column: 2}
	unterminated := UnterminatedStringError("x.sentra", pos)
	invalid := InvalidHereDocSyntaxError("x.sentra", pos)

	var lexErr *LexerError
	require.True(t, errors.As(unterminated, &lexErr))
	require.True(t, errors.As(invalid, &lexErr))
	require.Contains(t, unterminated.Error(), "unterminated string literal")
	require.Contains(t, invalid.Error(), "invalid heredoc syntax")
}

func TestReadHereDocReportsNewWrappedSyntaxBranches(t *testing.T) {
	t.Run("missing tag after marker", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<\n"), "bad.sentra")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "heredoc requires identifier tag after <<<")
		var lexErr *LexerError
		require.True(t, errors.As(err, &lexErr))
	})

	t.Run("invalid tag characters", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<1tag\n"), "bad.sentra")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "heredoc requires identifier tag after <<<")
	})

	t.Run("trailing non whitespace after tag", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<TAG!\nTAG\n"), "bad.sentra")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected characters after heredoc tag")
		var lexErr *LexerError
		require.True(t, errors.As(err, &lexErr))
	})

	t.Run("unicode tag fails ASCII identifier regex", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<café\nbody\ncafé\n"), "uni.sentra")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid heredoc tag")
	})
}
