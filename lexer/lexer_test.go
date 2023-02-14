package lexer_test

import (
	"runtime"
	"testing"

	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/token"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// packagepath = filepath.Dir(b)
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
		{token.INT, "10", "10", ""},

		{token.FLOAT, "3.4", "3.4", ""},
	}

	for _, tst := range nums {
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
