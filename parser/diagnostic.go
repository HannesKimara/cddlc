package parser

import (
	"fmt"

	"github.com/HannesKimara/cddlc/errors"
	"github.com/HannesKimara/cddlc/token"
)

type Error struct {
	// Range - the range of positions in the source causing the error
	Range token.PositionRange

	// Msg - the short message
	Msg string

	// Prefix - Prefix of the string source e.g. parser, lexer
	Prefix string
}

// String returns the string representation of the Error in the form
//
// `module` error: msg
func (e *Error) String() string {
	return fmt.Sprintf("%s error: %s", e.Prefix, e.Msg)
}

// Diagnostic returns string formatted error with position
func (e *Error) Diagnostic() string {
	return fmt.Sprintf("%s at %s", e, e.Range.String())
}

// Start returns the beginning position
func (e *Error) Start() token.Position {
	return e.Range.Start
}

// End returns the beginning position
func (e *Error) End() token.Position {
	return e.Range.End
}

// Error satisfies the error interface. Returns the same value as String
func (e *Error) Error() string {
	return e.String()
}

// NewError returns an Error with the provided parameters.
func NewError(msg string, start token.Position, end token.Position) *Error {
	return &Error{
		Prefix: "parser",
		Range: token.PositionRange{
			Start: start,
			End:   end,
		},
		Msg: msg,
	}
}

// ErrorList encapsulates a collection of related errors.
type ErrorList []errors.Diagnostic

// String returns the concatenated individual errors.
func (er ErrorList) String() string {
	out := ""
	for _, err := range er {
		out += err.Error() + "\n"
	}

	return out
}

func (er ErrorList) Error() string {
	return er.String()
}

func (er ErrorList) Start() token.Position {
	if len(er) > 0 {
		return er[0].Start()
	}
	return token.Position{}
}

func (er ErrorList) End() token.Position {
	l := len(er)
	if l > 0 {
		return er[l-1].End()
	}
	return token.Position{}
}

func (er ErrorList) Diagnostic() string {
	return er.String()
}

func (er ErrorList) Collect() ErrorList {
	if len(er) == 0 {
		return nil
	}
	return er
}
