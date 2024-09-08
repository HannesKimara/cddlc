package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// Rule represents the AST Node for typed identifer.
// It maps the name of the type to the type
type Rule struct {
	Pos             token.Position
	Name            *Identifier
	Value           Node
	TrailingComment *Comment
}

func (r *Rule) Start() token.Position {
	return r.Name.Start()
}
func (r *Rule) End() token.Position {
	return r.Value.End()
}

func (r *Rule) String() string {
	s := fmt.Sprintf("%s - %T:%s", r.Name, r.Value, r.Value)

	return s
}

func (r *Rule) ge() {}

func (r *Rule) cddlEntry() {}
