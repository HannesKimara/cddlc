package ast

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

// Identifer represents the AST Node for named identifiers
type Identifier struct {
	Pos  token.Position
	Name string
}

func (i *Identifier) String() string {
	return fmt.Sprintf("Identifier( Name: %s ) ", i.Name)
}

func (i *Identifier) Start() token.Position {
	return i.Pos
}

func (i *Identifier) End() token.Position {
	return i.Pos.To(len(i.Name))
}

func (i *Identifier) IsSocket() bool {
	if len(i.Name) == 0 {
		return false
	}
	return i.Name[0] == '$' && !i.IsPlug()
}

func (i *Identifier) IsPlug() bool {
	if len(i.Name) < 2 {
		return false
	}
	return i.Name[0:2] == "$$"
}
func (i *Identifier) groupEntry() {}
