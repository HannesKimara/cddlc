package ast

import "github.com/HannesKimara/cddlc/token"

type Unwrap struct {
	Pos   token.Position
	Token token.Token
	Item  Node
}

func (u *Unwrap) Start() token.Position {
	return u.Pos
}

func (u *Unwrap) End() token.Position {
	return u.Item.End()
}
