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
