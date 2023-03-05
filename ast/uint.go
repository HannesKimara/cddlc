package ast

import "github.com/flowfunction/cddl/token"

// UintType represents the AST Node for `uint` data definition type
type UintType struct {
	Pos   token.Position
	Token token.Token
}

func (ut *UintType) Start() token.Position {
	return ut.Pos
}

func (ut *UintType) End() token.Position {
	return ut.Pos.To(4)
}
