package ast

import "github.com/flowfunction/cddl/token"

type Group struct {
	Pos   token.Position
	Rules []Node
}

func (g *Group) Start() token.Position {
	return g.Pos
}

func (g *Group) End() token.Position {
	if len(g.Rules) == 0 {
		return token.Position{Offset: -1}
	}
	return g.Rules[len(g.Rules)-1].End()
}
