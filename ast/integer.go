package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
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

type NegativeIntegerType struct {
	Pos   token.Position
	Token token.Token
}

func (nt *NegativeIntegerType) Start() token.Position {
	return nt.Pos
}

func (nt *NegativeIntegerType) End() token.Position {
	return nt.Pos.To(4) // length of `nint`
}

// IntegerLiteral represents the AST Node for an integer literal i.e 3
type IntegerLiteral struct {
	Pos     token.Position
	Token   token.Token
	Literal int64
}

func (il *IntegerLiteral) Start() token.Position {
	return il.Pos
}

func (il *IntegerLiteral) End() token.Position {
	return il.Pos.To(len(fmt.Sprintf("%d", il.Literal)))
}
