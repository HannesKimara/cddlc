package ast

import "github.com/HannesKimara/cddlc/token"

// Optional represents the AST Node for the `?` prefixed optional entry
type Optional struct {
	Pos   token.Position
	Token token.Token
	Item  Node
}

func (i *Optional) Start() token.Position {
	return i.Pos
}

func (i *Optional) End() token.Position {
	return i.Item.End()
}

type NMOccurrence struct {
	Pos   token.Position
	Token token.Token
	N, M  *UintLiteral
	Item  Node
}

func (nm *NMOccurrence) Start() token.Position {
	return nm.N.Start()
}

func (nm *NMOccurrence) End() token.Position {
	return nm.M.End()
}
