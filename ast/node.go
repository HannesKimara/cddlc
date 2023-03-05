package ast

import "github.com/flowfunction/cddl/token"

// Node interface represents an AST Node
type Node interface {
	// Start returns the start token of the node
	Start() token.Position

	// End returns the end token of the node
	End() token.Position
}
