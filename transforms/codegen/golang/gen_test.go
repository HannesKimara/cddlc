package gogen_test

import (
	"os"
	"testing"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	gogen "github.com/HannesKimara/cddlc/transforms/codegen/golang"
)

func TestGen(t *testing.T) {
	gen := gogen.NewGenerator("lib")

	src := `person = [name: tstr, age: uint]`
	l := lexer.NewLexer([]byte(src))
	p := parser.NewParser(l)

	cddl, err := p.ParseFile()
	if err != nil {
		t.Fatal(err)
	}

	gen.Visit(cddl)
	gen.String(os.Stdout)
}
