package errors

import "github.com/flowfunction/cddl/token"

type Diagnostic interface {
	error
	String() string
	Diagnostic() string

	// Pos returns beginning position
	Start() token.Position
	End() token.Position
}
