package parser_test

import (
	"testing"

	"github.com/flowfunction/cddl/ast"
	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/parser"
)

func TestParseIdentifer(t *testing.T) {
	tests := []struct {
		src string
		ast ast.Node
		err string
	}{
		{
			"name = tstr",
			&ast.CDDL{
				Rules: []*ast.Rule{
					{Name: &ast.Identifier{Name: "name"}},
				},
			},
			"",
		},
	}

	for _, tst := range tests {
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		ast := p.Parse()
		t.Log(ast.String())
	}

}
