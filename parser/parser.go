package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flowfunction/cddl/ast"
	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/token"
)

type Parser struct {
	lexer  *lexer.Lexer
	errors []lexer.Error

	pos     token.Position
	peekPos token.Position

	currToken token.Token
	peekToken token.Token

	currliteral string
	peekLiteral string

	// NULL DENOTATIONS
	nuds map[token.Token]nudParseFn

	// LEFT DENOTATIONS
	leds map[token.Token]ledParseFn
}

type nudParseFn func() ast.Entry
type ledParseFn func(ast.Entry) ast.Entry

type ParserConfig struct{}

func (p *Parser) Parse() *ast.CDDL {
	cddl := &ast.CDDL{}
	cddl.Rules = []*ast.Rule{}

	// Don't parse if lexer errors are non zero
	if len(p.lexer.Errors) != 0 {
		return nil
	}

	for p.currToken != token.EOF {
		rule := p.parseRule()
		if rule != nil {
			cddl.Rules = append(cddl.Rules, rule)
		}
		p.next()
	}

	return cddl

}

func (p *Parser) expect(tok token.Token) bool {
	if p.currToken != tok {
		p.errorTokenExpected(p.pos, tok)
		return false
	}
	p.next()
	return true
}

func (p *Parser) expectIntLiteral2() bool {
	if p.peekToken != token.INT {
		p.errorTokenExpected(p.pos, token.INT)
		return false
	}
	if p.peekToken.IsLiteral(p.peekLiteral) {
		p.errors = append(p.errors, lexer.Error{
			Pos: p.pos,
			Msg: fmt.Sprintf("expected integer literal at line %d column %d got `%s`", p.pos.Line, p.pos.Column, p.currliteral),
		})
		p.next()
		return true
	}

	return false
}

func (p *Parser) Errors() []lexer.Error {
	return p.errors
}

func (p *Parser) parseRule() *ast.Rule {
	rule := &ast.Rule{}
	if p.currToken != token.IDENT {
		p.errorTokenExpected(p.pos, token.IDENT)
		return nil
	}
	rule.Name = &ast.Identifer{Pos: p.pos, Name: p.currliteral}
	p.next()

	var entry ast.Entry

	switch p.currToken {
	case token.ASSIGN:
		p.next()
		entry = p.parseEntry(token.LOWEST)
	case token.TYPE_CHOICE_ASSIGN:
		p.next()
		entry = p.parseEntry(token.LOWEST)
	case token.GROUP_CHOICE_ASSIGN:
		p.next()
		entry = p.parseEntry(token.LOWEST)
	default:
		p.errors = append(p.errors, lexer.Error{
			Pos: p.pos,
			Msg: fmt.Sprintf("expected assigment operators =, /= or //= after identifer `%s`", rule.Name.Name),
		})
		return nil
	}
	rule.Value = entry

	return rule
}

func (p *Parser) parseEntry(precedence int) ast.Entry {
	nudFn := p.nuds[p.currToken]
	if nudFn == nil {
		p.errorNudExpected(p.pos, p.currToken)
		return nil
	}
	exp := nudFn()

	for precedence < p.peekToken.Precedence() {
		ledFn := p.leds[p.peekToken]
		if ledFn == nil {
			return exp
		}
		p.next()
		exp = ledFn(exp)
	}

	return exp
}

func (p *Parser) parseIdentifer() ast.Entry {
	return &ast.Identifer{Pos: p.pos, Name: p.currliteral}
}

func (p *Parser) parseBooleanType() ast.Entry {
	return &ast.BooleanType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseTstrType() ast.Entry {
	return &ast.TstrType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseFloatType() ast.Entry {
	return &ast.FloatType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseIntegerType() ast.Entry {
	return &ast.IntegerType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseBstrType() ast.Entry {
	return &ast.BstrType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseBytesType() ast.Entry {
	return &ast.BytesType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseNullType() ast.Entry {
	return &ast.NullType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseColon(left ast.Entry) ast.Entry {
	var ident *ast.Identifer
	if ide, ok := left.(*ast.Identifer); !ok {
		p.errorTokenExpected(p.pos, token.IDENT)
		return nil
	} else {
		ident = ide
	}
	rule := &ast.Rule{
		Pos:   p.pos,
		Token: p.currToken,
		Name:  ident,
	}
	p.next()

	rule.Value = p.parseEntry(p.currToken.Precedence())

	return rule
}

func (p *Parser) parseTypeChoice(left ast.Entry) ast.Entry {
	tc := &ast.TypeChoice{
		Pos:   p.pos,
		Token: p.currToken,
		First: left,
	}
	p.next()
	tc.Second = p.parseEntry(p.currToken.Precedence())
	return tc
}

func (p *Parser) parseSizeOperator(left ast.Entry) ast.Entry {
	sop := &ast.SizeOperatorControl{
		Pos:   p.pos,
		Token: p.currToken,
		Type:  left,
	}
	fmt.Printf("parseSizeOperator() -> %s, %s\n", p.currliteral, p.peekLiteral)

	if p.expectIntLiteral2() {
		size := p.parseIntegerLiteral()
		sop.Size = size
		return sop
	} else {
		p.errorTokenExpected(p.pos, token.INT)
	}

	return sop
}

func (p *Parser) parseIntegerLiteral() *ast.IntegerLiteral {
	if p.currToken.IsLiteral(p.currliteral) {
		lit, err := strconv.Atoi(p.currliteral)
		if err != nil {
			p.errorTokenExpected(p.pos, token.INT)
			return nil
		}
		return &ast.IntegerLiteral{
			Pos:     p.pos,
			Token:   p.currToken,
			Literal: lit,
		}
	}
	return nil
}

func (p *Parser) parseGroup() ast.Entry {
	g := &ast.Group{}
	g.Pos = p.pos
	p.next()

	for p.currToken != token.RBRACE {
		rule := p.parseEntry(token.LOWEST)
		if rule != nil {
			g.Rules = append(g.Rules, rule)
		}
		p.next()
	}
	return g
}

func (p *Parser) errorTokenExpected(pos token.Position, tok token.Token) {
	p.errors = append(p.errors, lexer.Error{
		Pos: pos,
		Msg: fmt.Sprintf("parser error: expected %s at line %d, column %d", tok.String(), pos.Line, pos.Column)})
}

func (p *Parser) errorTokensExpected(pos token.Position, toks ...token.Token) {
	if len(toks) == 0 {
		return
	}
	ts := make([]string, len(toks))
	for _, tok := range toks {
		ts = append(ts, tok.String())
	}

	p.errors = append(p.errors, lexer.Error{
		Pos: pos,
		Msg: fmt.Sprintf("parser error: expected one of %s at line %d, column %d", strings.Join(ts, ","), pos.Line, pos.Column),
	})
}

func (p *Parser) errorNudExpected(pos token.Position, tok token.Token) {
	p.errors = append(p.errors, lexer.Error{
		Pos: pos,
		Msg: fmt.Sprintf("parser internal error: expected registration for prefix operator %s", tok),
	})
}

func (p *Parser) next() {
	p.currToken = p.peekToken
	p.currliteral = p.peekLiteral
	p.pos = p.peekPos
	p.peekToken, p.peekPos, p.peekLiteral = p.lexer.Scan()
	// fmt.Printf("Token: %s: %s\n", p.currToken, p.currliteral)
}

func NewParser(l *lexer.Lexer, opts ...ParserConfig) *Parser {
	p := &Parser{}
	p.lexer = l

	p.nuds = make(map[token.Token]nudParseFn)
	p.leds = make(map[token.Token]ledParseFn)

	p.nuds[token.IDENT] = p.parseIdentifer
	p.nuds[token.BOOL] = p.parseBooleanType
	p.nuds[token.TSTR] = p.parseTstrType
	p.nuds[token.TEXT] = p.parseTstrType
	p.nuds[token.FLOAT] = p.parseFloatType
	p.nuds[token.FLOAT16] = p.parseFloatType
	p.nuds[token.FLOAT32] = p.parseFloatType
	p.nuds[token.FLOAT64] = p.parseFloatType
	p.nuds[token.INT] = p.parseIntegerType
	p.nuds[token.BSTR] = p.parseBstrType
	p.nuds[token.BYTES] = p.parseBytesType
	p.nuds[token.NULL] = p.parseNullType
	p.nuds[token.NIL] = p.parseNullType
	p.nuds[token.LBRACE] = p.parseGroup

	p.leds[token.COLON] = p.parseColon
	p.leds[token.TYPE_CHOICE] = p.parseTypeChoice
	p.leds[token.SIZE] = p.parseSizeOperator

	p.next()
	p.next()

	return p
}
