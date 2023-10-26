package ast

import "github.com/HannesKimara/cddlc/token"

// SizeOperatorControl represents the AST Node for `.size` control operator
type SizeOperatorControl struct {
	// Pos: the position of the .size token
	Pos token.Position

	// Token: the token responsible for the node
	Token token.Token

	// Size: number of bytes
	Size Node

	// Type: the base type to apply the contraint to
	Type Node
}

// Start returns the start of the base type
func (sc *SizeOperatorControl) Start() token.Position {
	return sc.Type.Start()
}

// End returns the end of the size
func (sc *SizeOperatorControl) End() token.Position {
	return sc.Size.End()
}
