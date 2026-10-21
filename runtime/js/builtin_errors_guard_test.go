// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// errorObjectViolations reports every place in a Go source file where a
// JavaScript Error object is built without being thrown. It is the guard for
// #111: goja treats a returned value as a normal result, so an Error object
// that is returned, stored or wrapped reaches JavaScript as a truthy success.
//
// Two shapes are flagged:
//   - a NewGoError or NewTypeError call whose parent (ignoring parentheses)
//     is anything other than panic(...);
//   - a ToValue call wrapping a Go error expression (errors.New, errors.Join,
//     fmt.Errorf, or an identifier named err).
func errorObjectViolations(filename string, src []byte) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	var stack []ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pos := fset.Position(call.Pos())
		switch sel.Sel.Name {
		case "NewGoError", "NewTypeError":
			if !isPanicArgument(stack) {
				out = append(out, fmt.Sprintf("%s:%d builds an Error object without throwing it; use `return throwErr(vm, err)`", pos.Filename, pos.Line))
			}
		case "ToValue":
			if len(call.Args) == 1 && isGoErrorExpr(call.Args[0]) {
				out = append(out, fmt.Sprintf("%s:%d returns a Go error as a value; use `return throwErr(vm, err)`", pos.Filename, pos.Line))
			}
		}
		return true
	})
	return out, nil
}

// isPanicArgument reports whether the node on top of stack is, ignoring
// parentheses, the argument of a panic call.
func isPanicArgument(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		switch parent := stack[i].(type) {
		case *ast.ParenExpr:
			continue
		case *ast.CallExpr:
			id, ok := parent.Fun.(*ast.Ident)
			return ok && id.Name == "panic"
		default:
			return false
		}
	}
	return false
}

func isGoErrorExpr(e ast.Expr) bool {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			break
		}
		e = p.X
	}
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name == "err"
	case *ast.CallExpr:
		sel, ok := v.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}
		switch pkg.Name + "." + sel.Sel.Name {
		case "errors.New", "errors.Join", "fmt.Errorf":
			return true
		}
	}
	return false
}

// TestBuiltinsNeverReturnErrorObjects scans every non-test Go file in
// runtime/js with errorObjectViolations.
func (s *JSTestSuite) TestBuiltinsNeverReturnErrorObjects() {
	files, err := filepath.Glob("*.go")
	s.Require().NoError(err)
	s.Require().NotEmpty(files)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		s.Require().NoError(err)
		violations, err := errorObjectViolations(file, src)
		s.Require().NoError(err)
		s.Empty(violations)
	}
}

// TestErrorObjectGuardCatchesEverySpelling checks the guard itself, so it
// cannot quietly regress to matching only one spelling of the bug.
func (s *JSTestSuite) TestErrorObjectGuardCatchesEverySpelling() {
	flagged := map[string]string{
		"direct return":      `return vm.NewGoError(err)`,
		"type error return":  `return vm.NewTypeError("bad")`,
		"assign then return": `e := vm.NewGoError(err); return e`,
		"parenthesised":      `return (vm.NewGoError(err))`,
		"ToValue wrap":       `return vm.ToValue(vm.NewGoError(err))`,
		"ToValue errors.New": `return vm.ToValue(errors.New("x"))`,
		"ToValue fmt.Errorf": `return vm.ToValue(fmt.Errorf("x: %w", err))`,
		"ToValue err":        `return vm.ToValue(err)`,
		"field receiver":     `return ar.VM.NewGoError(err)`,
	}
	allowed := map[string]string{
		"throw":               `panic(vm.NewGoError(err))`,
		"parenthesised throw": `panic((vm.NewGoError(err)))`,
		"field receiver":      `panic(ar.VM.NewGoError(err))`,
		"ordinary value":      `return vm.ToValue(true)`,
	}
	wrap := func(body string) []byte {
		return []byte("package p\n\nfunc f() any {\n\t" + body + "\n\treturn nil\n}\n")
	}
	for name, body := range flagged {
		s.Run("flags "+name, func() {
			got, err := errorObjectViolations("snippet.go", wrap(body))
			s.Require().NoError(err)
			s.Len(got, 1, body)
		})
	}
	for name, body := range allowed {
		s.Run("allows "+name, func() {
			got, err := errorObjectViolations("snippet.go", wrap(body))
			s.Require().NoError(err)
			s.Empty(got, body)
		})
	}
}
