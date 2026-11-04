// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package format prints a parsed Sentrie program in one canonical style. It
// is the engine behind `sentrie fmt`.
//
// The style follows the dominant style of example_pack/ and the docs:
//
//   - Two spaces per indentation level, no tabs, no trailing whitespace, and
//     a file ends with exactly one newline.
//   - One statement per line; semicolons are dropped.
//   - At the top level a blank line follows every statement that is not a
//     comment. Elsewhere blank lines between statements are kept, a run of
//     them collapses to one, and there is never a blank line just inside a
//     brace.
//   - Binary operators, `=`, `=>`, `|>`, `as`, and the `?` and `:` of a
//     ternary have one space on each side; the Elvis operator prints as `?:`.
//     `name: Type` has a space after the colon only. Lists, maps and calls
//     have no space inside their brackets and use `, ` between items; `use`
//     lists print as `use { a, b }`.
//   - A block prints as `{ yield x }` when it has no statements and was
//     written on one line; otherwise every statement and the yield get a line.
//     Policy and shape bodies always put each member on its own line.
//   - Line breaks inside lists, maps, call arguments, pipeline chains, binary
//     expressions, `with` clauses and `attach` clauses follow the source: a
//     construct written on one line stays on one line, and one written across
//     lines gets one item per line, indented one level. A binary expression
//     breaks after its operator. Long lines are not wrapped.
//   - Parentheses are printed where precedence needs them and nowhere else.
//   - Comments stay where they were and print as `-- text`, keeping any
//     indentation inside the text: a comment on its own line stays on its
//     own line, and a trailing comment stays at the end of the line it
//     followed. In a broken list, map or argument list, an item
//     followed by a trailing comment drops its comma, and so does an item
//     that is a heredoc, since the heredoc's terminator ends the line. Both
//     rely on the parser treating these commas as optional (parseListLiteral,
//     parseMapLiteral and parseExpressionList); when the next item would
//     continue the previous one, the comma starts the next line instead. A
//     comment after a comma stays after it, on the same line.
//   - A heredoc stays a heredoc with its tag; any other string prints quoted.
//   - Spellings the AST does not keep are normalised: a fact or `use` alias
//     equal to the default name is dropped, `not (x in y)` prints as
//     `x not in y`, `not (x is defined)` as `x is not defined`, and a shape
//     reference in `is` prints without spaces around `/`.
//
// Source formats a whole file and then checks that its output parses back to
// an equivalent program (see Equal), so a printer bug fails loudly instead of
// changing what a policy means.
package format

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
)

// Source formats src, the contents of the file filename. It returns an error
// when src does not parse, or when the formatted text would not parse back
// to a program equal to the original.
func Source(ctx context.Context, filename string, src []byte) ([]byte, error) {
	prg, err := parse(ctx, filename, src)
	if err != nil {
		return nil, err
	}
	out, err := Program(prg, src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	again, err := parse(ctx, filename, out)
	if err != nil {
		return nil, fmt.Errorf("%s: formatted output does not parse: %w", filename, err)
	}
	if err := Equal(prg, again); err != nil {
		return nil, fmt.Errorf("%s: formatted output changes the program: %w", filename, err)
	}
	return out, nil
}

func parse(ctx context.Context, filename string, src []byte) (*ast.Program, error) {
	return parser.NewParser(bytes.NewReader(src), filename).ParseProgram(ctx)
}

// Program prints prg. src is the text prg was parsed from; it is read for the
// blank lines between statements, the tags of heredocs and the exact text of
// comments. Without it comments print from their trimmed content and heredocs
// as quoted strings. A nil program prints as nothing.
func Program(prg *ast.Program, src []byte) ([]byte, error) {
	if prg == nil {
		return []byte{}, nil
	}
	p := &printer{src: src, lines: strings.Split(string(src), "\n"), out: &strings.Builder{}, atStart: true}
	entries := make([]entry, 0, len(prg.Statements))
	for _, s := range prg.Statements {
		entries = append(entries, p.stmtEntry(s))
	}
	p.entries(entries, true)
	if p.err != nil {
		return nil, p.err
	}
	out := p.out.String()
	if out == "" {
		return []byte{}, nil
	}
	return []byte(out + "\n"), nil
}

// printer writes text with lazy indentation. A line comment or a heredoc
// sets pendingNL, so whatever is written next starts a new line.
type printer struct {
	src       []byte
	lines     []string
	out       *strings.Builder
	base      int  // indentation of the current statement list
	cur       int  // indentation of a new line in the current statement
	atStart   bool // nothing written on the current line yet
	pendingNL bool
	groupOpen bool // a group was just opened and nothing written after it
	err       error
}

func (p *printer) fail(format string, args ...any) {
	if p.err == nil {
		p.err = fmt.Errorf(format, args...)
	}
}

func (p *printer) write(s string) {
	if s == "" {
		return
	}
	p.groupOpen = false
	if p.pendingNL {
		p.newline()
		p.cur = p.base + 1
	}
	if p.atStart {
		s = strings.TrimLeft(s, " ")
		if s == "" {
			return
		}
		p.out.WriteString(strings.Repeat("  ", p.cur))
		p.atStart = false
	}
	p.out.WriteString(s)
}

func (p *printer) newline() {
	p.out.WriteByte('\n')
	p.atStart = true
	p.pendingNL = false
}

// line starts a new line at the level of the current statement list.
func (p *printer) line() {
	if !p.atStart || p.pendingNL {
		p.newline()
	}
	p.cur = p.base
}

// contLine starts a continuation line of the current statement.
func (p *printer) contLine() {
	if !p.atStart || p.pendingNL {
		p.newline()
	}
	p.cur = p.base + 1
}

// comment writes text, a rendered comment, and ends the line after it.
func (p *printer) comment(text string) {
	p.groupOpen = false
	if p.pendingNL {
		p.newline()
		p.cur = p.base + 1
	}
	if !p.atStart && !strings.HasSuffix(p.out.String(), " ") {
		p.out.WriteByte(' ')
	}
	p.write(text)
	p.pendingNL = true
}

// commentText renders the comment spanning rng. It reads the source, since
// the lexer trims the comment's content and would lose inner indentation:
// only trailing whitespace goes, and a space is added after `--` when the
// text does not start with one. content is the fallback without a source.
func (p *printer) commentText(rng tokens.Range, content string) string {
	from, to := rng.From.Offset, rng.To.Offset
	if from >= 0 && from+2 <= to && to <= len(p.src) && bytes.HasPrefix(p.src[from:], []byte("--")) {
		content = strings.TrimRight(string(p.src[from+2:to]), " \t\r")
		if content == "" || content[0] == ' ' || content[0] == '\t' {
			return "--" + content
		}
	}
	if content == "" {
		return "--"
	}
	return "-- " + content
}

// nest prints body one level deeper than the current line, then puts close
// on a new line at the current line's level.
func (p *printer) nest(body func(), close string) {
	base, cur := p.base, p.cur
	p.base = cur + 1
	body()
	p.base, p.cur = base, cur
	if !p.atStart || p.pendingNL {
		p.newline()
	}
	p.write(close)
}

// blankBefore reports whether the source line just above line is empty.
func (p *printer) blankBefore(line int) bool {
	i := line - 2
	return i >= 0 && i < len(p.lines) && strings.TrimSpace(p.lines[i]) == ""
}

// entry is one member of a statement list: a statement, a shape field, or
// the yield of a block.
type entry struct {
	start   int
	comment *ast.CommentStatement
	print   func()
}

func (p *printer) stmtEntry(s ast.Statement) entry {
	if isNil(s) {
		return entry{print: func() { p.fail("cannot print a missing statement") }}
	}
	if c, ok := s.(*ast.CommentStatement); ok {
		return entry{start: c.Span().From.Line, comment: c}
	}
	return entry{start: s.Span().From.Line, print: func() { p.stmt(s) }}
}

// entries prints a statement list. A trailing comment goes back on the line
// before it, so after an opening brace when it is the first entry.
func (p *printer) entries(es []entry, top bool) {
	prevStart, prevStmt := 0, false
	for i, e := range es {
		if e.comment != nil && e.comment.Trailing && !p.pendingNL && (i > 0 || !top) {
			p.comment(p.commentText(e.comment.Span(), e.comment.Content))
			continue
		}
		blank := i > 0 && e.start > prevStart && p.blankBefore(e.start)
		if top && prevStmt {
			blank = true
		}
		p.line()
		if blank && p.out.Len() > 0 {
			p.out.WriteByte('\n')
		}
		if e.comment != nil {
			p.comment(p.commentText(e.comment.Span(), e.comment.Content))
		} else {
			e.print()
		}
		prevStart, prevStmt = e.start, e.comment == nil
	}
}

func (p *printer) body(es []entry, empty string) {
	if len(es) == 0 {
		p.write(empty)
		return
	}
	p.write(" {")
	p.nest(func() { p.entries(es, false) }, "}")
}

func (p *printer) stmt(s ast.Statement) {
	switch n := s.(type) {
	case *ast.NamespaceStatement:
		p.write("namespace " + n.Name.String())
	case *ast.PolicyStatement:
		p.write("policy " + n.Name)
		es := make([]entry, 0, len(n.Statements))
		for _, st := range n.Statements {
			es = append(es, p.stmtEntry(st))
		}
		p.body(es, " {}")
	case *ast.CommentStatement:
		p.comment(p.commentText(n.Span(), n.Content))
	case *ast.TitleStatement:
		p.write("title " + quote(n.Value))
	case *ast.DescriptionStatement:
		p.write("description " + quote(n.Value))
	case *ast.VersionStatement:
		p.write("version " + quote(n.Literal))
	case *ast.TagStatement:
		p.write("tag " + quote(n.Key) + " = " + quote(n.Value))
	case *ast.FactStatement:
		p.write("fact " + n.Name)
		if n.Optional {
			p.write("?")
		}
		p.write(": ")
		p.typeRef(n.Type)
		if n.Alias != "" && n.Alias != n.Name {
			p.write(" as " + n.Alias)
		}
		if n.Default != nil {
			p.write(" default ")
			p.sub(n.Default, parser.LOWEST, true)
		}
	case *ast.UseStatement:
		p.use(n)
	case *ast.VarDeclaration:
		p.write("let " + n.Name)
		if n.Type != nil {
			p.write(": ")
			p.typeRef(n.Type)
		}
		p.write(" = ")
		p.sub(n.Value, parser.LOWEST, true)
	case *ast.RuleStatement:
		p.rule(n)
	case *ast.RuleExportStatement:
		p.write("export decision of " + n.Of)
		broken := false
		for _, a := range n.Attachments {
			broken = broken || a.Span().From.Line != n.Span().From.Line
		}
		for _, a := range n.Attachments {
			p.clause(broken)
			p.write("attach " + a.What + " as ")
			p.sub(a.As, parser.LOWEST, true)
		}
	case *ast.ShapeStatement:
		p.shape(n)
	case *ast.DeriveStatement:
		p.write("derive " + n.Name + " = ")
		p.expr(n.Value)
	case *ast.ExportDeriveStatement:
		p.write("export derive " + n.Name)
	case *ast.ShapeExportStatement:
		p.write("export shape " + n.Name)
	default:
		p.fail("cannot print statement %T at %s", s, s.Span())
	}
}

// clause starts a `with` or `attach` clause on a continuation line when the
// clauses are broken, and after a space otherwise.
func (p *printer) clause(broken bool) {
	if broken {
		p.contLine()
	} else {
		p.write(" ")
	}
}

func (p *printer) use(n *ast.UseStatement) {
	p.write("use { " + strings.Join(n.Modules, ", ") + " } from ")
	var dflt string
	if n.RelativeFrom != "" {
		p.write(quote(n.RelativeFrom))
		parts := strings.Split(n.RelativeFrom, "/")
		dflt = parts[len(parts)-1]
	} else {
		p.write("@" + strings.Join(n.LibFrom, "/"))
		if len(n.LibFrom) > 0 {
			dflt = n.LibFrom[len(n.LibFrom)-1]
		}
	}
	if n.As != "" && n.As != dflt {
		p.write(" as " + n.As)
	}
}

func (p *printer) rule(n *ast.RuleStatement) {
	p.write("rule " + n.RuleName + " =")
	if n.Default != nil {
		p.write(" default ")
		p.guard(n.Default)
	}
	if n.When != nil {
		p.write(" when ")
		p.guard(n.When)
	}
	p.write(" ")
	p.sub(n.Body, parser.LOWEST, true)
}

// guard prints a `default` or `when` expression. One that ends in a nullable
// shape type (`x is Foo?`) is parenthesised, since a body after it would make
// the parser read the `?` as a ternary.
func (p *printer) guard(e ast.Expression) {
	if endsInNullableName(e) {
		p.group(e)
		return
	}
	p.sub(e, parser.LOWEST, true)
}

func (p *printer) shape(n *ast.ShapeStatement) {
	p.write("shape " + n.Name)
	if n.Complex == nil {
		if n.Simple == nil {
			p.fail("cannot print shape %s at %s: it has no type", n.Name, n.Span())
			return
		}
		p.write(" ")
		p.typeRef(n.Simple)
		return
	}
	c := n.Complex
	if c.With != nil {
		p.write(" with " + c.With.String())
	}
	type member struct {
		offset int
		e      entry
	}
	members := make([]member, 0, len(c.Fields)+len(c.Comments))
	for _, f := range c.Fields {
		members = append(members, member{f.Range.From.Offset, entry{start: f.Range.From.Line, print: func() {
			p.write(f.Name)
			if f.Optional {
				p.write("?")
			}
			p.write(": ")
			p.typeRef(f.Type)
		}}})
	}
	for _, cm := range c.Comments {
		members = append(members, member{cm.Span().From.Offset, entry{start: cm.Span().From.Line, comment: cm}})
	}
	sort.SliceStable(members, func(i, j int) bool { return members[i].offset < members[j].offset })
	es := make([]entry, len(members))
	for i, m := range members {
		es[i] = m.e
	}
	p.body(es, " {}")
}

func (p *printer) typeRef(t ast.TypeRef) {
	if isNil(t) {
		p.fail("cannot print a missing type")
		return
	}
	base, nullable := t, false
	if n, ok := t.(*ast.NullableTypeRef); ok {
		base, nullable = n.Inner, true
	}
	switch b := base.(type) {
	case *ast.StringTypeRef:
		p.write("string")
	case *ast.NumberTypeRef:
		p.write("number")
	case *ast.TrinaryTypeRef:
		if b.Boolean {
			p.write("boolean")
		} else {
			p.write("trinary")
		}
	case *ast.DocumentTypeRef:
		p.write("document")
	case *ast.ShapeTypeRef:
		p.write(b.Ref.String())
	case *ast.ListTypeRef:
		p.write("list[")
		p.typeRef(b.ElemType)
		p.write("]")
	case *ast.DictTypeRef:
		p.write("dict[")
		p.typeRef(b.ValueType)
		p.write("]")
	case *ast.RecordTypeRef:
		p.write("record[")
		for i, f := range b.Fields {
			if i > 0 {
				p.write(", ")
			}
			p.typeRef(f)
		}
		p.write("]")
	default:
		p.fail("cannot print type %T at %s", base, base.Span())
		return
	}
	if nullable {
		p.write("?")
	}
	for _, c := range t.GetConstraints() {
		p.write(" @" + c.Name + "(")
		for i, a := range c.Args {
			if i > 0 {
				p.write(", ")
			}
			p.expr(a)
		}
		p.write(")")
	}
}
