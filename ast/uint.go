package ast

import "github.com/HannesKimara/cddlc/token"

// UintType represents the AST Node for `uint` data definition type
type UintType struct {
	Range token.PositionRange
	Token token.Token
}

func (ut *UintType) Start() token.Position {
	return ut.Range.Start
}

func (ut *UintType) End() token.Position {
	return ut.Range.End
}
