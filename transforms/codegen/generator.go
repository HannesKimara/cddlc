package codegen

import (
	"io"

	"github.com/flowfunction/cddl/ast"
)

// Generator decribes methods for a code generator
type Generator interface {
	// String writes the string representation out to the provided io.Writer
	String(io.Writer) error

	// Visit takes a CDDL Node and recursively walks it
	Visit(ast.Node) Generator
}
