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

func (g *Generator) transpileIdentifier(ident *ast.Identifier) *gast.Ident {
	formatted := strcase.ToCamel(ident.Name)

	if ident.IsSocket() || ident.IsPlug() {
		formatted = strings.TrimLeft(ident.Name, "$")
	}

	if token.IsIdentifier(formatted) {
		return &gast.Ident{
			Name: formatted,
		}
	}
	_, err := strconv.ParseInt(formatted, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("Identifier %s -> %s could not be transformed to valid go identifier", ident.Name, formatted))
	}
	return &gast.Ident{
		Name: "IntKey_" + formatted,
	}
}

func init() {
	converters := make(map[string]stringConverter)

	converters["camelcase"] = strcase.ToCamel
	converters["snakecase"] = strcase.ToSnake
}
