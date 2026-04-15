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

package parser

import (
	"github.com/sentrie-sh/sentrie/lexer"
	"github.com/sentrie-sh/sentrie/tokens"
)

// tryReadLambdaSignature reads ( paramList ) => from the lexer and returns param names.
// On success, tokens through the fat arrow are consumed. On failure, all tokens read are pushed back.
func tryReadLambdaSignature(lex *lexer.Lexer) (params []string, ok bool) {
	var buf []tokens.Instance
	read := func() tokens.Instance {
		t := lex.NextToken()
		buf = append(buf, t)
		return t
	}
	undo := func() {
		for i := len(buf) - 1; i >= 0; i-- {
			lex.PushBack(buf[i])
		}
		buf = buf[:0]
	}

	t := read()
	if t.Kind == tokens.PunctRightParentheses {
		t2 := read()
		if t2.Kind == tokens.TokenFatArrow {
			return []string{}, true
		}
		undo()
		return nil, false
	}

	if t.Kind != tokens.Ident {
		undo()
		return nil, false
	}
	names := []string{t.Value}
	for {
		t = read()
		if t.Kind == tokens.PunctRightParentheses {
			t2 := read()
			if t2.Kind == tokens.TokenFatArrow {
				return names, true
			}
			undo()
			return nil, false
		}
		if t.Kind != tokens.PunctComma {
			undo()
			return nil, false
		}
		t = read()
		if t.Kind != tokens.Ident {
			undo()
			return nil, false
		}
		names = append(names, t.Value)
	}
}
