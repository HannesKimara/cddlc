package gogen

import (
	gast "go/ast"

	"github.com/HannesKimara/cddlc/ast"
)

type Architecture uint

const (
	ArchitectureGeneric Architecture = iota + 1
	Architecture32
	Architecture64
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
	name := "int"
	switch g.architecture {
	case Architecture32:
		name = "int32"
	case Architecture64:
		name = "int64"
	}
	return &gast.Ident{
		Name: name,
	}
}

func (g *Generator) transpileUintType(it *ast.UintType) *gast.Ident {
	name := "uint"
	switch g.architecture {
	case Architecture32:
		name = "uint32"
	case Architecture64:
		name = "uint64"
	}
	return &gast.Ident{
		Name: name,
	}
}

func (g *Generator) transpileBoolType(bt *ast.BooleanType) *gast.Ident {
	return &gast.Ident{
		Name: "bool",
	}
}

func (g *Generator) transpileFloatType(fl *ast.FloatType) *gast.Ident {
	name := "float64"
	switch g.architecture {
	case Architecture32:
		name = "float32"
	case Architecture64:
		name = "float64"
	}
	return &gast.Ident{
		Name: name,
	}
}
