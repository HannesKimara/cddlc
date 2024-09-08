package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

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

func (tc *TypeChoice) String() string {
	s := fmt.Sprintf("%s - %s", tc.Start(), tc.End())

	return s
}
