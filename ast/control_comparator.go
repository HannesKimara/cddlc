package ast

import "github.com/flowfunction/cddl/token"

// ControlOpControl represents the AST Node for operators `.lt, .le, .gt, .ge, .eq, .ne` with numeric left right values.
// This also takes identifiers that resolve to numeric types.
//
// Numeric values are: `int, uint, nint, float, float16, float32, float64`
type ComparatorOpControl struct {
	Pos         token.Position
	Token       token.Token
	Operator    string
	Left, Right Node
}

func (cc *ComparatorOpControl) Start() token.Position {
	return cc.Left.End()
}

func (cc *ComparatorOpControl) End() token.Position {
	return cc.Right.End()
}
