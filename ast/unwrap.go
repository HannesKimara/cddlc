package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

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

func (u *Unwrap) String() string {
	s := fmt.Sprintf("%s - %s", u.Start(), u.End())

	return s
}
