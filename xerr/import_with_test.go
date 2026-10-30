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
		{"sorted aliases", "x", map[string]string{"b": "b", "a": "a"}, `import binds unknown fact "x": policy t/q accepts fact aliases: a, b`},
		{"alias hint", "user", map[string]string{"u": "user"}, `import binds unknown fact "user": policy t/q accepts fact aliases: u; fact "user" is bound by its alias "u"`},
		{"no hint when only an alias matches", "u2", map[string]string{"u": "user"}, `import binds unknown fact "u2": policy t/q accepts fact aliases: u`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ImportUnknownFactMessage(c.with, "t/q", c.facts); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
