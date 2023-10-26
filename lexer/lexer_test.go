package lexer_test

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/token"
	"github.com/google/uuid"
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

var seed int64

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

var fuzzRuns uint = 1000
var maxGenTokens uint = 10
var maxGenStringLen uint = 50

func init() {
	seed = time.Now().UnixNano()
	rand.Seed(seed)
}

type Source struct {
	source []string
	tokens []token.Token
	pos    []token.Position

	ix uint
}

func (s *Source) Source() string {
	return strings.Join(s.source, " ")
}

func (s *Source) String() string {
	src := ""
	if len(s.source)+len(s.tokens) == 2*len(s.pos) { // check all slices are of equal length
		for i := 0; i < len(s.source); i++ {
			src += fmt.Sprintf("%s: %s -> %s\n", s.pos[i], s.tokens[i], s.source[i])
		}
	}
	return src
}

func (s *Source) Scan() (token.Token, token.Position, string) {
	tok := s.tokens[s.ix]
	lit := strings.TrimSuffix(s.source[s.ix], " ")
	pos := s.pos[s.ix]

	if int(s.ix) < len(s.tokens)-1 {
		s.ix += 1
	}
	return tok, pos, lit
}

func posFromOffset(offset int) token.Position {
	return token.Position{
		Filename: "",
		Offset:   offset,
		Line:     1,
		Column:   offset + 1,
	}
}

func genText() string {
	src := []rune{}
	for i := 0; i < int(maxGenStringLen); i++ {
		char := rune(rand.Int31n(90) + 35)
		src = append(src, char)
	}

	return string(src)
}

func genLiteral(t token.Token) (string, token.Token) {
	switch t {
	case token.INT:
		return fmt.Sprintf("%d", rand.Intn(math.MaxInt32)), t
	// case token.UINT:
	// 	return fmt.Sprintf("%d", rand.Uint64()), t
	case token.FLOAT, token.FLOAT16, token.FLOAT32, token.FLOAT64: // TODO: differentiate these
		return fmt.Sprintf("%f", rand.Float64()*math.MaxInt64), token.FLOAT
	case token.TEXT_LITERAL:
		return `"` + genText() + `"`, t
	}
	return t.String(), t
}

func genSource() *Source {
	src := []string{}
	toks := []token.Token{}
	pos := []token.Position{}
	offset := 0
	rangeTokens := int(token.FEATURE - token.IDENT)

	for i := 0; i < int(maxGenTokens); i++ {
		tok := rand.Intn(rangeTokens) + int(token.IDENT)
		var tokn token.Token = token.Token(tok)
		if strings.HasPrefix(tokn.String(), "INTERNAL_MARKER") {
			tokn = token.CBORSEQ // arbitrary replacement of internal markers
		}

		literal, token := genLiteral(tokn)
		src = append(src, literal)
		toks = append(toks, token)
		pos = append(pos, posFromOffset(offset))
		offset += len(literal + " ")
	}
	return &Source{
		ix:     0,
		source: src,
		tokens: toks,
		pos:    pos,
	}
}

type failStore struct {
	fd *os.File
	bw *bufio.Writer
}

func (fs *failStore) Store(src string) error {
	_, err := fs.bw.WriteString(src)
	if err != nil {
		return err
	}

	return nil
}

func (fs *failStore) Close() error {
	err := fs.bw.Flush()
	if err != nil {
		return err
	}
	return fs.fd.Close()
}

func NewFailStore(fp string) (*failStore, error) {
	dir := filepath.Dir(fp)
	if _, err := os.Stat(dir); err != nil {
		switch {
		case os.IsNotExist(err):
			err := os.Mkdir(dir, os.ModeDir)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	fd, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	bw := bufio.NewWriter(fd)
	bw.WriteString(fmt.Sprintf("Seed: %d\n\n", seed))
	bw.Flush()
	return &failStore{fd: fd, bw: bw}, nil
}

func assert[T comparable](t *testing.T, expected, got T) {
	if expected != got {
		t.Logf("\nFAIL: expected %+v, got %+v (%T)\n", expected, got, expected)
		t.Fail()
		return
	}
}

func TestLexerFuzz(t *testing.T) {
	fs, err := NewFailStore(filepath.Join(rootDir(), "artifacts", uuid.New().String()))
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	for i := 0; i < int(fuzzRuns); i += 1 {
		source := genSource()
		lex := lexer.NewLexer([]byte(source.Source()))

		t.Run(fmt.Sprintf("Source: %d", i+1), func(t *testing.T) {
			for {
				tok, pos, lit := lex.Scan()
				// HACK
				if tok == token.TEXT_LITERAL {
					lit = `"` + lit + `"`
				}

				if tok == token.EOF {
					break
				}
				exTok, exPos, exLit := source.Scan()

				assert[token.Token](t, exTok, tok)
				assert[token.Position](t, exPos, pos)
				assert[string](t, exLit, lit)
			}
			if t.Failed() {
				fs.Store("\nFuzz Test: " + t.Name() + "\n")
				fs.Store(strings.Repeat("-", 90))
				fs.Store(fmt.Sprintf("\nSource: %s\n", source.Source()))
				if len(lex.Errors) > 0 {
					fs.Store("Lexer Errors: ")
					for _, err := range lex.Errors {
						fs.Store(fmt.Sprintf("\t%s -> %s\n", err.Pos, err.Msg))
					}
				}
			}
		})
	}
}
