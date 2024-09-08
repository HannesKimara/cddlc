package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

type Enumeration struct {
	Pos   token.Position
	Token token.Token
	Value Node
}

func (e *Enumeration) Start() token.Position {
	return e.Pos
}
func (e *Enumeration) End() token.Position {
	return e.Value.End()
}
func (e *Enumeration) String() string {
	s := fmt.Sprintf("%s - %s", e.Start(), e.End())

	return s
}

func (e *Enumeration) groupEntry() {}
