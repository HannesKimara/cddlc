package ast

import "github.com/flowfunction/cddl/token"

// BooleanType represents the AST Node for the `bool` type definition token.
type BooleanType struct {
	Pos   token.Position
	Token token.Token
}

func (b *BooleanType) Start() token.Position {
	return b.Pos
}

func (b *BooleanType) End() token.Position {
	return b.Pos.To(4)
}
