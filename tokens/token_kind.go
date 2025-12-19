// SPDX-License-Identifier: Apache-2.0

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

package tokens

type Kind string

const (
	EOF     Kind = "EOF"
	Error   Kind = "Error"
	Unknown Kind = "Unknown"

	Whitespace Kind = "Whitespace"

	// Literals
	Ident  Kind = "Ident"
	String Kind = "String"
	Int    Kind = "Int"
	Float  Kind = "Float"
	Bool   Kind = "Bool"

	// Keywords
	KeywordNull Kind = "null"

	KeywordNamespace Kind = "namespace"
	KeywordPolicy    Kind = "policy"
	KeywordLet       Kind = "let"
	KeywordRule      Kind = "rule"
	KeywordFact      Kind = "fact"
	KeywordExport    Kind = "export"
	KeywordDecision  Kind = "decision"
	KeywordOf        Kind = "of"
	KeywordAttach    Kind = "attach"
	KeywordUse       Kind = "use"
	KeywordShape     Kind = "shape"
	KeywordFrom      Kind = "from"
	KeywordAs        Kind = "as"
	KeywordWith      Kind = "with"
	KeywordImport    Kind = "import"
	KeywordWhen      Kind = "when"
	KeywordDefault   Kind = "default"
	KeywordAnd       Kind = "and"
	KeywordCast      Kind = "cast"
	KeywordOr        Kind = "or"
	KeywordXor       Kind = "xor"
	KeywordNot       Kind = "not"
	KeywordIn        Kind = "in"
	KeywordIs        Kind = "is"
	KeywordMatches   Kind = "matches"
	KeywordContains  Kind = "contains"
	KeywordAny       Kind = "any"
	KeywordAll       Kind = "all"
	KeywordFilter    Kind = "filter"
	KeywordFirst     Kind = "first"
	KeywordDistinct  Kind = "distinct"
	KeywordReduce    Kind = "reduce"
	KeywordDefined   Kind = "defined"
	KeywordEmpty     Kind = "empty"
	KeywordYield     Kind = "yield"
	KeywordTransform Kind = "transform"

	KeywordTrue    Kind = "true"
	KeywordFalse   Kind = "false"
	KeywordUnknown Kind = "unknown"

	KeywordString   Kind = "string"
	KeywordNumber   Kind = "number"
	KeywordBoolean  Kind = "boolean"
	KeywordTrinary  Kind = "trinary"
	KeywordList     Kind = "list"
	KeywordMap      Kind = "map"
	KeywordRecord   Kind = "record"
	KeywordDocument Kind = "document"

	// Operators
	TokenAssign    Kind = "Assign"
	TokenEq        Kind = "Equals"
	TokenNeq       Kind = "NotEquals"
	TokenLte       Kind = "LessThanOrEqual"
	TokenGte       Kind = "GreaterThanOrEqual"
	TokenLt        Kind = "LessThan"
	TokenGt        Kind = "GreaterThan"
	TokenPlus      Kind = "Plus"
	TokenMinus     Kind = "Minus"
	TokenMul       Kind = "Multiply"
	TokenDiv       Kind = "Divide"
	TokenMod       Kind = "Modulo"
	TokenQuestion  Kind = "Question"
	PunctColon     Kind = "Colon"
	TokenBang      Kind = "Bang"
	TokenDot       Kind = "Dot"
	TokenDotDotDot Kind = "DotDotDot"
	TokenAt        Kind = "At"

	// Punctuation
	PunctComma            Kind = "Comma"
	PunctSemicolon        Kind = "Semicolon"
	PunctLeftParentheses  Kind = "LeftParen"
	PunctRightParentheses Kind = "RightParen"
	PunctLeftCurly        Kind = "LeftBrace"
	PunctRightCurly       Kind = "RightBrace"
	PunctLeftBracket      Kind = "LeftBracket"
	PunctRightBracket     Kind = "RightBracket"

	// Comments
	LineComment     Kind = "LineComment"
	TrailingComment Kind = "TrailingComment"
)

func IsKeyword(str string) (Kind, bool) {
	kind, exists := keywords[str]
	return kind, exists
}

// Keywords map for fast lookup
var keywords = map[string]Kind{
	"decision":  KeywordDecision,
	"yield":     KeywordYield,
	"transform": KeywordTransform,
	"shape":     KeywordShape,
	"of":        KeywordOf,
	"attach":    KeywordAttach,
	"namespace": KeywordNamespace,
	"policy":    KeywordPolicy,
	"let":       KeywordLet,
	"rule":      KeywordRule,
	"when":      KeywordWhen,
	"default":   KeywordDefault,
	"fact":      KeywordFact,
	"export":    KeywordExport,
	"use":       KeywordUse,
	"cast":      KeywordCast,
	"from":      KeywordFrom,
	"as":        KeywordAs,
	"with":      KeywordWith,
	"import":    KeywordImport,
	"and":       KeywordAnd,
	"or":        KeywordOr,
	"xor":       KeywordXor,
	"not":       KeywordNot,
	"in":        KeywordIn,
	"is":        KeywordIs,
	"matches":   KeywordMatches,
	"contains":  KeywordContains,
	"any":       KeywordAny,
	"all":       KeywordAll,
	"filter":    KeywordFilter,
	"first":     KeywordFirst,
	"distinct":  KeywordDistinct,
	"reduce":    KeywordReduce,
	"defined":   KeywordDefined,
	"empty":     KeywordEmpty,

	"true":    KeywordTrue,
	"false":   KeywordFalse,
	"unknown": KeywordUnknown,

	"null": KeywordNull,

	"string":   KeywordString,
	"number":   KeywordNumber,
	"boolean":  KeywordBoolean,
	"trinary":  KeywordTrinary,
	"list":     KeywordList,
	"map":      KeywordMap,
	"record":   KeywordRecord,
	"document": KeywordDocument,
}

func (k Kind) String() string {
	return string(k)
}
