package ast

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

// IntegerType represents the AST Node for the `int` type decalration token
type IntegerType struct {
	Pos   token.Position
	Token token.Token
}

func (it *IntegerType) Start() token.Position {
	return it.Pos
}

func (it *IntegerType) End() token.Position {
	return it.Pos.To(3) // length of `int`
}

// IntegerLiteral represents the AST Node for an integer literal i.e 3
type IntegerLiteral struct {
	Pos     token.Position
	Token   token.Token
	Literal int
}

func (il *IntegerLiteral) Start() token.Position {
	return il.Pos
}

func (il *IntegerLiteral) End() token.Position {
	return il.Pos.To(len(fmt.Sprintf("%d", il.Literal)))
}
