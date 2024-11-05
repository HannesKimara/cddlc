package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// Bits represents the AST Node for `.bits` control operator
type Bits struct {
	// Pos: the position of the .size token
	Pos token.Position

	// Token: the token responsible for the node
	Token token.Token

	// Base: the base node
	Base Node

	// Constraint: the bits constraint to apply
	Contstraint Node
}

func (r *Bits) Start() token.Position {
	return r.Base.Start()
}

func (r *Bits) End() token.Position {
	return r.Contstraint.End()
}

func (r *Bits) String() string {
	s := fmt.Sprintf("%s - %s", r.Start(), r.End())

	return s
}
