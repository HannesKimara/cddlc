package ast

import "github.com/HannesKimara/cddlc/token"

type Map struct {
	Pos   token.Position
	Token token.Token
	Rules []Node
}

func (m *Map) Start() token.Position {
	return m.Pos
}

func (m *Map) End() token.Position {
	if len(m.Rules) == 0 {
		return token.Position{Offset: -1}
	}
	return m.Rules[len(m.Rules)-1].End()
}
