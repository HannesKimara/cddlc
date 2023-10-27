// Package ast implements nodes for the CDDL Abstract Syntax Tree

package ast

import "github.com/HannesKimara/cddlc/token"

// Node interface represents an AST Node
type Node interface {
	// Start returns the start token of the node
	Start() token.Position

	// End returns the end token of the node
	End() token.Position
}
