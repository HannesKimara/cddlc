package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

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
func (ut *UintType) String() string {
	s := fmt.Sprintf("%s - %s", ut.Start(), ut.End())

	return s
}
