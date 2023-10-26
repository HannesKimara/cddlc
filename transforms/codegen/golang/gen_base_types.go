package gogen

import (
	gast "go/ast"

	"github.com/HannesKimara/cddlc/ast"
)

func (g *Generator) transpileTstrType(tstr *ast.TstrType) *gast.Ident {
	return &gast.Ident{
		Name: "string",
	}
}

func (g *Generator) transpileBytesType(bt *ast.BytesType) *gast.ArrayType {
	return &gast.ArrayType{
		Elt: &gast.Ident{
			Name: "byte",
		},
	}
}

func (g *Generator) transpileNullType(nt *ast.NullType) *gast.Ident {
	return &gast.Ident{
		Name: "nil",
	}
}

func (g *Generator) transpileIntegerType(it *ast.IntegerType) *gast.Ident {
	return &gast.Ident{
		Name: "int",
	}
}

func (g *Generator) transpileUintType(it *ast.UintType) *gast.Ident {
	return &gast.Ident{
		Name: "uint",
	}
}

func (g *Generator) transpileBoolType(bt *ast.BooleanType) *gast.Ident {
	return &gast.Ident{
		Name: "bool",
	}
}
