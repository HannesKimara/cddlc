package errors

import "github.com/HannesKimara/cddlc/token"

type Diagnostic interface {
	error
	String() string
	Diagnostic() string

	// Pos returns beginning position
	Start() token.Position
	End() token.Position
}
