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

package lexer

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/tokens"
)

type LexerError struct {
	Position tokens.Pos
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("at %s", e.Position)
}

func UnterminatedStringError(pos tokens.Pos) error {
	return errors.Wrap(&LexerError{Position: pos}, "unterminated string literal")
}

func InvalidHereDocSyntaxError(pos tokens.Pos) error {
	return errors.Wrap(&LexerError{Position: pos}, "invalid heredoc syntax")
}
