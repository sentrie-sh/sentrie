// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"github.com/dop251/goja"
)

// stringModuleVM returns a VM with @sentrie/string bound to the global `s`.
func (s *JSTestSuite) stringModuleVM() *goja.Runtime {
	vm := goja.New()
	ex, err := BuiltinStringGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("s", ex))
	return vm
}

// stringEval evaluates a script against @sentrie/string and returns the
// exported result.
func (s *JSTestSuite) stringEval(script string) any {
	v, err := s.stringModuleVM().RunString(script)
	s.Require().NoError(err, "script: %s", script)
	return v.Export()
}

// stringEvalThrows asserts that a script throws a JS exception whose message
// contains want.
func (s *JSTestSuite) stringEvalThrows(script, want string) {
	_, err := s.stringModuleVM().RunString(script)
	s.Require().Error(err, "script: %s", script)
	var ex *goja.Exception
	s.Require().ErrorAs(err, &ex, "script: %s", script)
	s.Contains(err.Error(), want, "script: %s", script)
}

func (s *JSTestSuite) TestStringLowerUpper() {
	s.Equal("straße ǆ αβγ", s.stringEval(`s.lower("STRAßE ǅ ΑΒΓ")`))
	s.Equal("ÀÉÎ İ", s.stringEval(`s.upper("àéî İ")`))
	s.Equal("", s.stringEval(`s.lower("")`))
	// Simple case mapping: no one-to-many expansions.
	s.Equal("ß", s.stringEval(`s.lower("ẞ")`))
	s.Equal("ß", s.stringEval(`s.upper("ß")`))
	s.stringEvalThrows(`s.lower(1)`, "lower: argument 1 must be a string, got number")
	s.stringEvalThrows(`s.upper()`, "upper: expected 1 argument(s), got 0")
	s.stringEvalThrows(`s.upper("a", "b")`, "upper: expected 1 argument(s), got 2")
}

func (s *JSTestSuite) TestStringTrim() {
	// U+3000 IDEOGRAPHIC SPACE and U+00A0 NO-BREAK SPACE are Unicode White_Space.
	s.Equal("a b", s.stringEval("s.trim(\"\\u3000\\u00a0 a b \\t\\n\")"))
	s.Equal("a b \t", s.stringEval("s.trimStart(\"\\u3000 a b \\t\")"))
	s.Equal("\u3000 a b", s.stringEval("s.trimEnd(\"\\u3000 a b \\t\\u00a0\")"))
	// Unicode White_Space, not JS trim: NEL is stripped, a BOM is kept.
	s.Equal("\ufeffa", s.stringEval("s.trim(\"\\u0085\\ufeffa\\u0085\")"))
	s.stringEvalThrows(`s.trim(null)`, "trim: argument 1 must be a string, got null")
	s.stringEvalThrows(`s.trimStart(undefined)`, "trimStart: argument 1 must be a string, got undefined")
	s.stringEvalThrows(`s.trimEnd([])`, "trimEnd: argument 1 must be a string, got array")
}

func (s *JSTestSuite) TestStringStartsEndsWithIncludes() {
	s.Equal(true, s.stringEval(`s.startsWith("/api/v1/users", "/api/")`))
	s.Equal(false, s.stringEval(`s.startsWith("/api", "/api/")`))
	s.Equal(true, s.stringEval(`s.startsWith("héllo", "hé")`))
	s.Equal(true, s.stringEval(`s.endsWith("report.pdf", ".pdf")`))
	s.Equal(false, s.stringEval(`s.endsWith("report.pdf", ".PDF")`))
	s.Equal(true, s.stringEval(`s.includes("日本語テキスト", "語テ")`))
	s.Equal(true, s.stringEval(`s.includes("abc", "")`))
	s.Equal(false, s.stringEval(`s.includes("abc", "d")`))
	s.stringEvalThrows(`s.startsWith("a")`, "startsWith: expected 2 argument(s), got 1")
	s.stringEvalThrows(`s.endsWith("a", 1)`, "endsWith: argument 2 must be a string, got number")
	s.stringEvalThrows(`s.includes({}, "a")`, "includes: argument 1 must be a string, got object")
}

func (s *JSTestSuite) TestStringSplit() {
	s.Equal([]any{"a", "b", "", "c"}, s.stringEval(`s.split("a,b,,c", ",")`))
	s.Equal([]any{"abc"}, s.stringEval(`s.split("abc", ",")`))
	// An empty separator splits into code points, so an astral character stays whole.
	s.Equal([]any{"a", "😀", "é"}, s.stringEval(`s.split("a😀é", "")`))
	s.Equal(true, s.stringEval(`Array.isArray(s.split("a b", " "))`))
	s.stringEvalThrows(`s.split("a")`, "split: expected 2 argument(s), got 1")
	s.stringEvalThrows(`s.split("a", null)`, "split: argument 2 must be a string, got null")
}

func (s *JSTestSuite) TestStringJoin() {
	s.Equal("a/b/c", s.stringEval(`s.join(["a", "b", "c"], "/")`))
	s.Equal("", s.stringEval(`s.join([], "/")`))
	s.Equal("ü·😀", s.stringEval(`s.join(["ü", "😀"], "·")`))
	s.stringEvalThrows(`s.join("abc", "")`, "join: argument 1 must be an array of strings, got string")
	s.stringEvalThrows(`s.join(["a", 1], "")`, "join: argument 1 element 1 must be a string, got number")
	s.stringEvalThrows(`s.join(["a"])`, "join: expected 2 argument(s), got 1")
}

func (s *JSTestSuite) TestStringJoinBoundsArrayLength() {
	// A sparse array with a huge length costs nothing in JS; join must refuse
	// it before allocating, not size a buffer from the length.
	s.stringEvalThrows(`var a = []; a.length = 2**32 - 1; s.join(a, ",")`,
		"join: argument 1 has 4294967295 elements, more than the limit of 1048576")
	// Within the limit, a missing element throws at that element.
	s.stringEvalThrows(`var a = []; a.length = 1048576; s.join(a, ",")`,
		"join: argument 1 element 0 must be a string, got undefined")
	s.Equal("ab", s.stringEval(`var a = []; a[0] = "a"; a[1] = "b"; s.join(a, "")`))
}

func (s *JSTestSuite) TestStringJoinAcceptsGoSlice() {
	// Lists arriving from policy code are Go slices wrapped by goja, not JS arrays.
	vm := s.stringModuleVM()
	s.Require().NoError(vm.Set("list", []any{"x", "y"}))
	v, err := vm.RunString(`s.join(list, "-")`)
	s.Require().NoError(err)
	s.Equal("x-y", v.Export())
}

func (s *JSTestSuite) TestStringReplace() {
	s.Equal("a-b.c", s.stringEval(`s.replace("a.b.c", ".", "-")`))
	s.Equal("a-b-c", s.stringEval(`s.replaceAll("a.b.c", ".", "-")`))
	s.Equal("café", s.stringEval(`s.replace("cafe", "e", "é")`))
	s.Equal("abc", s.stringEval(`s.replace("abc", "x", "y")`))
	s.Equal("-abc", s.stringEval(`s.replace("abc", "", "-")`))
	s.Equal("-a-b-c-", s.stringEval(`s.replaceAll("abc", "", "-")`))
	s.stringEvalThrows(`s.replace("abc", "a")`, "replace: expected 3 argument(s), got 2")
	s.stringEvalThrows(`s.replaceAll("abc", "a", 1)`, "replaceAll: argument 3 must be a string, got number")
}

func (s *JSTestSuite) TestStringEqualFold() {
	s.Equal(true, s.stringEval(`s.equalFold("Admin", "ADMIN")`))
	s.Equal(true, s.stringEval(`s.equalFold("ΣΑΣ", "σας")`))
	s.Equal(false, s.stringEval(`s.equalFold("admin", "admins")`))
	s.stringEvalThrows(`s.equalFold("a")`, "equalFold: expected 2 argument(s), got 1")
	s.stringEvalThrows(`s.equalFold("a", true)`, "equalFold: argument 2 must be a string, got boolean")
}

func (s *JSTestSuite) TestStringFormat() {
	s.Equal("user alice has 3 roles", s.stringEval(`s.format("user %s has %d roles", "alice", 3)`))
	s.Equal("ratio 0.33 / 1.500000", s.stringEval(`s.format("ratio %.2f / %f", 1/3, 1.5)`))
	// %.Nf rounds the exact binary value, ties to even.
	s.Equal("round 2 4", s.stringEval(`s.format("round %.0f %.0f", 2.5, 3.5)`))
	s.Equal(`id "a\"b"`, s.stringEval(`s.format("id %q", "a\"b")`))
	s.Equal("v: x 1.5 true null 100%", s.stringEval(`s.format("v: %v %v %v %v 100%%", "x", 1.5, true, null)`))
	s.Equal("0", s.stringEval(`s.format("%d", -0)`))
	s.Equal("ünï 😀", s.stringEval(`s.format("ünï %s", "😀")`))
	s.Equal("plain", s.stringEval(`s.format("plain")`))
	// Negative zero prints as 0 under every numeric verb.
	s.Equal("0.000000 0.00 0 0", s.stringEval(`s.format("%f %.2f %d %v", -0, -0, -0, -0)`))
	// A negative value that rounds to zero keeps its sign, as toFixed does.
	s.Equal("-0", s.stringEval(`s.format("%.0f", -0.4)`))
	// Literal text is copied byte for byte, so multi-byte characters survive.
	s.Equal("héllo wörld ✓", s.stringEval(`s.format("héllo %s ✓", "wörld")`))
}

func (s *JSTestSuite) TestStringFormatErrors() {
	cases := []struct{ script, want string }{
		{`s.format()`, "format: expected a format string and its arguments, got no arguments"},
		{`s.format(1)`, "format: argument 1 must be a string, got number"},
		{`s.format("%s")`, "format: missing argument for %s (verb 1)"},
		{`s.format("%s", "a", "b")`, "format: 2 arguments given but the format uses 1"},
		{`s.format("%x", 1)`, "format: unsupported verb %x"},
		{`s.format("50%")`, "format: trailing '%' at the end of the format string"},
		{`s.format("%s", 1)`, "format: %s expects a string, got number"},
		{`s.format("%q", null)`, "format: %q expects a string, got null"},
		{`s.format("%d", "1")`, "format: %d expects a number, got string"},
		{`s.format("%d", 1.5)`, "format: %d expects an integer, got 1.5"},
		{`s.format("%f", NaN)`, "format: %f expects a finite number, got NaN"},
		{`s.format("%d", Infinity)`, "format: %d expects a finite number, got Infinity"},
		{`s.format("%.2d", 1)`, "format: a precision is only supported as %.Nf"},
		{`s.format("%.21f", 1)`, "format: precision in %.21f must be between 0 and 20"},
		{`s.format("%v", [1])`, "format: %v expects a string, number, boolean or null, got array"},
		{`s.format("%v", undefined)`, "format: %v expects a string, number, boolean or null, got undefined"},
		{`s.format("%é", 1)`, "format: unsupported verb %é"},
		{`s.format("%v", 10n)`, "format: %v expects a string, number, boolean or null, got bigint"},
		{`s.format("%s", Symbol("x"))`, "format: %s expects a string, got symbol"},
	}
	for _, tc := range cases {
		s.stringEvalThrows(tc.script, tc.want)
	}
}

func (s *JSTestSuite) TestStringErrorsAreCatchable() {
	// Errors are thrown, never returned, so a try/catch sees them.
	s.Equal("caught", s.stringEval(`(() => { try { s.lower(1); return "returned" } catch (e) { return "caught" } })()`))
}
