package ast

import (
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
