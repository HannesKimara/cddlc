package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// Entry represents the Node for a group entry
// It maps the name of the type to the type
type Entry struct {
	Pos             token.Position
	Name            *Identifier
	Value           Node
	TrailingComment *Comment
}

func (r *Entry) Start() token.Position {
	return r.Name.Start()
}
func (r *Entry) End() token.Position {
	return r.Value.End()
}
func (r *Entry) String() string {
	s := fmt.Sprintf("%s - %s", r.Start(), r.End())

	return s
}
func (r *Entry) groupEntry() {}
