package parser

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

type Diagnostic interface {
	String() string
	Diagnostic() string

	// Pos returns beginning position
	Pos() token.Position
}

// Warning
type Warning struct {
	// Range - the range of positions in the source causing the warning
	Range token.PositionRange

	// Msg - the short message
	Msg string

	// Prefix - Prefix of the string formatted warning e.g. parser, lexer
	Prefix string
}

// String returns the string representation of the Warning in the form
//
// `module` warning: msg
func (w *Warning) String() string {
	return fmt.Sprintf("%s warning: %s", w.Prefix, w.Msg)
}

// Diagnostic returns string formatted warning with position
func (w *Warning) Diagnostic() string {
	return fmt.Sprintf("%s at %s", w, w.Range)
}

// Pos returns the beginning position
func (w *Warning) Pos() token.Position {
	return w.Range.Start
}

type Error struct {
	// Range - the range of positions in the source causing the error
	Range token.PositionRange

	// Msg - the short message
	Msg string

	// Prefix - Prefix of the string formatted warning e.g. parser, lexer
	Prefix string
}

// String returns the string representation of the Warning in the form
//
// `module` error: msg
func (e *Error) String() string {
	return fmt.Sprintf("%s error: %s", e.Prefix, e.Msg)
}

// Diagnostic returns string formatted error with position
func (w *Error) Diagnostic() string {
	return fmt.Sprintf("%s at %s", w, w.Range)
}

// Pos returns the beginning position
func (e *Error) Pos() token.Position {
	return e.Range.Start
}

func (e *Error) Error() string {
	return e.String()
}

// NewWarning returns a Warning with the provided parameters.
func NewWarning(prefix, msg string, start token.Position, end token.Position) *Warning {
	return &Warning{
		Prefix: prefix,
		Range: token.PositionRange{
			Start: start,
			End:   end,
		},
		Msg: msg,
	}
}

// NewWarning returns a Warning with the provided parameters.
func NewError(prefix, msg string, start token.Position, end token.Position) *Error {
	return &Error{
		Prefix: prefix,
		Range: token.PositionRange{
			Start: start,
			End:   end,
		},
		Msg: msg,
	}
}
