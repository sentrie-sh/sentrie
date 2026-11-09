// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
)

var opPrec = map[string]parser.Precedence{
	"or": parser.OR, "xor": parser.XOR, "and": parser.AND,
	"==": parser.EQUALITY, "!=": parser.EQUALITY, "is": parser.EQUALITY,
	"<": parser.COMPARISON, ">": parser.COMPARISON, "<=": parser.COMPARISON, ">=": parser.COMPARISON,
	"in": parser.COMPARISON, "matches": parser.COMPARISON, "contains": parser.COMPARISON,
	"+": parser.SUM, "-": parser.SUM, "*": parser.PRODUCT, "/": parser.PRODUCT, "%": parser.PRODUCT,
}

// negatedInfix holds the operators `x not <op> y` can negate.
var negatedInfix = map[string]bool{"in": true, "matches": true, "contains": true, "not": true}

// precOf reports how tightly e binds as printed, and whether its printed form
// ends with an operand the parser reads as a full expression. A trailing
// comment after an open expression would attach to that last operand, so it
// is printed after parentheses instead.
func precOf(e ast.Expression) (parser.Precedence, bool) {
	switch n := e.(type) {
	case *ast.InfixExpression:
		return opPrec[n.Operator], true
	case *ast.UnaryExpression:
		if isNegatedIs(n) {
			return parser.EQUALITY, true
		}
		return parser.UNARY, true
	case *ast.TernaryExpression:
		return parser.TERNARY, true
	case *ast.CallExpression:
		if n.Piped != nil {
			return parser.PIPELINE, true
		}
		return parser.CALL, false
	case *ast.FieldAccessExpression, *ast.IndexAccessExpression:
		return parser.INDEX, false
	case *ast.CastExpression:
		return parser.UNARY, false
	case *ast.IsDefinedExpression, *ast.IsEmptyExpression:
		return parser.EQUALITY, false
	case *ast.IsExpression:
		return parser.EQUALITY, n.Right != nil
	case *ast.PrecedingCommentExpression:
		return parser.LOWEST, true
	case *ast.TrailingCommentExpression:
		pr, open := precOf(n.Wrap)
		if open {
			pr = parser.PRIMARY
		}
		return pr, true
	}
	return parser.PRIMARY, false
}

// prefixLed reports whether e starts with a prefix operator, which
// the parser reads whatever the surrounding precedence. As the last operand
// of an expression such a node never needs parentheses.
func prefixLed(e ast.Expression) bool {
	switch n := e.(type) {
	case *ast.UnaryExpression:
		return !isNegatedInfix(n) && !isNegatedIs(n)
	}
	return false
}

func isNegatedInfix(u *ast.UnaryExpression) bool {
	in, ok := u.Right.(*ast.InfixExpression)
	return ok && u.Operator == "not" && negatedInfix[in.Operator]
}

func isNegatedIs(u *ast.UnaryExpression) bool {
	if u.Operator != "not" {
		return false
	}
	switch r := u.Right.(type) {
	case *ast.IsDefinedExpression, *ast.IsEmptyExpression, *ast.IsExpression:
		return true
	case *ast.InfixExpression:
		return r.Operator == "is"
	}
	return false
}

// sub prints e as an operand that must bind at least as tightly as min.
// rightmost marks an operand nothing follows within its expression.
func (p *printer) sub(e ast.Expression, min parser.Precedence, rightmost bool) {
	if needsParens(e, min, rightmost) {
		p.group(e)
		return
	}
	p.expr(e)
}

func needsParens(e ast.Expression, min parser.Precedence, rightmost bool) bool {
	if pc, ok := e.(*ast.PrecedingCommentExpression); ok {
		// The parser wraps the comment around whatever the surrounding
		// expression reads after it, so the comment absorbs anything that
		// follows: it needs parentheses unless it is the last operand, and
		// then exactly when the expression it precedes does.
		return !rightmost || needsParens(pc.Wrap, min, true)
	}
	pr, _ := precOf(e)
	return pr < min && (!rightmost || !prefixLed(e))
}

// postfix prints the left side of a call, field access or index. The parser
// reads the right side of `/` at index precedence, so `a / b.c` is
// `(a / b).c` and `ns/fn(x)` calls `ns/fn`; that form needs no parentheses.
func (p *printer) postfix(e ast.Expression) {
	if in, ok := e.(*ast.InfixExpression); ok && in.Operator == "/" {
		if pr, _ := precOf(in.Right); pr == parser.PRIMARY {
			if lp, _ := precOf(in.Left); lp >= parser.PRODUCT {
				p.expr(e)
				return
			}
		}
	}
	p.sub(e, parser.CALL, false)
}

// group prints e in parentheses. The parser reads `(a ?` as the start of a
// typed lambda, so a ternary whose condition is a bare name and that opens
// the group gets that name parenthesised too (see ternary).
func (p *printer) group(e ast.Expression) {
	p.write("(")
	p.groupOpen = true
	p.expr(e)
	p.write(")")
}

func (p *printer) expr(e ast.Expression) {
	if p.err != nil {
		return
	}
	if isNil(e) {
		p.fail("cannot print a missing expression")
		return
	}
	switch n := e.(type) {
	case *ast.Identifier:
		p.write(n.Value)
	case *ast.IntegerLiteral:
		p.write(strconv.FormatInt(n.Value, 10))
	case *ast.FloatLiteral:
		s := strconv.FormatFloat(n.Value, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		p.write(s)
	case *ast.StringLiteral:
		p.str(n)
	case *ast.TrinaryLiteral:
		p.write(n.Value.String())
	case *ast.NullLiteral:
		p.write("null")
	case *ast.PipelineHoleExpression:
		p.write("#")
	case *ast.ListLiteral:
		items := make([]seqItem, len(n.Values))
		for i, v := range n.Values {
			items[i] = seqItem{startLine(v), v, func() { p.sub(v, parser.LOWEST, true) }}
		}
		p.seq("[", "]", n.Span().From.Line, items)
	case *ast.MapLiteral:
		items := make([]seqItem, len(n.Entries))
		for i, en := range n.Entries {
			items[i] = seqItem{startLine(en.Key), en, func() {
				if s, ok := en.Key.(*ast.StringLiteral); ok {
					p.write(quote(s.Value))
				} else {
					p.write("[")
					p.expr(en.Key)
					p.write("]")
				}
				p.write(": ")
				p.sub(en.Value, parser.LOWEST, true)
			}}
		}
		p.seq("{", "}", n.Span().From.Line, items)
	case *ast.BlockExpression:
		p.block(n)
	case *ast.LambdaExpression:
		p.lambda(n)
	case *ast.CallExpression:
		p.call(n)
	case *ast.FieldAccessExpression:
		p.postfix(n.Left)
		p.write("." + n.Field)
	case *ast.IndexAccessExpression:
		p.postfix(n.Left)
		p.write("[")
		p.expr(n.Index)
		p.write("]")
	case *ast.InfixExpression:
		p.infix(n)
	case *ast.UnaryExpression:
		p.unary(n)
	case *ast.TernaryExpression:
		if _, isName := n.Condition.(*ast.Identifier); isName && p.groupOpen {
			p.group(n.Condition)
		} else {
			// A ternary nested as a condition is parenthesised for the
			// reader; the parser groups `a ? b : c ? d : e` to the left.
			p.sub(n.Condition, parser.TERNARY+1, false)
		}
		if n.Elvis {
			p.write(" ?: ")
			p.sub(n.ElseBranch, parser.COMPARISON+1, true)
			return
		}
		p.write(" ? ")
		p.sub(n.ThenBranch, parser.TERNARY+1, true)
		p.write(" : ")
		p.sub(n.ElseBranch, parser.TERNARY+1, true)
	case *ast.CastExpression:
		p.sub(n.Expr, parser.UNARY, false)
		p.write(" as ")
		p.typeRef(n.TargetType)
	case *ast.IsDefinedExpression, *ast.IsEmptyExpression, *ast.IsExpression:
		p.is(e, false)
	case *ast.PrecedingCommentExpression:
		// A comment that opens its own line keeps what it precedes at the
		// same indentation.
		ownLine, level := p.atStart && !p.pendingNL, p.cur
		p.comment(p.commentText(n.Span(), n.CommentContent))
		if ownLine {
			p.newline()
			p.cur = level
		}
		p.expr(n.Wrap)
	case *ast.TrailingCommentExpression:
		if _, open := precOf(n.Wrap); open {
			p.group(n.Wrap)
		} else {
			p.expr(n.Wrap)
		}
		p.comment(p.commentText(n.Span(), n.CommentContent))
	case *ast.TransformExpression:
		p.write("transform ")
		p.expr(n.Argument)
		p.write(" with " + quote(n.Transformer))
	case *ast.ImportClause:
		p.write("import decision " + n.RuleToImport + " from " + n.FromPolicyFQN.String())
		broken := false
		for _, w := range n.Withs {
			broken = broken || w.Span().From.Line != n.Span().From.Line
		}
		for _, w := range n.Withs {
			p.clause(broken)
			p.write("with " + w.Name + " as ")
			p.sub(w.Expr, parser.LOWEST, true)
		}
	default:
		p.fail("cannot print expression %T at %s", e, e.Span())
	}
}

func (p *printer) infix(n *ast.InfixExpression) {
	pr, ok := opPrec[n.Operator]
	if !ok {
		p.fail("cannot print operator %q at %s", n.Operator, n.Span())
		return
	}
	p.sub(n.Left, pr, false)
	p.write(" " + n.Operator)
	opLine := n.Span().From.Line
	if startLine(n.Right) > opLine || (endLine(n.Left) < opLine && !p.endsInNewline(n.Left)) {
		p.contLine()
	} else {
		p.write(" ")
	}
	rmin := pr + 1
	if n.Operator == "/" {
		rmin = parser.PRIMARY
	}
	p.sub(n.Right, rmin, true)
}

func (p *printer) unary(n *ast.UnaryExpression) {
	if isNegatedInfix(n) {
		in := n.Right.(*ast.InfixExpression)
		p.sub(in.Left, parser.UNARY, false)
		p.write(" not " + in.Operator + " ")
		p.sub(in.Right, parser.UNARY+1, true)
		return
	}
	if isNegatedIs(n) {
		p.is(n.Right, true)
		return
	}
	p.write(n.Operator)
	if n.Operator == "not" {
		p.write(" ")
	}
	if r, ok := n.Right.(*ast.UnaryExpression); ok && n.Operator == "-" && r.Operator == "-" {
		// `--` would start a comment.
		p.group(r)
		return
	}
	p.sub(n.Right, parser.UNARY+1, true)
}

func (p *printer) is(e ast.Expression, negated bool) {
	var left ast.Expression
	switch n := e.(type) {
	case *ast.IsDefinedExpression:
		left = n.Left
	case *ast.IsEmptyExpression:
		left = n.Left
	case *ast.IsExpression:
		left = n.Left
	case *ast.InfixExpression:
		left = n.Left
	}
	p.sub(left, parser.EQUALITY, false)
	p.write(" is ")
	if negated {
		p.write("not ")
	}
	switch n := e.(type) {
	case *ast.IsDefinedExpression:
		p.write("defined")
	case *ast.IsEmptyExpression:
		p.write("empty")
	case *ast.IsExpression:
		if st, ok := n.Type.(*ast.ShapeTypeRef); ok && n.Right != nil {
			p.write(st.Ref.String())
		} else if n.Type != nil {
			p.typeRef(n.Type)
		} else {
			p.sub(n.Right, parser.EQUALITY+1, true)
		}
	case *ast.InfixExpression:
		p.sub(n.Right, parser.EQUALITY+1, true)
	}
}

func (p *printer) call(n *ast.CallExpression) {
	if n.Piped == nil {
		if parts, ok := ast.FlattenSlashIdentChain(n.Callee); ok && len(parts) > 1 {
			// A namespace-qualified callee, `ns/fn(x)`.
			p.write(strings.Join(parts, "/"))
		} else {
			p.postfix(n.Callee)
		}
		p.args(n.Callee, n.Arguments)
		p.memo(n)
		return
	}
	var stages []*ast.CallExpression
	var head ast.Expression = n
	for {
		c, ok := head.(*ast.CallExpression)
		if !ok || c.Piped == nil {
			break
		}
		stages = append([]*ast.CallExpression{c}, stages...)
		head = c.Piped
	}
	broken := false
	for _, s := range stages {
		broken = broken || startLine(s.Callee) > endLine(s.Piped) || p.endsInNewline(s.Piped)
	}
	p.sub(head, parser.PIPELINE, false)
	for _, s := range stages {
		if broken {
			p.contLine()
		} else {
			p.write(" ")
		}
		p.write("|> ")
		p.postfix(s.Callee)
		p.args(s.Callee, s.PipeArgs)
		p.memo(s)
	}
}

func (p *printer) args(callee ast.Expression, args []ast.Expression) {
	items := make([]seqItem, len(args))
	for i, a := range args {
		items[i] = seqItem{startLine(a), a, func() { p.sub(a, parser.LOWEST, true) }}
	}
	openLine := endLine(callee)
	if p.endsInNewline(callee) && len(items) > 0 {
		// The arguments start below a trailing comment after the callee.
		openLine = items[0].start
	}
	p.seq("(", ")", openLine, items)
}

func (p *printer) memo(n *ast.CallExpression) {
	if !n.Memoized {
		return
	}
	p.write("!")
	if n.MemoizeTTL != nil {
		p.write(strconv.FormatInt(int64(*n.MemoizeTTL/time.Second), 10))
	}
}

func (p *printer) lambda(n *ast.LambdaExpression) {
	p.write("(")
	for i, name := range n.Params {
		if i > 0 {
			p.write(", ")
		}
		p.write(name)
		if i < len(n.ParamOpts) && n.ParamOpts[i] {
			p.write("?")
		}
		if i < len(n.ParamTypes) && n.ParamTypes[i] != nil {
			p.write(": ")
			p.typeRef(n.ParamTypes[i])
		}
	}
	p.write(")")
	if n.ReturnType != nil {
		p.write(": ")
		p.typeRef(n.ReturnType)
	}
	p.write(" => ")
	if n.Body == nil {
		p.fail("cannot print lambda without a body at %s", n.Span())
		return
	}
	p.block(n.Body)
}

func (p *printer) block(b *ast.BlockExpression) {
	if b.Yield == nil {
		p.fail("cannot print block without a yield at %s", b.Span())
		return
	}
	if len(b.Statements) == 0 && b.Span().From.Line == b.Span().To.Line && !p.forcesBreak(b.Yield, true) {
		p.write("{ yield ")
		p.expr(b.Yield)
		p.write(" }")
		return
	}
	es := make([]entry, 0, len(b.Statements)+1)
	for _, s := range b.Statements {
		es = append(es, p.stmtEntry(s))
	}
	es = append(es, entry{start: startLine(b.Yield), print: func() {
		p.write("yield ")
		p.expr(b.Yield)
	}})
	p.write("{")
	p.nest(func() { p.entries(es, false) }, "}")
}

// seqItem is one element of a list, map or argument list.
type seqItem struct {
	start int
	node  any
	print func()
}

// seq prints a bracketed list on one line, or one item per line when an item
// starts on a different line from the one before it (or from the opening
// bracket) or holds a comment or heredoc.
func (p *printer) seq(open, close string, openLine int, items []seqItem) {
	if len(items) == 0 {
		p.write(open + close)
		return
	}
	broken, prev := false, openLine
	for _, it := range items {
		broken = broken || it.start != prev || p.forcesBreak(it.node, false)
		prev = it.start
	}
	p.write(open)
	if !broken {
		for i, it := range items {
			if i > 0 {
				p.write(", ")
			}
			it.print()
		}
		p.write(close)
		return
	}
	p.nest(func() {
		glued, prevEnd := false, openLine
		for i, it := range items {
			// A comment that followed the comma (or the opening bracket) on
			// the previous item's line stays there; the parser attaches it
			// to this item as a preceding comment.
			pc, hoist := it.node.(*ast.PrecedingCommentExpression)
			hoist = hoist && !p.pendingNL && !glued && pc.Span().From.Line == prevEnd
			if hoist {
				p.comment(p.commentText(pc.Span(), pc.CommentContent))
			}
			if !glued {
				p.line()
			}
			glued = false
			if hoist {
				p.expr(pc.Wrap)
			} else {
				it.print()
			}
			prevEnd = 0
			if e, ok := it.node.(ast.Expression); ok {
				prevEnd = endLine(e)
			}
			if i == len(items)-1 {
				break
			}
			if !p.pendingNL {
				p.write(",")
			} else if !p.safeAfterComment(items[i+1]) {
				// The next item would continue this one without a comma, so
				// it starts its own line with the comma in front.
				p.newline()
				p.cur = p.base
				p.write(", ")
				glued = true
			}
		}
	}, close)
}

// safeAfterComment reports whether it can follow an item that ended in a line
// comment with no comma between them. It renders it on its own and rejects a
// first token that would continue the previous item.
func (p *printer) safeAfterComment(it seqItem) bool {
	saved := *p
	p.out, p.atStart, p.pendingNL = &strings.Builder{}, true, false
	it.print()
	text := p.out.String()
	*p = saved
	text = strings.TrimLeft(text, " ")
	switch {
	case text == "", strings.HasPrefix(text, "--"):
		return true
	case strings.ContainsAny(text[:1], "-+!(["), strings.HasPrefix(text, "not ") || strings.HasPrefix(text, "not("):
		return false
	}
	return true
}

// str prints a string literal, as a heredoc when the source wrote it as one.
func (p *printer) str(n *ast.StringLiteral) {
	tag, ok := p.heredocTag(n.Span().From.Offset)
	if !ok {
		p.write(quote(n.Value))
		return
	}
	for i := 1; hasLine(n.Value, tag); i++ {
		tag = fmt.Sprintf("EOT%d", i)
	}
	p.write("<<<" + tag)
	p.out.WriteString("\n" + n.Value + tag)
	p.pendingNL = true
}

// heredocTag returns the tag of the heredoc starting at offset in the
// source, if one does.
func (p *printer) heredocTag(offset int) (string, bool) {
	if offset < 0 || offset > len(p.src) || !bytes.HasPrefix(p.src[offset:], []byte("<<<")) {
		return "", false
	}
	rest := p.src[offset+3:]
	end := bytes.IndexFunc(rest, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' })
	if end < 0 {
		end = len(rest)
	}
	if end == 0 {
		return "EOT", true
	}
	return string(rest[:end]), true
}

func hasLine(s, line string) bool {
	for _, l := range strings.Split(s, "\n") {
		if l == line {
			return true
		}
	}
	return false
}

func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// isNil reports whether v is nil or a nil pointer in an interface.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}

// startLine is the source line where e's printed form begins.
func startLine(e ast.Expression) int {
	switch n := e.(type) {
	case *ast.InfixExpression:
		return startLine(n.Left)
	case *ast.CallExpression:
		if n.Piped != nil {
			return startLine(n.Piped)
		}
		return startLine(n.Callee)
	case *ast.FieldAccessExpression:
		return startLine(n.Left)
	case *ast.IndexAccessExpression:
		return startLine(n.Left)
	case *ast.TernaryExpression:
		return startLine(n.Condition)
	case *ast.CastExpression:
		return startLine(n.Expr)
	case *ast.IsDefinedExpression:
		return startLine(n.Left)
	case *ast.IsEmptyExpression:
		return startLine(n.Left)
	case *ast.IsExpression:
		return startLine(n.Left)
	case *ast.UnaryExpression:
		if isNegatedInfix(n) || isNegatedIs(n) {
			return startLine(n.Right)
		}
	case *ast.TrailingCommentExpression:
		return startLine(n.Wrap)
	}
	return e.Span().From.Line
}

// endLine is the source line where e ends.
func endLine(e ast.Expression) int {
	if pc, ok := e.(*ast.PrecedingCommentExpression); ok {
		return endLine(pc.Wrap)
	}
	return e.Span().To.Line
}

// endsInNewline reports whether e's printed form ends with a line comment or
// a heredoc, which force a line break after it.
func (p *printer) endsInNewline(e ast.Expression) bool {
	switch n := e.(type) {
	case *ast.TrailingCommentExpression:
		return true
	case *ast.StringLiteral:
		_, heredoc := p.heredocTag(n.Span().From.Offset)
		return heredoc
	case *ast.InfixExpression:
		return p.endsInNewline(n.Right)
	case *ast.UnaryExpression:
		return p.endsInNewline(n.Right)
	case *ast.TernaryExpression:
		return p.endsInNewline(n.ElseBranch)
	case *ast.PrecedingCommentExpression:
		return p.endsInNewline(n.Wrap)
	case *ast.IsExpression:
		return n.Type == nil && n.Right != nil && p.endsInNewline(n.Right)
	}
	return false
}

// endsInNullableName reports whether e prints ending in `Name?`.
func endsInNullableName(e ast.Expression) bool {
	switch n := e.(type) {
	case *ast.IsExpression:
		nt, ok := n.Type.(*ast.NullableTypeRef)
		if !ok || n.Right != nil || len(nt.GetConstraints()) > 0 {
			return false
		}
		_, isName := nt.Inner.(*ast.ShapeTypeRef)
		return isName
	case *ast.InfixExpression:
		return endsInNullableName(n.Right)
	case *ast.UnaryExpression:
		return endsInNullableName(n.Right)
	case *ast.TernaryExpression:
		return endsInNullableName(n.ElseBranch)
	}
	return false
}

var (
	precedingType = reflect.TypeFor[*ast.PrecedingCommentExpression]()
	trailingType  = reflect.TypeFor[*ast.TrailingCommentExpression]()
	commentType   = reflect.TypeFor[*ast.CommentStatement]()
	stringType    = reflect.TypeFor[*ast.StringLiteral]()
	blockType     = reflect.TypeFor[*ast.BlockExpression]()
)

// forcesBreak reports whether node holds something that always prints over
// several lines: a comment, a heredoc and, when blocks is set, a block with
// statements.
func (p *printer) forcesBreak(node any, blocks bool) bool {
	return walk(reflect.ValueOf(node), func(v reflect.Value) bool {
		switch v.Type() {
		case precedingType, trailingType, commentType:
			return true
		case stringType:
			// Read through reflection: the literal may sit behind an
			// unexported field, where calling Span would panic.
			offset := v.Elem().Field(0).Elem().FieldByName("Rnge").FieldByName("From").FieldByName("Offset")
			_, ok := p.heredocTag(int(offset.Int()))
			return ok
		case blockType:
			return blocks && v.Elem().FieldByName("Statements").Len() > 0
		}
		return false
	})
}

// walk visits every pointer reachable from v and stops at the first one for
// which visit returns true.
func walk(v reflect.Value, visit func(reflect.Value) bool) bool {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return false
		}
		return visit(v) || walk(v.Elem(), visit)
	case reflect.Interface:
		return !v.IsNil() && walk(v.Elem(), visit)
	case reflect.Struct:
		if v.Type() == rangeType {
			return false
		}
		for i := range v.NumField() {
			if walk(v.Field(i), visit) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			if walk(v.Index(i), visit) {
				return true
			}
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			if walk(it.Value(), visit) {
				return true
			}
		}
	}
	return false
}
