package ast

import "github.com/HannesKimara/cddlc/token"

type GroupEntry interface {
	Node
	groupEntry() // convenience function
}

type Group struct {
	Pos     token.Position
	Entries []GroupEntry
}

func (g *Group) Start() token.Position {
	return g.Pos
}

func (g *Group) End() token.Position {
	if len(g.Entries) == 0 {
		return token.Position{Offset: -1}
	}
	return g.Entries[len(g.Entries)-1].End()
}

func (g *Group) groupEntry() {}
