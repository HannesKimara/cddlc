package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

type Tag struct {
	Pos              token.Position
	Token            token.Token
	Major, TagNumber *UintLiteral
	Item             Node
}

func (t *Tag) Start() token.Position {
	return t.Pos
}

func (t *Tag) End() token.Position {
	return t.Item.End().To(1) // add )
}

func (t *Tag) String() string {
	s := fmt.Sprintf("%s - %s", t.Start(), t.End())

	return s
}
