package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

type Array struct {
	Pos   token.Position
	Rules []GroupEntry
}

func (a *Array) Start() token.Position {
	return a.Pos
}

func (a *Array) End() token.Position {
	if len(a.Rules) == 0 {
		return token.Position{Offset: -1}
	}
	return a.Rules[len(a.Rules)-1].End()
}

func (a *Array) String() string {
	s := fmt.Sprintf("%s - %s", a.Start(), a.End())

	return s
}
