package gogen

import (
	gast "go/ast"
	"go/token"
	"strconv"

	"github.com/HannesKimara/cddlc/ast"
)

func (g *Generator) transpileTextLiteral(tl *ast.TextLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.STRING,
		Value: tl.Literal,
	}
}

func (g *Generator) transpileFloatLiteral(fl *ast.FloatLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.FLOAT,
		Value: strconv.FormatFloat(fl.Literal, 'f', -1, 64),
	}
}

func (g *Generator) transpileIntegerLiteral(il *ast.IntegerLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.INT,
		Value: strconv.FormatInt(il.Literal, 10),
	}
}

func (g *Generator) transpileComment(comment *ast.Comment) *gast.Comment {
	return &gast.Comment{
		Text: comment.Text,
	}
}

func (g *Generator) transpileBoolLiteral(bl *ast.BooleanLiteral) *gast.Ident {
	return &gast.Ident{
		Name: strconv.FormatBool(bl.Bool),
	}
}
