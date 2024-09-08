package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

type GroupEntry interface {
	Node
	groupEntry() // convenience function
}

// TODO: set during parsing instead
type GroupType uint

const (
	GroupTypeStruct = iota + 1
	GroupTypeEnum
)

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
func (g *Group) String() string {
	s := fmt.Sprintf("%s - %s", g.Start(), g.End())

	return s
}
func (g *Group) groupEntry() {}
