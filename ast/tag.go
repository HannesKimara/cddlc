package ast

import "github.com/flowfunction/cddl/token"

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
