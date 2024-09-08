package gogen

import (
	"fmt"
	gast "go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/iancoleman/strcase"
)

type stringConverter func(string) string

var converters map[string]stringConverter

func (g *Generator) formatIdentifier(name string, trailingComment *ast.Comment) *gast.Ident {
	formatted := strcase.ToCamel(name)

	if token.IsIdentifier(formatted) {
		formatted = g.identifierPrefix + formatted
	} else {
		_, err := strconv.ParseInt(formatted, 0, 64)
		if err != nil {
			panic(fmt.Sprintf("Identifier %s -> %s could not be transformed to valid go identifier", name, formatted))
		}

		formatted = "IntKey_" + formatted
	}

	if g.customIdentifierFormatter != nil {
		comment := ""
		if trailingComment != nil {
			comment = strings.TrimSpace(trailingComment.Text)
		}
		formatted = g.customIdentifierFormatter(name, comment, formatted)
	}

	return &gast.Ident{
		Name: formatted,
	}
}

func (g *Generator) transpileIdentifier(ident *ast.Identifier, trailingComment *ast.Comment) *gast.Ident {
	name := ident.Name
	if ident.IsSocket() || ident.IsPlug() {
		name = strings.TrimLeft(ident.Name, "$")
	}

	return g.formatIdentifier(name, trailingComment)
}

func init() {
	converters := make(map[string]stringConverter)

	converters["camelcase"] = strcase.ToCamel
	converters["snakecase"] = strcase.ToSnake
}
