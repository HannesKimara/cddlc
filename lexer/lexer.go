// Package Lexer implements a scanner/tokenizer for the CDDL source
// Tokens are produced by repeated calls to Next() method
package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/flowfunction/cddl/token"
)

const (
	EOF = -1
)

type ErrorFunction func(pos token.Position, msg string)

type TokenSer struct {
	Pos     token.Position `json:"pos"`
	Token   token.Token    `json:"token"`
	Literal string         `json:"literal"`
}

type Error struct {
	Pos token.Position
	Msg string
}

type Lexer struct {
	src         []byte
	chr         rune
	offset      int // points to current character
	rdOffset    int // points to the next character(used for peeking)
	ErrCount    int
	Errors      []Error
	lineOffsets []int
}

func (l *Lexer) addLineOffset(offset int) {
	lenOffsets := len(l.lineOffsets)
	prevOffset := l.lineOffsets[lenOffsets-1]
	if offset < prevOffset {
		return
	}

	l.lineOffsets = append(l.lineOffsets, offset)

}

// NewLexer returns a new Lexer given code sourcce in bytes
func NewLexer(src []byte) *Lexer {
	src = append(src, 32) // Hack:: trailing byte ignored without adding an ignorable byte
	l := &Lexer{
		src:      src,
		offset:   0,
		rdOffset: 0,
		ErrCount: 0,
	}

	l.lineOffsets = append(l.lineOffsets, -1)

	l.next()

	return l
}

// Scan returns the next token.
func (l *Lexer) Scan() (tok token.Token, pos token.Position, lit string) {
	l.skipWhitespace()
	startOffset := l.offset

	switch chr := l.chr; {
	case isIdentiferStart(chr):
		lit = l.scanIdentifier()
		// keywords are longer than two characters so avoid lookups for smaller
		// TODO: check if keywords are fixed in spec or could be redefined. Optionally, have this as a compiler flag
		if len(lit) > 2 {
			tok = token.Lookup(lit)
		} else {
			tok = token.IDENT
		}
	case isDigit(chr):
		tok, lit = l.scanNumber()
	default:
		l.next() // consume . Don't have to peek for two character tokens
		lit = string(chr)
		switch chr {
		case '-':
			tok = token.MINUS
		case '?':
			tok = token.OPTIONAL
		case ';':
			tok = token.COMMENT
			lit = l.scanComment()
		case ':':
			tok = token.COLON
		case '(':
			tok = token.LPAREN
		case ')':
			tok = token.RPAREN
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '+':
			tok = token.ONE_OR_MORE
		case '*':
			tok = token.ZERO_OR_MORE
		case '<':
			tok = token.LEFT_ANGLE_BRACKET
		case '>':
			tok = token.RIGHT_ANGLE_BRACKET
		case '#':
			tok = token.HASH
		case '~':
			tok = token.UNWRAP
		case '=':
			if l.chr == '>' {
				l.next()
				tok = token.ARROW_MAP
				lit = "=>"
			} else {
				tok = token.ASSIGN
			}
		case ',':
			tok = token.COMMA
		// case '$': // TODO: Never reached since these are valif identififer start characters
		// 	if l.chr == '$' {
		// 		l.next()
		// 		tok = token.GROUP_SOCKET
		// 		lit = "&&"
		// 	} else {
		// 		tok = token.TYPE_SOCKET
		// 	}
		case '&':
			tok = token.AMPERSAND
		case '/':
			if l.chr == '=' {
				l.next()
				tok = token.TYPE_CHOICE_ASSIGN
				lit = "/="
			} else if l.chr == '/' && l.peek() == '=' {
				l.next()
				l.next()
				tok = token.GROUP_CHOICE_ASSIGN
				lit = "//="
			} else if l.chr == '/' {
				l.next()
				tok = token.GROUP_CHOICE
				lit = "//"
			} else {
				tok = token.TYPE_CHOICE
			}
		case '.':
			if l.chr == '.' && l.peek() == '.' {
				l.next()
				l.next()
				tok = token.EXCLUSIVE_BOUND
				lit = "..."
			} else if l.chr == '.' {
				l.next()
				tok = token.INCLUSIVE_BOUND
				lit = ".."
			} else if unicode.IsLetter(l.chr) {
				// Scan for control operators
				copString := l.scanControlOp()
				lit = fmt.Sprintf("%s%s", string(chr), copString)

				tok = token.Lookup(lit)
			} else {
				tok = token.PERIOD
			}
		case '"':
			tok = token.TEXT_LITERAL
			lit = l.scanString()
		case EOF:
			tok = token.EOF
			lit = ""
		}
	}
	lenLines := len(l.lineOffsets)

	pos = token.Position{
		Offset: startOffset,
		Line:   lenLines,
		Column: startOffset - l.lineOffsets[lenLines-1],
	}

	return

}

func (l *Lexer) next() {
	if l.rdOffset >= len(l.src) {
		l.chr = EOF
		return
	}
	ch, w := utf8.DecodeRune(l.src[l.rdOffset:])
	if ch == utf8.RuneError {
		return
	}
	l.chr = ch
	l.offset = l.rdOffset
	l.rdOffset += w
}

func (l *Lexer) skipWhitespace() {
	for l.chr == ' ' || l.chr == '\r' || l.chr == '\n' || l.chr == '\t' {
		if l.chr == '\n' {
			l.addLineOffset(l.offset)
		}
		l.next()
	}
}

func (l *Lexer) peek() byte {
	if l.rdOffset < len(l.src) {
		return l.src[l.rdOffset]
	}
	return 0
}

func (l *Lexer) scanComment() string {
	offsetPre := l.offset

	// scan to end of line
	// TODO: account for carriage returns
	for l.chr >= 0 && l.chr != '\n' {
		l.next()
	}
	return string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanIdentifier() string {
	offsetPre := l.offset

	for isIdentifierBodyChar(l.chr) {
		l.next()
	}

	return string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanControlOp() string {
	offsetPre := l.offset

	for isLetter(l.chr) {
		l.next()
	}

	return string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanNumber() (token.Token, string) {
	offsetPre := l.offset
	tok := token.ILLEGAL

	// first character already consumed

	// TODO : Scan for hex, octal and binary fractionals
	if l.chr == '0' && isxboDigitStart(rune(l.peek())) {
		l.next()
		switch unicode.ToLower(l.chr) {
		case 'x':
			l.next()
			tok, lit := l.scanHexDigits()
			return tok, lit
		case 'b':
			l.next()
			tok, lit := l.scanBinaryDigits()
			return tok, lit
		case 'o':
			l.next()
			tok, lit := l.scanOctalDigits()
			return tok, lit
		}
	}
	tok = token.INT

	// TODO : Add support for all number types, int, floats
	for unicode.IsDigit(l.chr) || l.chr == '.' && unicode.IsDigit(rune(l.peek())) {
		l.next()
		if l.chr == '.' && l.peek() != '.' { // check that its not part of a bound .., ...
			tok = token.FLOAT
		}
	}
	return tok, string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanHexDigits() (tok token.Token, lit string) {
	offsetPre := l.offset
	for isHex(l.chr) {
		l.next()
	}
	return token.INT, "0x" + string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanBinaryDigits() (tok token.Token, lit string) {
	offsetPre := l.offset
	for isBinary(l.chr) {
		l.next()
	}
	return token.INT, "0b" + string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanOctalDigits() (tok token.Token, lit string) {
	offsetPre := l.offset
	for isOctal(l.chr) {
		l.next()
	}
	return token.INT, "0o" + string(l.src[offsetPre:l.offset])
}

func (l *Lexer) scanString() string {
	offsetPre := l.offset

	for {
		if l.chr < 0 || l.chr == '\n' {
			l.error(l.offset, "unexpected newline character before string termination")
			break
		}
		if l.chr == '"' {
			l.next()
			break
		}
		l.next()
	}

	return string(l.src[offsetPre : l.offset-1])
}

func isLetter(x rune) bool {
	if 'A' <= x && x <= 'Z' || 'a' <= x && x <= 'z' {
		return true
	}
	return false
}

func isHex(x rune) bool {
	if isDigit(x) || 'A' <= x && x <= 'F' || 'a' <= x && x <= 'f' {
		return true
	}
	return false
}

func isOctal(x rune) bool {
	if '0' <= x && x <= '7' {
		return true
	}
	return false
}

func isBinary(x rune) bool {
	if x == '0' || x == '1' {
		return true
	}
	return false
}

func isDigit(x rune) bool {
	if '0' <= x && x <= '9' {
		return true
	}
	return false
}

func isIdentiferStart(x rune) bool {
	if 'A' <= x && x <= 'Z' || 'a' <= x && x <= 'z' || x == '@' || x == '_' || x == '$' {
		return true
	}
	return false
}

func isIdentifierBodyChar(x rune) bool {
	if isIdentiferStart(x) || isDigit(x) || x == '-' || x == '.' {
		return true
	}
	return false
}

func (l *Lexer) error(offset int, message string) {
	l.Errors = append(l.Errors, Error{Pos: token.Position{Offset: offset}, Msg: message})
	l.ErrCount += 1
}

func isxboDigitStart(x rune) bool {
	if x == 'x' || x == 'b' || x == 'o' {
		return true
	}
	return false
}
