package token

import (
	"fmt"
)

type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

func (pos Position) String() string {
	s := pos.Filename
	if s != "" {
		s += ":"
	}
	s += fmt.Sprintf("%d", pos.Line)

	if pos.Column > 0 {
		s += fmt.Sprintf(":%d", pos.Column)
	}

	return s
}

// To returns the Position(2) after a positive offset from the current position(1)
//
// name = tstr
//        ^  ^
//        |  |
//        1  2
// where offset is the length of the token(s)
func (pos Position) To(offset int) Position {
	return Position{
		Filename: pos.Filename,
		Offset:   pos.Offset + offset,
		Line:     pos.Line,
		Column:   pos.Column + offset,
	}
}

type PositionRange struct {
	Start, End Position
}
