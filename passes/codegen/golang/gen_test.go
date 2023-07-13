package gogen_test

import (
	"os"
	"testing"

	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/parser"
	gogen "github.com/flowfunction/cddl/passes/codegen/golang"
)

func TestGen(t *testing.T) {
	gen := gogen.NewGenerator()

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
