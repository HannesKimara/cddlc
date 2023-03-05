package ast

import "github.com/flowfunction/cddl/token"

// NullType represents the AST Node for the `null` and `nil` tokens
type NullType struct {
	Pos   token.Position
	Token token.Token
}

func (nt *NullType) Start() token.Position {
	return nt.Pos
}

func (nt *NullType) End() token.Position {
	return nt.Pos.To(4) // lenth of `null`
}
