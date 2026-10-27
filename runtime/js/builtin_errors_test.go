// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/constants"
)

// testExecutionStartUnix is the execution start time the builtin tests pin,
// standing in for the global runtime.ModuleBinding.Call sets.
const testExecutionStartUnix int64 = 1_700_000_000

// runBuiltinScript installs provider as the global `m` in a fresh VM, with
// the execution start time set, and runs script against it.
func (s *JSTestSuite) runBuiltinScript(provider ModuleProvider, script string) (goja.Value, error) {
	vm := goja.New()
	s.Require().NoError(vm.Set(constants.ExecutionStartTimeUnixKey, testExecutionStartUnix))
	ex, err := provider(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("m", ex))
	return vm.RunString(script)
}

// TestBuiltinFailuresThrow checks, per builtin module, that a failure is a
// catchable JavaScript exception and never a returned value. uuid has no
// reachable failure path, and crypto, math and string report none today.
func (s *JSTestSuite) TestBuiltinFailuresThrow() {
	cases := []struct {
		module   string
		provider ModuleProvider
		expr     string
		msg      string
	}{
		{"collection", BuiltinCollectionGo, `m.map_keys()`, "requires exactly 1 argument"},
		{"collection", BuiltinCollectionGo, `m.list_sort(1)`, "list_sort requires an array"},
		{"collection", BuiltinCollectionGo, `m.list_includes("abc", "a")`, "list_includes requires an array"},
		{"collection", BuiltinCollectionGo, `m.list_indexOf(null, 1)`, "list_indexOf requires an array"},
		{"collection", BuiltinCollectionGo, `m.map_has([], "k")`, "map_has requires a map"},
		{"collection", BuiltinCollectionGo, `m.map_get(1, "k", "fallback")`, "map_get requires a map"},
		{"collection", BuiltinCollectionGo, `m.list_chunk([1, 2], 0)`, "size must be positive"},
		{"net", BuiltinNetGo, `m.cidrContains("not-a-cidr", "1.2.3.4")`, "invalid CIDR address"},
		{"net", BuiltinNetGo, `m.cidrContains("10.0.0.0/8", "nope")`, "invalid IP or CIDR"},
		{"net", BuiltinNetGo, `m.cidrIntersects("10.0.0.0/8")`, "requires exactly 2 arguments"},
		{"net", BuiltinNetGo, `m.cidrMerge(["10.0.0.0/8", 1])`, "requires an array of IPs/CIDRs"},
		{"semver", BuiltinSemverGo, `m.compare("not-a-version", "1.0.0")`, "invalid semantic version"},
		{"semver", BuiltinSemverGo, `m.satisfies("1.0.0", "~~bad")`, "improper constraint: ~~bad"},
		{"jwt", BuiltinJwtGo, `m.decode("x.y.z", "secret")`, "illegal base64 data"},
		{"jwt", BuiltinJwtGo, `m.decode("not-a-token", "secret")`, "invalid token format"},
		{"jwt", BuiltinJwtGo, `m.getPayload("a.b")`, "invalid token format"},
		{"jwt", BuiltinJwtGo, `m.verify("a.b.c", "secret", "none")`, `algorithm "none" is not allowed`},
		{"url", BuiltinUrlGo, `m.parse("://missing-scheme")`, "missing protocol scheme"},
		{"url", BuiltinUrlGo, `m.join()`, "requires at least 1 argument"},
		{"encoding", BuiltinEncodingGo, `m.base64Decode("!!!")`, "illegal base64 data"},
		{"encoding", BuiltinEncodingGo, `m.hexDecode("zz")`, "invalid byte"},
		{"encoding", BuiltinEncodingGo, `m.urlDecode("%zz")`, "invalid URL escape"},
		{"time", BuiltinTimeGo, `m.parse("yesterday")`, "cannot parse"},
		{"time", BuiltinTimeGo, `m.addDuration(0, "forever")`, "invalid duration"},
		{"time", BuiltinTimeGo, `m.now(1)`, "requires exactly 0 arguments"},
		{"regex", BuiltinRegexGo, `m.match("(", "x")`, "missing closing )"},
		{"hash", BuiltinHashGo, `m.hmac("md4", "data", "key")`, "unsupported algorithm: md4"},
		{"hash", BuiltinHashGo, `m.sha256()`, "requires exactly 1 argument"},
		{"json", BuiltinJsonGo, `m.isValid()`, "requires exactly 1 argument"},
	}

	for _, tc := range cases {
		s.Run(tc.module+"/"+tc.expr, func() {
			caught, err := s.runBuiltinScript(tc.provider, `(function () {
				try {
					const r = `+tc.expr+`;
					return "returned:" + typeof r;
				} catch (e) {
					return "threw:" + e.message;
				}
			})()`)
			s.Require().NoError(err)
			got := caught.String()
			s.Truef(strings.HasPrefix(got, "threw:"), "expected %s to throw, got %q", tc.expr, got)
			s.Contains(got, tc.msg)

			// Uncaught, the failure must surface as an exception, not a value.
			_, err = s.runBuiltinScript(tc.provider, tc.expr)
			var exc *goja.Exception
			s.Require().ErrorAs(err, &exc)
		})
	}
}

// TestBuiltinPredicatesAndNoResult pins the two non-throwing conventions:
// predicates answer false for input they cannot parse, and a genuine "no
// result" answer is null.
func (s *JSTestSuite) TestBuiltinPredicatesAndNoResult() {
	cases := []struct {
		provider ModuleProvider
		expr     string
		want     string
	}{
		{BuiltinNetGo, `m.cidrIsValid("not-a-cidr")`, "false"},
		{BuiltinNetGo, `m.isPrivate("not-an-ip")`, "false"},
		{BuiltinNetGo, `m.isPublic("not-an-ip")`, "false"},
		{BuiltinNetGo, `m.parseIP("not-an-ip") === null`, "true"},
		{BuiltinJsonGo, `m.isValid("{")`, "false"},
		{BuiltinSemverGo, `m.isValid("not-a-version")`, "false"},
		{BuiltinSemverGo, `m.prerelease("1.0.0") === null`, "true"},
		{BuiltinUrlGo, `m.isValid("")`, "false"},
		{BuiltinRegexGo, `m.find("a", "bbb") === null`, "true"},
		{BuiltinJwtGo, `m.verify("x.y.z", "secret")`, "false"},
		{BuiltinCollectionGo, `m.map_get({}, "missing", "fallback")`, "fallback"},
	}
	for _, tc := range cases {
		s.Run(tc.expr, func() {
			got, err := s.runBuiltinScript(tc.provider, tc.expr)
			s.Require().NoError(err)
			s.Equal(tc.want, got.String())
		})
	}
}
