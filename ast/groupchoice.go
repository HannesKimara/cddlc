package ast

import "github.com/flowfunction/cddl/token"

type GroupChoice struct {
	Pos           token.Position
	Token         token.Token
	First, Second Node
}

func (gc *GroupChoice) Start() token.Position {
	return gc.First.Start()
}

func (gc *GroupChoice) End() token.Position {
	return gc.Second.End()
}
