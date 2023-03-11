package lexer_test

import (
	"testing"

	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/token"
)

func ScanAll(lex *lexer.Lexer) []lexer.TokenSer {
	var tokens []lexer.TokenSer

	for {
		tok, pos, lit := lex.Scan()
		if tok == token.EOF {
			break
		}
		tokens = append(tokens, lexer.TokenSer{Token: tok, Pos: pos, Literal: lit})
	}
	return tokens
}

func TestIdentifers(t *testing.T) {
	tests := []struct {
		tok token.Token
		src string
		lit string
		err string
	}{
		{token.IDENT, "a", "a", ""},
		{token.IDENT, "ab", "ab", ""},
		{token.IDENT, "ab_cd", "ab_cd", ""},

		// test non-letter starting characters
		{token.IDENT, "@a", "@a", ""},
		{token.IDENT, "$a", "$a", ""},
		{token.IDENT, "_a", "_a", ""},

		// test nonconventional body characters ., -, @
		{token.IDENT, "ab.ba", "ab.ba", ""},
		{token.IDENT, "ab-ba", "ab-ba", ""},
		{token.IDENT, "ab@ba", "ab@ba", ""},
		{token.IDENT, "ab-ba@gh.kl8", "ab-ba@gh.kl8", ""},

		// test case sensitivity
		{token.IDENT, "AB", "AB", ""},
		{token.IDENT, "AaBb", "AaBb", ""},

		// test dot continuation from https://www.rfc-editor.org/rfc/rfc8610#section-2.2.2.1
		{token.IDENT, "min..max", "min..max", ""},
		{token.IDENT, "min...max", "min...max", ""},
	}

	for _, tst := range tests {
		l := lexer.NewLexer([]byte(tst.src))

		for {
			tok, _, lit := l.Scan()
			if tok == token.EOF {
				break
			}

			if tok != tst.tok {
				t.Errorf("token mismatch, expected `%s` got `%s`", tst.tok, tok)
			}
			if lit != tst.lit {
				t.Errorf("literal mismatch, expected `%s` got `%s`", tst.lit, lit)
			}
		}
	}
}

func TestNumbers(t *testing.T) {
	nums := []struct {
		tok token.Token
		src string
		lit string
		err string
	}{
		// Base 10
		{token.INT, "0", "0", ""},
		{token.INT, "00", "00", ""},
		{token.INT, "10", "10", ""},

		{token.FLOAT, "00.0", "00.0", ""},
		{token.FLOAT, "0.0", "0.0", ""},
		{token.FLOAT, "3.4", "3.4", ""},
		{token.FLOAT, "3.5694298374823", "3.5694298374823", ""},

		// Base 2
		{token.INT, "0b00", "0b00", ""},
		{token.INT, "0b1", "0b1", ""},
		{token.INT, "0b111010010", "0b111010010", ""},

		// Base 8
		{token.INT, "0o0", "0o0", ""},
		{token.INT, "0o1", "0o1", ""},
		{token.INT, "0o4672726", "0o4672726", ""},

		// Baase 16
		{token.INT, "0x0", "0x0", ""},
		{token.INT, "0x01", "0x01", ""},
		{token.INT, "0x0f755b863f", "0x0f755b863f", ""},
	}

	for _, tst := range nums {
		l := lexer.NewLexer([]byte(tst.src))

		for {
			tok, _, lit := l.Scan()
			if tok == token.EOF {
				break
			}

			if tok != tst.tok {
				t.Errorf("token mismatch, expected `%s` got `%s` in `%s`", tst.tok, tok, tst.lit)
			}
			if lit != tst.lit {
				t.Errorf("literal mismatch, expected `%s` got `%s`", tst.lit, lit)
			}
		}
	}

}
