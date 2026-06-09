// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "strings"

// FlattenSlashIdentChain returns identifier segments if e is an expression tree of
// only Identifier and InfixExpression with operator "/", left-associative.
// Examples: a/b -> [a,b]; com/example/x -> [com,example,x]. Non-chain returns ok false.
func FlattenSlashIdentChain(e Expression) (parts []string, ok bool) {
	switch v := e.(type) {
	case *Identifier:
		return []string{v.Value}, true
	case *InfixExpression:
		if v.Operator != "/" {
			return nil, false
		}
		left, okL := FlattenSlashIdentChain(v.Left)
		if !okL {
			return nil, false
		}
		right, okR := FlattenSlashIdentChain(v.Right)
		if !okR || len(right) != 1 {
			return nil, false
		}
		return append(left, right[0]), true
	default:
		return nil, false
	}
}

// SlashCalleeFQNS returns "a/b/c" for a slash chain, or "" if not a pure slash-ident chain.
func SlashCalleeFQNS(e Expression) string {
	parts, ok := FlattenSlashIdentChain(e)
	if !ok || len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, FQNSeparator)
}
