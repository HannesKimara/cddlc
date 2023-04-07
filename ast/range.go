package ast

import "github.com/flowfunction/cddl/token"

// Range represents the AST Node for `..` and `...` range definition tokens.
type Range struct {
	Pos      token.Position
	Token    token.Token
	From, To Node
}

// Start returns the start of the node
func (r *Range) Start() token.Position {
	return r.From.Start()
}

// End returns the last position of the node
func (r *Range) End() token.Position {
	// Assuming token is either .. or ... the statement yields (2 + 0) or (2 +1)
	return r.To.End()
}
