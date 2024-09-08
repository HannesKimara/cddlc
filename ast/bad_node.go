package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// BadNode represents a node formed due to  parsing error
type BadNode struct {
	// The start position of the bad node
	Pos token.Position

	// The position of the triggering token
	Token token.Token

	// The base node of the bad node
	Base Node

	// The end position
	EndPos token.Position
}

// Start returns the start if the base node
func (b *BadNode) Start() token.Position {
	return b.Base.Start()
}

// End returns the estimated end of the bad node
func (b *BadNode) End() token.Position {
	return b.EndPos
}

func (b *BadNode) String() string {
	s := fmt.Sprintf("%s - %s", b.Start(), b.End())

	return s
}
