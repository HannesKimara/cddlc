package astutils

import (
	"fmt"
	"strings"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/HannesKimara/cddlc/token"
)

const spaceChar = "\t"

type printer struct {
	level   uint
	builder *strings.Builder
}

func (p *printer) format(node ast.Node) string {
	var s string

	p.level += 1
	switch val := node.(type) {
	case *ast.Identifier:
		s = p.formatIdent(val)
	case *ast.TypeChoice:
		s = p.formatTypeChoice(val)
	case *ast.IntegerLiteral:
		s = p.formatIntegerLiteral(val)
	case *ast.FloatLiteral:
		s = p.formatFloatLiteral(val)
	case *ast.IntegerType:
		s = p.formatIntType(val)
	case *ast.FloatType:
		s = p.formatFloatType(val)
	case *ast.Enumeration:
		s = p.formatEnumeration(val)
	case *ast.Group:
		s = p.formatGroup(val)
	default:
		s = fmt.Sprintf("%T", val)
	}
	p.level -= 1

	return s
}

func (p *printer) formatIdent(ident *ast.Identifier) string {
	return fmt.Sprintf("Ident ( Name: '%s' ) %s", ident.Name, p.formatPosition(ident.Pos))
}

func (p *printer) formatTypeChoice(ch *ast.TypeChoice) string {
	pre := p.spaceLevel(spaceChar)
	return fmt.Sprintf("TypeChoice %s (\n%sFirst  : %s, \n%sSecond : %s\n)", p.formatPosition(ch.Pos), pre, p.format(ch.First), pre, p.format(ch.Second))
}

func (p *printer) formatIntegerLiteral(il *ast.IntegerLiteral) string {
	return fmt.Sprintf("IntegerLiteral (Literal: '%d') %s", il.Literal, p.formatPosition(il.Pos))
}

func (p *printer) formatFloatLiteral(fl *ast.FloatLiteral) string {
	return fmt.Sprintf("FloatLiteral ( Literal: '%f' ) %s", fl.Literal, p.formatPosition(fl.Start()))
}

func (p *printer) formatIntType(it *ast.IntegerType) string {
	return fmt.Sprintf("IntegerType ( Token: '%s' ) %s", it.Token, p.formatPosition(it.Start()))
}

func (p *printer) formatFloatType(ft *ast.FloatType) string {
	return fmt.Sprintf("FloatType ( Token: '%s' ) %s", ft.Token, p.formatPosition(ft.Start()))
}

func (p *printer) formatEnumeration(en *ast.Enumeration) string {
	pre := p.spaceLevel(spaceChar)
	return fmt.Sprintf("Enumeration %s (\n%sValue: %s\n%s)", p.formatPosition(en.Pos), pre, p.format(en.Value), pre)
}

func (p *printer) formatGroup(gl *ast.Group) string {
	pre := p.spaceLevel(spaceChar)
	epre := pre + spaceChar
	return fmt.Sprintf("Group %s (\n%sEntries: %d\n%s)", p.formatPosition(gl.Pos), epre, len(gl.Entries), pre)
}

func (p *printer) formatPosition(pos token.Position) string {
	return fmt.Sprintf("{ Column: %d, Line: %d }", pos.Column, pos.Line)
}

func (p *printer) spaceLevel(space string) string {
	return strings.Repeat(space, int(p.level))
}

func Print(node ast.Node) {
	p := printer{
		level:   0,
		builder: &strings.Builder{},
	}

	fmt.Println(p.format(node))
}

func Format(node ast.Node) string {
	p := printer{
		level:   0,
		builder: &strings.Builder{},
	}

	return p.format(node)
}
