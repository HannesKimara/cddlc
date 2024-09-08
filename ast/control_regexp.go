package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// Regexp represents the AST Node for `.regexp` control operator
type Regexp struct {
	Pos   token.Position
	Token token.Token
	Base  *TstrType
	Regex Node
}

func (r *Regexp) Start() token.Position {
	return r.Base.Start()
}

func (r *Regexp) End() token.Position {
	return r.Regex.End()
}

func (r *Regexp) String() string {
	s := fmt.Sprintf("%s - %s", r.Start(), r.End())

	return s
}
