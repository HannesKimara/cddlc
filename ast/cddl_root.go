package ast

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

// File Node represents a cddl source file
type CDDL struct {
	Pos   token.Position
	Rules []*Rule
}

func (c *CDDL) String() string {
	out := "CDDL ("
	for _, rule := range c.Rules {
		out += fmt.Sprintf("%+v,", rule)
	}
	out += ")"
	return out
}

func (c *CDDL) Start() token.Position {
	return c.Pos
}

func (c *CDDL) End() token.Position {
	if len(c.Rules) == 0 {
		return token.Position{Offset: -1}
	}
	return c.Rules[len(c.Rules)-1].End()
}
