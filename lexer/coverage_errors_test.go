// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
	unterminated := UnterminatedStringError("x.sentrie", pos)
	invalid := InvalidHereDocSyntaxError("x.sentrie", pos)

	var lexErr *LexerError
	require.True(t, errors.As(unterminated, &lexErr))
	require.True(t, errors.As(invalid, &lexErr))
	require.Contains(t, unterminated.Error(), "unterminated string literal")
	require.Contains(t, invalid.Error(), "invalid heredoc syntax")
}

func TestReadHereDocReportsNewWrappedSyntaxBranches(t *testing.T) {
	t.Run("missing tag after marker", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<\n"), "bad.sentrie")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "heredoc requires identifier tag after <<<")
		var lexErr *LexerError
		require.True(t, errors.As(err, &lexErr))
	})

	t.Run("invalid tag characters", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<1tag\n"), "bad.sentrie")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "heredoc requires identifier tag after <<<")
	})

	t.Run("trailing non whitespace after tag", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<TAG!\nTAG\n"), "bad.sentrie")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected characters after heredoc tag")
		var lexErr *LexerError
		require.True(t, errors.As(err, &lexErr))
	})

	t.Run("unicode tag fails ASCII identifier regex", func(t *testing.T) {
		lx := NewLexer(strings.NewReader("<<<café\nbody\ncafé\n"), "uni.sentrie")
		_, err := lx.readHereDoc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid heredoc tag")
	})
}
