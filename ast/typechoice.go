package ast

import "github.com/HannesKimara/cddlc/token"

// TypeChoice represents the AST Node for `/` type choice operator
type TypeChoice struct {
	Pos           token.Position
	Token         token.Token
	First, Second Node
}

func (tc *TypeChoice) Start() token.Position {
	return tc.First.Start()
}

func (tc *TypeChoice) End() token.Position {
	return tc.Second.End()
}
