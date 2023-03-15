package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/flowfunction/cddl/ast"
	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/token"

	env "github.com/flowfunction/cddl/environment"
)

type taskFn func()

type Parser struct {
	// instance of lexer
	lexer *lexer.Lexer

	// diagnostics contains the slice of errors and warnings in order
	diagnostics []Diagnostic

	// current position
	pos token.Position

	// next peeked position
	peekPos token.Position

	// current token
	currToken token.Token

	// next peeked literal
	peekToken token.Token

	// current literal
	currliteral string

	// next peeked literal
	peekLiteral string

	// null Denotations
	nuds map[token.Token]nudParseFn

	// left Denotations
	leds map[token.Token]ledParseFn

	environment *env.Environment

	// hold tasks to be run after the completed ast build.
	// used mostly to check types in type specific operators that may not exist
	// in the environment at first pass
	tasks []taskFn
}

type nudParseFn func() ast.Node
type ledParseFn func(ast.Node) ast.Node

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

	for _, task := range p.tasks {
		task()
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
		p.errorTokenExpected(p.peekPos, token.INT)
		return false
	}
	if !p.peekToken.IsLiteral(p.peekLiteral) {
		p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("expected integer literal at line %d column %d got `%s`", p.peekPos.Line, p.peekPos.Column, p.peekLiteral), p.pos, p.pos))
		return false
	}
	p.next()

	return true
}

func (p *Parser) Errors() []Diagnostic {
	return p.diagnostics
}

func (p *Parser) parseRule() *ast.Rule {
	rule := &ast.Rule{}

	switch p.currToken {
	case token.COMMENT:
		comment := p.parseComment()
		rule.Value = comment
		return rule
	case token.IDENT:

	default:
		p.errorTokenExpected(p.pos, token.IDENT)
		return nil
	}

	rule.Name = &ast.Identifier{Pos: p.pos, Name: p.currliteral}
	p.next()

	var entry ast.Node

	tok := p.currToken
	switch tok {
	case token.ASSIGN:
		p.next()
		entry = p.parseEntry(p.currToken.Precedence())
		defer func() {
			err := p.environment.Add(rule.Name.Name, &entry)

			// Since the only error returned is ErrSymbolExists, check for that and
			// append a error that type is already decalred.
			if err == env.ErrSymbolExists {
				p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("identifier %s already declared", rule.Name.Name), p.pos, p.pos))
			}
		}()

	case token.TYPE_CHOICE_ASSIGN, token.GROUP_CHOICE_ASSIGN:
		p.next()
		entry = p.parseEntry(token.LOWEST)
	default:
		p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("expected assigment operators =, /= or //= after identifer `%s`", rule.Name.Name), rule.Name.Pos, rule.Name.Pos))
		return nil
	}
	rule.Value = entry
	if p.peekToken == token.COMMENT && isSameLineTokens(p.pos, p.peekPos) {
		p.next()
		rule.TrailingComment = p.parseInnerComment()
		p.next()
	}

	return rule
}

func (p *Parser) parseEntry(precedence int) ast.Node {
	var exp ast.Node
	if p.currToken == token.COMMENT {
		_ = p.parseComment()
		p.next()
	}
	nudFn := p.nuds[p.currToken]
	if p.currToken == token.IDENT && p.peekToken == token.COLON {
		nudFn = p.parseIdentifier
	}
	if nudFn == nil {
		p.errorUnexpectedPrefix(p.pos, p.currToken)
		return nil
	}
	if p.currToken == token.ONE_OR_MORE || p.currToken == token.ZERO_OR_MORE {
		exp = &ast.UintLiteral{Pos: p.pos, Literal: 0}
	} else {
		exp = nudFn()
	}

	for p.currToken != token.COMMA && precedence < p.peekToken.Precedence() {
		ledFn := p.leds[p.peekToken]
		if ledFn == nil {
			return exp
		}
		p.next()
		exp = ledFn(exp)
	}

	return exp
}

func (p *Parser) parseIdentifier() ast.Node {
	return &ast.Identifier{Pos: p.pos, Name: p.currliteral}
}

func (p *Parser) parseNamedIdentifier() ast.Node {
	literal := p.currliteral
	pos := p.pos
	if literal[0] != '$' {
		p.tasks = append(p.tasks, func() {
			if !p.environment.Exists(literal) {
				p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("identifier %s referenced does not exist", literal), pos, pos))
			}
		})
	}

	return &ast.Identifier{Pos: p.pos, Name: p.currliteral}
}

func (p *Parser) parseBooleanType() ast.Node {
	return &ast.BooleanType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseTstrType() ast.Node {
	return &ast.TstrType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseFloatType() ast.Node {
	return &ast.FloatType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseUintType() ast.Node {
	if p.currToken.IsLiteral(p.currliteral) {
		return p.parseUintLiteral()
	}
	return &ast.UintType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseUintLiteral() *ast.UintLiteral {
	if p.currToken.IsLiteral(p.currliteral) {
		lit, err := strconv.ParseUint(p.currliteral, 0, 64)
		if err != nil {
			p.errorTokenExpected(p.pos, token.INT)
			return nil
		}
		return &ast.UintLiteral{
			Pos:     p.pos,
			Token:   p.currToken,
			Literal: lit,
		}
	}
	return nil
}

func (p *Parser) parseIntegerType() ast.Node {
	if p.currToken.IsLiteral(p.currliteral) {
		return p.parseIntegerLiteral()
	}
	return &ast.IntegerType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseBstrType() ast.Node {
	return &ast.BstrType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseBytesType() ast.Node {
	return &ast.BytesType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseNullType() ast.Node {
	return &ast.NullType{Pos: p.pos, Token: p.currToken}
}

func (p *Parser) parseColon(left ast.Node) ast.Node {
	var ident *ast.Identifier
	switch val := left.(type) {
	case *ast.Identifier:
		ident = val
	case *ast.IntegerLiteral:
		ident = &ast.Identifier{
			Pos:  p.pos,
			Name: fmt.Sprintf("%d", val.Literal),
		}
	default:
		p.errorUnsupportedTypes(p.pos, p.currliteral, token.IDENT, token.INT)
		return nil
	}
	rule := &ast.Rule{
		Pos:  p.pos,
		Name: ident,
	}
	p.next()

	rule.Value = p.parseEntry(p.currToken.Precedence())
	if p.peekToken == token.COMMENT && isSameLineTokens(p.pos, p.peekPos) {
		p.next()
		rule.TrailingComment = p.parseInnerComment()
	}

	return rule
}

func (p *Parser) parseComma(left ast.Node) ast.Node {
	return left
}

func (p *Parser) parseTypeChoice(left ast.Node) ast.Node {
	tc := &ast.TypeChoice{
		Pos:   p.pos,
		Token: p.currToken,
		First: left,
	}
	p.next()
	tc.Second = p.parseEntry(p.currToken.Precedence())
	return tc
}

func (p *Parser) parseUnwrap() ast.Node {
	un := &ast.Unwrap{
		Pos:   p.pos,
		Token: p.currToken,
	}

	p.next()
	un.Item = p.parseEntry(p.currToken.Precedence())

	return un
}

func (p *Parser) parseTag() ast.Node {
	tag := &ast.Tag{
		Pos:   p.pos,
		Token: p.currToken,
	}
	if p.peekToken != token.FLOAT {
		p.errorTokenExpected(p.pos, token.FLOAT)
		return nil
	}
	p.next()
	lit := p.currliteral
	if lit[0] != '6' {
		p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("Major tag 6 expected instead got %s", string(lit[0])), p.pos, p.pos))
		return nil
	}
	tag.Major = &ast.UintLiteral{Literal: 6}
	if lit[1] != '.' {
		p.diagnostics = append(p.diagnostics, NewError("parser", "expected tag in the format n.nnn", p.pos, p.pos))
		return nil
	}
	nnn := lit[2:]
	nnnU, err := strconv.ParseUint(nnn, 0, 64)
	if err != nil {
		p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("incvalid tag number %s", nnn), p.pos, p.pos))
		return nil
	}
	tag.TagNumber = &ast.UintLiteral{Literal: nnnU}
	p.next()
	if !p.expect(token.LPAREN) {
		return nil
	}
	tag.Item = p.parseEntry(p.currToken.Precedence())
	p.next()

	return tag
}

func (p *Parser) parseOptional() ast.Node {
	tc := &ast.Optional{
		Pos:   p.pos,
		Token: p.currToken,
	}
	p.next()

	tc.Item = p.parseEntry(p.currToken.Precedence())
	return tc
}

func (p *Parser) parseZMOccurrence() ast.Node {
	tc := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
		N:     &ast.UintLiteral{Literal: 0},
	}

	tc.Item = p.parseEntry(p.currToken.Precedence())
	return tc
}

func (p *Parser) parseOMOccurrence() ast.Node {
	tc := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
		N:     &ast.UintLiteral{Literal: 1},
	}

	tc.Item = p.parseEntry(p.currToken.Precedence())
	return tc
}

func (p *Parser) parseSizeOperator(left ast.Node) ast.Node {
	sop := &ast.SizeOperatorControl{
		Pos:   p.pos,
		Token: p.currToken,
		Type:  left,
	}

	if p.expectIntLiteral2() {
		size := p.parseUintLiteral()
		sop.Size = size
		return sop
	} else {
		p.errorTokenExpected(p.pos, token.INT)
	}

	return sop
}

func (p *Parser) parseRegexp(left ast.Node) ast.Node {
	var base *ast.TstrType
	if b, ok := left.(*ast.TstrType); ok {
		base = b
	} else {
		return nil
	}
	r := &ast.Regexp{
		Pos:   p.pos,
		Token: p.currToken,
		Base:  base,
	}

	p.next()
	if !p.expect(token.TEXT_LITERAL) {
		return nil
	}
	r.Regex = p.parseEntry(p.currToken.Precedence())
	return r
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

func (p *Parser) parseTextLiteral() ast.Node {
	return &ast.TextLiteral{Pos: p.pos, Token: p.currToken, Literal: p.currliteral}
}

func (p *Parser) parseGroup() ast.Node {
	g := &ast.Group{}
	g.Pos = p.pos
	p.next()

	for p.currToken != token.RPAREN {
		rule := p.parseEntry(p.currToken.Precedence())
		if rule != nil {
			g.Rules = append(g.Rules, rule)

		}
		p.next()
	}
	return g
}

func isSameLineTokens(tok1, tok2 token.Position) bool {
	if tok1.Line == tok2.Line {
		return true
	}
	return false
}

func (p *Parser) parseMap() ast.Node {
	g := &ast.Map{}
	g.Pos = p.pos
	p.next()

	for p.currToken != token.RBRACE {
		rule := p.parseEntry(p.currToken.Precedence())
		if rule != nil {
			g.Rules = append(g.Rules, rule)
		}
		p.next()
	}
	return g
}

func (p *Parser) parseArray() ast.Node {
	arr := &ast.Array{}
	arr.Pos = p.pos
	p.next()

	for p.currToken != token.RBRACK {
		rule := p.parseEntry(p.currToken.Precedence())
		if rule != nil {
			arr.Rules = append(arr.Rules, rule)
		}
		p.next()
	}

	return arr
}

func (p *Parser) parseComment() ast.Node {
	cg := &ast.CommentGroup{}
	cg.List = append(cg.List, &ast.Comment{Pos: p.pos, Text: p.currliteral})

	for p.peekToken == token.COMMENT && p.peekPos.Line == (p.pos.Line+1) {
		p.next()
		cg.List = append(cg.List, p.parseInnerComment())
	}

	if len(cg.List) == 1 {
		return cg.List[0]
	}
	return cg
}

func (p *Parser) parseInnerComment() *ast.Comment {
	comment := &ast.Comment{
		Pos:  p.pos,
		Text: p.currliteral,
	}
	return comment
}

func (p *Parser) parseComparatorOp(left ast.Node) ast.Node {
	var leftI *ast.IntegerType

	switch left.(type) {
	case *ast.UintType, *ast.IntegerType, *ast.FloatType:
	default:
		fmt.Println(left)
		p.errorUnsupportedTypes(p.pos, p.currliteral, token.INT, token.FLOAT, token.UINT)
		return nil
	}

	op := &ast.ComparatorOpControl{
		Pos:      p.pos,
		Token:    p.currToken,
		Left:     leftI,
		Operator: p.currliteral,
	}
	if !p.peekToken.IsNumeric() {
		p.errorUnsupportedType(p.peekPos, p.currliteral, token.INT)
	}
	p.next()

	right := p.parseEntry(p.currToken.Precedence())
	op.Right = right

	return op
}

func (p *Parser) parseBound(left ast.Node) ast.Node {
	var bound *ast.Bound
	switch val := left.(type) {
	case *ast.IntegerLiteral:
		bound = p.parseIntBound(val)
	case *ast.FloatLiteral:
		bound = p.parseFloatBound(val)
	case *ast.Identifier:
		bound = p.parseIdentBound(val)
	}
	return bound
}

func (p *Parser) parseOccurrence(left ast.Node) ast.Node {
	occ := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
	}
	var right *ast.UintLiteral

	leftU, ok := left.(*ast.UintLiteral)
	if !ok {
		p.errorUnsupportedType(left.Start(), p.currliteral, token.UINT)
		return nil
	}
	if p.peekToken == token.UINT && p.peekToken.IsLiteral(p.peekLiteral) {
		p.next()
		rightN := p.parseUintType()

		if val, ok := rightN.(*ast.UintLiteral); ok {
			right = val
		}
	} else {
		// assume no upper limit. Advance
		p.next()
	}

	occ.N = leftU
	occ.M = right
	occ.Item = p.parseEntry(p.currToken.Precedence())
	return occ
}

func (p *Parser) parseIdentBound(left *ast.Identifier) *ast.Bound {
	b := &ast.Bound{
		Pos:   p.pos,
		Token: p.currToken,
		From:  left,
	}

	p.next()
	b.To = p.parseEntry(p.currToken.Precedence())

	p.tasks = append(p.tasks, func() {
		valLeft := p.environment.Get(left.Name)
		to := b.To
		switch val := to.(type) {
		case *ast.Identifier:
			valRight := p.environment.Get(val.Name)
			if !(reflect.TypeOf(valLeft) == reflect.TypeOf(valRight)) {
				p.diagnostics = append(p.diagnostics, NewError("parser",
					fmt.Sprintf("operator %s expected same type min, max values. The values of %s and %s resolve to %+v and %+v", b.Token, left.Name, val.Name, *valLeft, valRight),
					left.Pos, val.Pos),
				)
			}
		}

	})
	return b
}
func (p *Parser) parseIntBound(left *ast.IntegerLiteral) *ast.Bound {
	b := &ast.Bound{
		Pos:   p.pos,
		Token: p.currToken,
		From:  left,
	}

	p.next()
	b.To = p.parseEntry(p.currToken.Precedence())

	return b
}

func (p *Parser) parseFloatBound(left *ast.FloatLiteral) *ast.Bound {
	return nil
}

func (p *Parser) errorUnsupportedType(pos token.Position, operator string, supported token.Token) {
	p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("operator %s only supports token %s", operator, supported), pos, pos))
}

func (p *Parser) errorUnsupportedTypes(pos token.Position, operator string, supported ...token.Token) {
	toks := []string{}
	if len(supported) == 0 {
		return
	}
	for _, tok := range supported {
		toks = append(toks, tok.String())
	}
	p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("operator %s only supports tokens %s", operator, strings.Join(toks, ", ")), pos, pos))
}

func (p *Parser) errorTokenExpected(pos token.Position, tok token.Token) {
	p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("expected %s at line %d, column %d", tok.String(), pos.Line, pos.Column), pos, pos))
}

func (p *Parser) errorUnexpectedPrefix(pos token.Position, tok token.Token) {
	p.diagnostics = append(p.diagnostics, NewError("parser", fmt.Sprintf("unexpected token %s at line %d, column %d", tok.String(), pos.Line, pos.Column), pos, pos))
}

func (p *Parser) next() {
	p.currToken = p.peekToken
	p.currliteral = p.peekLiteral
	p.pos = p.peekPos
	p.peekToken, p.peekPos, p.peekLiteral = p.lexer.Scan()
}

func NewParser(lexer *lexer.Lexer, opts ...ParserConfig) *Parser {
	p := &Parser{}
	p.lexer = lexer

	p.nuds = make(map[token.Token]nudParseFn)
	p.leds = make(map[token.Token]ledParseFn)

	p.environment = env.NewEnvironment()

	// Register token handlers
	p.nuds[token.IDENT] = p.parseNamedIdentifier
	p.nuds[token.BOOL] = p.parseBooleanType
	p.nuds[token.TSTR] = p.parseTstrType
	p.nuds[token.TEXT] = p.parseTstrType
	p.nuds[token.TEXT_LITERAL] = p.parseTextLiteral
	p.nuds[token.FLOAT] = p.parseFloatType
	p.nuds[token.FLOAT16] = p.parseFloatType
	p.nuds[token.FLOAT32] = p.parseFloatType
	p.nuds[token.FLOAT64] = p.parseFloatType
	p.nuds[token.UINT] = p.parseUintType
	p.nuds[token.INT] = p.parseIntegerType
	p.nuds[token.BSTR] = p.parseBstrType
	p.nuds[token.BYTES] = p.parseBytesType
	p.nuds[token.NULL] = p.parseNullType
	p.nuds[token.NIL] = p.parseNullType
	p.nuds[token.LBRACE] = p.parseMap
	p.nuds[token.LPAREN] = p.parseGroup
	p.nuds[token.LBRACK] = p.parseArray
	p.nuds[token.COMMENT] = p.parseComment
	p.nuds[token.OPTIONAL] = p.parseOptional
	p.nuds[token.ZERO_OR_MORE] = p.parseZMOccurrence
	p.nuds[token.ONE_OR_MORE] = p.parseOMOccurrence
	p.nuds[token.UNWRAP] = p.parseUnwrap
	p.nuds[token.HASH] = p.parseTag

	p.leds[token.COLON] = p.parseColon
	p.leds[token.TYPE_CHOICE] = p.parseTypeChoice

	// Control Operators
	for _, tok := range []token.Token{token.LT, token.LE, token.GT, token.GE, token.EQ, token.NE} { // Comparable control operators
		p.leds[tok] = p.parseComparatorOp
	}

	p.leds[token.SIZE] = p.parseSizeOperator
	p.leds[token.REGEXP] = p.parseRegexp

	p.leds[token.INCLUSIVE_BOUND] = p.parseBound
	p.leds[token.EXCLUSIVE_BOUND] = p.parseBound
	p.leds[token.ZERO_OR_MORE] = p.parseOccurrence
	p.leds[token.ONE_OR_MORE] = p.parseOccurrence
	p.leds[token.ARROW_MAP] = p.parseColon
	p.leds[token.COMMA] = p.parseComma
	p.next()
	p.next()

	return p
}
