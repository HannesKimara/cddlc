package token

import "fmt"

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

type PositionRange struct {
	Start, End Position
}
