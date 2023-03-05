package ast

import "github.com/flowfunction/cddl/token"

// Bound represents the AST Node for `..` and `...` range definition tokens.
type Bound struct {
	Pos      token.Position
	Token    token.Token
	From, To Node
}

func (b *Bound) Start() token.Position {
	return b.Pos
}

func (b *Bound) End() token.Position {
	return b.Pos.To(2 + int(b.Token-token.INCLUSIVE_BOUND)) // Assume token is either .. or ...
}
