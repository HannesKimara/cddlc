package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

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

func (gc *GroupChoice) String() string {
	s := fmt.Sprintf("%s - %s", gc.Start(), gc.End())

	return s
}
