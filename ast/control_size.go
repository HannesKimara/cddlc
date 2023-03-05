package ast

import "github.com/flowfunction/cddl/token"

// SizeOperatorControl represents the AST Node for `.size` control operator
type SizeOperatorControl struct {
	Pos   token.Position
	Token token.Token

	// Size: number of bytes
	Size *IntegerLiteral
	Type Node
}

func (sc *SizeOperatorControl) Start() token.Position {
	return sc.Type.Start()
}

func (sc *SizeOperatorControl) End() token.Position {
	return sc.Size.End()
}
