package ast

import "github.com/HannesKimara/cddlc/token"

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

type BooleanLiteral struct {
	Range token.PositionRange
	Bool  bool
}

func (bl *BooleanLiteral) Start() token.Position {
	return bl.Range.Start
}

func (bl *BooleanLiteral) End() token.Position {
	return bl.Range.End
}
