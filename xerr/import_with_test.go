// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import "testing"

func TestImportUnknownFactMessage(t *testing.T) {
	cases := []struct {
		name  string
		with  string
		facts map[string]string
		want  string
	}{
		{"no facts", "x", nil, `import binds unknown fact "x": policy t/q declares no facts`},
		{"sorted names", "x", map[string]string{"b": "b", "a": "a"}, `import binds unknown fact "x": policy t/q accepts facts: a, b`},
		{"alias hint", "u", map[string]string{"user": "u"}, `import binds unknown fact "u": policy t/q accepts facts: user; "u" is the local alias of fact "user", bind it by its name`},
		{"no hint when nothing matches", "u2", map[string]string{"user": "u"}, `import binds unknown fact "u2": policy t/q accepts facts: user`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ImportUnknownFactMessage(c.with, "t/q", c.facts); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
