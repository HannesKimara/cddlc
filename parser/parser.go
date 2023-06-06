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

type taskFn func() error

type nudParseFn func() (ast.Node, error)
type ledParseFn func(ast.Node) (ast.Node, error)

// ParseConfig contains the configuration options for the parser. Passed on call to NewParser function.
type Config struct{}

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

	// handles mapping of used symbols to objects
	environment *env.Environment

	// the error convenience function
	error func(string, token.Position, token.Position) error

	// hold tasks to be run after the completed ast build.
	// used mostly to check types in type specific operators that may not exist in the environment at first pass
	tasks []taskFn
}

// Parses the current file and build the AST from the top. Returns an instance reference of ast.CDDL.
func (p *Parser) Parse() (*ast.CDDL, error) {
	cddl := &ast.CDDL{}
	cddl.Rules = []ast.CDDLEntry{}

	// Don't parse if lexer errors are non zero
	if len(p.lexer.Errors) != 0 {
		return nil, nil
	}

	for p.currToken != token.EOF {
		cddlEntry, err := p.parseRule()
		if err != nil {
			return cddl, err // TODO: Collect or channel errors
		}
		if cddlEntry != nil {
			cddl.Rules = append(cddl.Rules, cddlEntry)
		}
		p.next()
	}

	for _, task := range p.tasks {
		err := task() // TODO: Collect or channel errors
		if err != nil {
			return cddl, err
		}
	}

	return cddl, nil

}

func (p *Parser) expect(tok token.Token) bool {
	if p.currToken != tok {
		p.errorTokenExpected(p.pos, tok)
		return false
	}
	p.next()
	return true
}

func (p *Parser) expectPeek(tok token.Token) bool {
	if p.peekToken != tok {
		p.errorTokenExpected(p.peekPos, tok)
		return false
	}
	p.next()
	return false
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

func (p *Parser) parseRule() (_ ast.CDDLEntry, err error) {
	rule := &ast.Rule{}

	switch p.currToken {
	case token.COMMENT:
		comment, err := p.parseComment()
		if err != nil {
			return rule, err
		}
		cast := comment.(ast.CDDLEntry)
		return cast, nil
	case token.IDENT:

	default:
		return nil, p.errorTokenExpected(p.pos, token.IDENT)
	}

	rule.Name = &ast.Identifier{Pos: p.pos, Name: p.currliteral}
	p.next()

	var entry ast.Node

	tok := p.currToken
	switch tok {
	case token.ASSIGN:
		p.next()
		entry, err = p.parseEntry(p.currToken.Precedence())
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
		entry, err = p.parseEntry(token.LOWEST)
		if err != nil {
			return rule, err
		}
	default:
		return nil, NewError("parser", fmt.Sprintf("expected assigment operators =, /= or //= after identifer `%s`", rule.Name.Name), rule.Name.Pos, rule.Name.Pos)
	}
	rule.Value = entry
	if p.peekToken == token.COMMENT && isSameLineTokens(p.pos, p.peekPos) {
		p.next()
		rule.TrailingComment = p.parseInnerComment()
		p.next()
	}

	return rule, nil
}

// ParseEntryShould returns a parsed entry if of expected value else returns the error
// TODO(HannesKimara): func(p *Parser) parseEntryShould(precedence, should ast.Node) (ast.Node, error)

func (p *Parser) parseEntry(precedence int) (ast.Node, error) {
	// TODO(HannesKimara): in error handling return an ast.BadNode that encapsulates the section when returning the error
	var exp ast.Node
	if p.currToken == token.COMMENT { // TODO(HannesKimara): use this :- commentgroup, preceding and trailing comments
		_, _ = p.parseComment()
		p.next()
	}
	nudFn := p.nuds[p.currToken]
	if p.currToken == token.IDENT && p.peekToken == token.COLON {
		nudFn = p.parseIdentifier
	}
	if nudFn == nil {
		err := p.errorUnexpectedPrefix(p.pos, p.currToken)
		return nil, err
	}
	if p.currToken == token.ONE_OR_MORE || p.currToken == token.ZERO_OR_MORE {
		exp = &ast.UintLiteral{Pos: p.pos, Literal: 0}
	} else {
		expR, err := nudFn()
		exp = expR
		if err != nil {
			return exp, err
		}
	}

	for p.currToken != token.COMMA && precedence < p.peekToken.Precedence() {
		ledFn := p.leds[p.peekToken]
		if ledFn == nil {
			return exp, nil
		}
		p.next()
		expR, err := ledFn(exp)
		exp = expR
		if err != nil {
			return exp, err
		}
	}

	return exp, nil
}

func (p *Parser) parseIdentifier() (ast.Node, error) {
	return &ast.Identifier{Pos: p.pos, Name: p.currliteral}, nil
}

func (p *Parser) parseNamedIdentifier() (ast.Node, error) {
	literal := p.currliteral
	pos := p.pos
	if literal[0] != '$' {
		p.tasks = append(p.tasks, func() error {
			if !p.environment.Exists(literal) {
				return NewError("parser", fmt.Sprintf("identifier %s referenced does not exist", literal), pos, pos)
			}
			return nil
		})
	}

	return &ast.Identifier{Pos: p.pos, Name: p.currliteral}, nil
}

func (p *Parser) parseBooleanType() (ast.Node, error) {
	ret := &ast.BooleanType{Pos: p.pos, Token: p.currToken}

	return ret, nil
}

func (p *Parser) parseTstrType() (ast.Node, error) {
	return &ast.TstrType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseFloatType() (ast.Node, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		return p.parseFloatLiteral()
	}
	return &ast.FloatType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseFloatLiteral() (ast.Node, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		lit, err := strconv.ParseFloat(p.currliteral, 64)
		if err != nil {
			return nil, p.errorTokenExpected(p.pos, token.FLOAT)
		}
		return &ast.FloatLiteral{
			Pos:     p.pos,
			Token:   p.currToken,
			Literal: lit,
		}, nil
	}
	return nil, NewError("parser", "expected float literal", p.pos, p.pos)
}

func (p *Parser) parseUintType() (ast.Node, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		return p.parseUintLiteral()
	}
	return &ast.UintType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseUintLiteral() (*ast.UintLiteral, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		lit, err := strconv.ParseUint(p.currliteral, 0, 64)
		if err != nil {
			return nil, p.errorTokenExpected(p.pos, token.UINT)
		}
		return &ast.UintLiteral{
			Pos:     p.pos,
			Token:   p.currToken,
			Literal: lit,
		}, nil
	}
	return nil, NewError("parser", "expected uint literal", p.pos, p.pos)
}

func (p *Parser) parseIntegerType() (ast.Node, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		return p.parseIntegerLiteral()
	}
	return &ast.IntegerType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseNegativeIntegerType() (ast.Node, error) {
	return &ast.NegativeIntegerType{Pos: p.pos, Token: p.currToken}, nil
}
func (p *Parser) parseBstrType() (ast.Node, error) {
	return &ast.BstrType{Pos: p.pos, Token: p.currToken}, nil
}

// TODO: Consider joining bstr and bytes, to singular ast and parser
func (p *Parser) parseBytesType() (ast.Node, error) {
	return &ast.BytesType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseNullType() (ast.Node, error) {
	return &ast.NullType{Pos: p.pos, Token: p.currToken}, nil
}

func (p *Parser) parseColon(left ast.Node) (ast.Node, error) {
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
		err := p.errorUnsupportedTypes(p.pos, p.currliteral, token.IDENT, token.INT)
		return nil, err
	}
	rule := &ast.Entry{
		Pos:  p.pos,
		Name: ident,
	}
	p.next()

	val, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return val, err
	}
	rule.Value = val
	if p.peekToken == token.COMMENT && isSameLineTokens(p.pos, p.peekPos) {
		p.next()
		rule.TrailingComment = p.parseInnerComment()
	}

	return rule, nil
}

func (p *Parser) parseComma(left ast.Node) (ast.Node, error) {
	return left, nil
}

func (p *Parser) parseTypeChoice(left ast.Node) (ast.Node, error) {
	tc := &ast.TypeChoice{
		Pos:   p.pos,
		Token: p.currToken,
		First: left,
	}
	p.next()
	sec, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return sec, err
	}
	tc.Second = sec
	return tc, nil
}

func (p *Parser) parseUnwrap() (ast.Node, error) {
	un := &ast.Unwrap{
		Pos:   p.pos,
		Token: p.currToken,
	}

	p.next()
	item, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return item, err
	}
	un.Item = item

	return un, nil
}

func (p *Parser) parseTag() (ast.Node, error) {
	tagBase := &ast.Tag{
		Pos:   p.pos,
		Token: p.currToken,
	}

	switch p.peekToken {
	case token.INT: // tag like #6
		p.next()
		major, err := strconv.ParseUint(p.currliteral, 0, 64)
		if err != nil {
			return nil, NewError("parser", "failed to parse tag major", p.pos, p.pos)
		}
		tagBase.Major = &ast.UintLiteral{Literal: major}
	case token.FLOAT:
		p.next()
		tagB, err := p.parseFloatTag()
		if err != nil {
			return nil, err
		}
		tagBase.Major = tagB.Major
		tagBase.TagNumber = tagB.TagNumber
	}
	if p.peekToken == token.LPAREN {
		p.next()
		p.next()
		item, err := p.parseEntry(p.currToken.Precedence())
		tagBase.Item = item
		if err != nil {
			return tagBase, err
		}
		p.expectPeek(token.RPAREN)
	}
	return tagBase, nil
}

func (p *Parser) parseFloatTag() (*ast.Tag, error) {
	tag := &ast.Tag{}
	sections := strings.Split(p.currliteral, ".")
	if len(sections) != 2 {
		return nil, p.error(fmt.Sprintf("invalid tag %s", p.currliteral), p.pos, p.pos)
	}

	majorUint, err := strconv.ParseUint(sections[0], 0, 64)
	if err != nil {
		return nil, p.error(err.Error(), p.pos, p.pos)
	}

	numberUint, err := strconv.ParseUint(sections[1], 0, 64)
	if err != nil {
		return nil, p.error(err.Error(), p.pos, p.pos)
	}

	tag.Major = &ast.UintLiteral{Pos: p.pos, Token: p.currToken, Literal: majorUint}
	tag.TagNumber = &ast.UintLiteral{Pos: p.pos, Token: p.currToken, Literal: numberUint}

	return tag, nil
}

func (p *Parser) parseOptional() (ast.Node, error) {
	tc := &ast.Optional{
		Pos:   p.pos,
		Token: p.currToken,
	}
	p.next()

	item, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return item, err
	}
	tc.Item = item

	return tc, nil
}

func (p *Parser) parseZMOccurrence() (ast.Node, error) {
	tc := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
		N:     &ast.UintLiteral{Literal: 0},
	}

	item, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return item, err
	}

	tc.Item = item
	return tc, nil
}

func (p *Parser) parseOMOccurrence() (ast.Node, error) {
	tc := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
		N:     &ast.UintLiteral{Literal: 1},
	}

	item, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return tc, err
	}
	tc.Item = item
	return tc, nil
}

func (p *Parser) parseSizeOperator(left ast.Node) (ast.Node, error) {
	sop := &ast.SizeOperatorControl{
		Pos:   p.pos,
		Token: p.currToken,
	}

	switch val := left.(type) {
	case *ast.BstrType, *ast.UintType, *ast.TstrType:
		sop.Type = val
	default:
		err := p.errorUnsupportedTypes(sop.Pos, p.currliteral, token.TSTR, token.BSTR, token.UINT)
		return sop, err
	}
	p.next()

	right, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return sop, err
	}
	sop.Size = right

	return sop, nil
}

// TODO :: Evaluate that the regex is valid and compiles according to
// https://www.rfc-editor.org/rfc/rfc8610#section-3.8.3
func (p *Parser) parseRegexp(left ast.Node) (ast.Node, error) {
	var base *ast.TstrType
	if b, ok := left.(*ast.TstrType); ok {
		base = b
	} else {
		return nil, p.errorUnsupportedTypes(b.Pos, p.currliteral, token.TSTR)
	}
	r := &ast.Regexp{
		Pos:   p.pos,
		Token: p.currToken,
		Base:  base,
	}
	if p.peekToken != token.TEXT_LITERAL {
		return r, p.errorTokenExpected(p.pos, token.TEXT_LITERAL)
	}
	p.next()
	regex, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return r, err
	}
	r.Regex = regex
	return r, nil
}

func (p *Parser) parseIntegerLiteral() (*ast.IntegerLiteral, error) {
	if p.currToken.IsLiteral(p.currliteral) {
		lit, err := strconv.ParseInt(p.currliteral, 0, 64)
		if err != nil {
			return nil, p.errorTokenExpected(p.pos, token.INT)
		}
		return &ast.IntegerLiteral{
			Pos:     p.pos,
			Token:   p.currToken,
			Literal: lit,
		}, nil
	}
	return nil, NewError("parser", "expected integer literal", p.pos, p.pos)
}

func (p *Parser) parseTextLiteral() (ast.Node, error) {
	return &ast.TextLiteral{Pos: p.pos, Token: p.currToken, Literal: p.currliteral}, nil
}

func (p *Parser) parseGroup() (ast.Node, error) {
	g := &ast.Group{}
	g.Pos = p.pos
	p.next()

	for p.currToken != token.RPAREN {
		rawEntry, err := p.parseEntry(p.currToken.Precedence())
		if err != nil {
			return g, err
		}
		var entry ast.GroupEntry

		if castRule, ok := rawEntry.(ast.GroupEntry); ok {
			entry = castRule
		}
		if entry != nil {
			g.Entries = append(g.Entries, entry)

		}
		p.next()
	}
	return g, nil
}

func isSameLineTokens(tok1, tok2 token.Position) bool {
	return tok1.Line == tok2.Line
}

func (p *Parser) parseMap() (ast.Node, error) {
	g := &ast.Map{}
	g.Pos = p.pos
	p.next()

	for p.currToken != token.RBRACE {
		rule, err := p.parseEntry(p.currToken.Precedence())
		if err != nil {
			return g, err
		}
		if rule != nil {
			g.Rules = append(g.Rules, rule)
		}
		p.next()
	}
	return g, nil
}

func (p *Parser) parseArray() (ast.Node, error) {
	arr := &ast.Array{}
	arr.Pos = p.pos
	p.next()

	for p.currToken != token.RBRACK {
		rule, err := p.parseEntry(p.currToken.Precedence())
		if err != nil {
			return arr, err
		}
		if rule != nil {
			arr.Rules = append(arr.Rules, rule)
		}
		p.next()
	}

	return arr, nil
}

func (p *Parser) parseComment() (ast.Node, error) {
	cg := &ast.CommentGroup{}
	cg.List = append(cg.List, &ast.Comment{Pos: p.pos, Text: p.currliteral})

	for p.peekToken == token.COMMENT && p.peekPos.Line == (p.pos.Line+1) {
		p.next()
		cg.List = append(cg.List, p.parseInnerComment())
	}

	if len(cg.List) == 1 {
		return cg.List[0], nil
	}
	return cg, nil
}

func (p *Parser) parseInnerComment() *ast.Comment {
	comment := &ast.Comment{
		Pos:  p.pos,
		Text: p.currliteral,
	}
	return comment
}

func (p *Parser) parseComparatorOp(left ast.Node) (ast.Node, error) {
	var leftI *ast.IntegerType

	switch left.(type) {
	case *ast.UintType, *ast.IntegerType, *ast.FloatType:
	default:
		return nil, p.errorUnsupportedTypes(p.pos, p.currliteral, token.INT, token.FLOAT, token.UINT)
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

	right, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return &ast.BadNode{Base: leftI, Token: p.currToken}, err
	}
	op.Right = right

	return op, nil
}

func (p *Parser) parseBound(left ast.Node) (bound ast.Node, err error) {
	bound = &ast.Range{}

	switch val := left.(type) {
	case *ast.IntegerLiteral:
		bound, err = p.parseIntBound(val)
	case *ast.FloatLiteral:
		bound, err = p.parseFloatBound(val)
	case *ast.Identifier:
		bound, err = p.parseIdentBound(val)
	}
	return bound, err
}

func (p *Parser) parseOccurrence(left ast.Node) (ast.Node, error) {
	occ := &ast.NMOccurrence{
		Pos:   p.pos,
		Token: p.currToken,
	}
	var right *ast.UintLiteral

	leftU, ok := left.(*ast.UintLiteral)
	if !ok {
		return nil, p.errorUnsupportedType(left.Start(), p.currliteral, token.UINT)
	}
	if p.peekToken == token.UINT && p.peekToken.IsLiteral(p.peekLiteral) {
		p.next()
		rightN, err := p.parseUintType()
		if err != nil {
			return occ, err
		}

		if val, ok := rightN.(*ast.UintLiteral); ok {
			right = val
		}
	} else {
		// assume no upper limit. Advance
		p.next()
	}

	occ.N = leftU
	occ.M = right
	item, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return &ast.BadNode{Base: occ, Token: p.currToken}, err
	}
	occ.Item = item
	return occ, nil
}

func (p *Parser) parseIdentBound(left *ast.Identifier) (*ast.Range, error) {
	b := &ast.Range{
		Pos:   p.pos,
		Token: p.currToken,
		From:  left,
	}

	p.next()
	to, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return b, err
	}
	b.To = to

	p.tasks = append(p.tasks, func() error {
		valLeft := p.environment.Get(left.Name)
		to := b.To
		switch val := to.(type) {
		case *ast.Identifier:
			valRight := p.environment.Get(val.Name)
			if !(reflect.TypeOf(valLeft) == reflect.TypeOf(valRight)) {
				return NewError("parser",
					fmt.Sprintf("operator %s expected same type min, max values. The values of %s and %s resolve to %+v and %+v", b.Token, left.Name, val.Name, *valLeft, valRight),
					left.Pos, val.Pos)
			}
		}
		return nil
	})
	return b, nil
}
func (p *Parser) parseIntBound(left *ast.IntegerLiteral) (*ast.Range, error) {
	b := &ast.Range{
		Pos:   p.pos,
		Token: p.currToken,
		From:  left,
	}

	p.next()
	to, err := p.parseEntry(p.currToken.Precedence())
	if err != nil {
		return b, err
	}
	switch right := to.(type) {
	case *ast.IntegerLiteral:
		b.To = right
	case *ast.Identifier:

	}

	return b, nil
}

func (p *Parser) parseFloatBound(left *ast.FloatLiteral) (*ast.Range, error) {
	return nil, nil
}

func (p *Parser) errorUnsupportedType(pos token.Position, operator string, supported token.Token) error {
	return NewError("parser", fmt.Sprintf("operator %s only supports token %s", operator, supported), pos, pos)
}

func (p *Parser) errorUnsupportedTypes(pos token.Position, operator string, supported ...token.Token) error {
	toks := []string{}
	if len(supported) == 0 {
		panic("parser internal error: construction of unsupportedTypes error with no supported alternatives")
	}
	for _, tok := range supported {
		toks = append(toks, tok.String())
	}
	return NewError("parser", fmt.Sprintf("operator %s only supports tokens %s", operator, strings.Join(toks, ", ")), pos, pos)
}

func (p *Parser) errorTokenExpected(pos token.Position, tok token.Token) error {
	return NewError("parser", fmt.Sprintf("expected %s at line %d, column %d", tok.String(), pos.Line, pos.Column), pos, pos)
}

func (p *Parser) errorUnexpectedPrefix(pos token.Position, tok token.Token) error {
	return NewError("parser", fmt.Sprintf("unexpected token %s at line %d, column %d", tok.String(), pos.Line, pos.Column), pos, pos)
}

func (p *Parser) next() {
	p.currToken = p.peekToken
	p.currliteral = p.peekLiteral
	p.pos = p.peekPos
	p.peekToken, p.peekPos, p.peekLiteral = p.lexer.Scan()
}

func (p *Parser) registerNud(tok token.Token, fn nudParseFn) {
	if _, ok := p.nuds[tok]; ok {
		panic(fmt.Sprintf("parser internal error: multiple registrations for token %s in the same block", tok))
	}
	p.nuds[tok] = fn
}

func NewParser(lexer *lexer.Lexer, opts ...Config) *Parser {
	p := &Parser{}
	p.lexer = lexer

	p.nuds = make(map[token.Token]nudParseFn)
	p.leds = make(map[token.Token]ledParseFn)

	p.environment = env.NewEnvironment()
	p.error = func(msg string, start, end token.Position) error {
		return NewError("parser", msg, start, end)
	}

	// Register token handlers
	p.registerNud(token.IDENT, p.parseNamedIdentifier)
	p.registerNud(token.BOOL, p.parseBooleanType)
	p.registerNud(token.TSTR, p.parseTstrType)
	p.registerNud(token.TEXT, p.parseTstrType)
	p.registerNud(token.TEXT_LITERAL, p.parseTextLiteral)
	p.registerNud(token.FLOAT, p.parseFloatType)
	p.registerNud(token.FLOAT16, p.parseFloatType)
	p.registerNud(token.FLOAT32, p.parseFloatType)
	p.registerNud(token.FLOAT64, p.parseFloatType)
	p.registerNud(token.UINT, p.parseUintType)
	p.registerNud(token.INT, p.parseIntegerType)
	p.registerNud(token.NINT, p.parseNegativeIntegerType)
	p.registerNud(token.BSTR, p.parseBstrType)
	p.registerNud(token.BYTES, p.parseBytesType)
	p.registerNud(token.NULL, p.parseNullType)
	p.registerNud(token.NIL, p.parseNullType)
	p.registerNud(token.LBRACE, p.parseMap)
	p.registerNud(token.LPAREN, p.parseGroup)
	p.registerNud(token.LBRACK, p.parseArray)
	p.registerNud(token.COMMENT, p.parseComment)
	p.registerNud(token.OPTIONAL, p.parseOptional)
	p.registerNud(token.ZERO_OR_MORE, p.parseZMOccurrence)
	p.registerNud(token.ONE_OR_MORE, p.parseOMOccurrence)
	p.registerNud(token.UNWRAP, p.parseUnwrap)
	p.registerNud(token.HASH, p.parseTag)

	p.leds[token.COLON] = p.parseColon
	p.leds[token.TYPE_CHOICE] = p.parseTypeChoice

	// Comparable control operators
	for _, tok := range []token.Token{token.LT, token.LE, token.GT, token.GE, token.EQ, token.NE} {
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
