// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"strconv"

	"github.com/dop251/goja"
)

func (s *JSTestSuite) TestCryptoSha256KnownAnswers() {
	vm := goja.New()
	ex, err := BuiltinCryptoGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("crypto", ex))

	cases := map[string]string{
		"":    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"abc": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
	}
	for in, want := range cases {
		got, err := vm.RunString(`crypto.sha256(` + strconv.Quote(in) + `)`)
		s.Require().NoError(err)
		s.Equal(want, got.Export(), "sha256(%q)", in)
	}
}

func (s *JSTestSuite) TestCryptoSha256MatchesHashModule() {
	vm := goja.New()
	cryptoEx, err := BuiltinCryptoGo(vm)
	s.Require().NoError(err)
	hashEx, err := BuiltinHashGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("crypto", cryptoEx))
	s.Require().NoError(vm.Set("hash", hashEx))

	got, err := vm.RunString(`crypto.sha256("secret") === hash.sha256("secret")`)
	s.Require().NoError(err)
	s.True(got.ToBoolean())

	// The digest must not carry the plaintext input.
	out, err := vm.RunString(`crypto.sha256("secret")`)
	s.Require().NoError(err)
	s.NotContains(out.String(), "secret")
	s.Len(out.String(), 64)
}

func (s *JSTestSuite) TestCryptoSha256ArityThrows() {
	vm := goja.New()
	ex, err := BuiltinCryptoGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("crypto", ex))

	got, err := vm.RunString(`(() => { try { crypto.sha256(); return "returned"; } catch (e) { return "threw"; } })()`)
	s.Require().NoError(err)
	s.Equal("threw", got.String())
}
