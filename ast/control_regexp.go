package ast

import "github.com/flowfunction/cddl/token"

type Regexp struct {
	Pos   token.Position
	Token token.Token
	Base  *TstrType
	Regex Node
}

func (r *Regexp) Start() token.Position {
	return r.Pos
}

func (r *Regexp) End() token.Position {
	return r.Regex.End()
}
