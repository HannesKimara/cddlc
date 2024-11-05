package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// FloatType represents the AST Node for `float, float16, float32, float64` type definition tokens.
type FloatType struct {
	Pos   token.Position
	Token token.Token
	Base  int // 16, 32, 64
}

func (ft *FloatType) Start() token.Position {
	return ft.Pos
}

func (ft *FloatType) End() token.Position {
	return ft.Pos.To(5) // TODO: support bases float64, 32, 16
}
func (ft *FloatType) String() string {
	s := fmt.Sprintf("%s - %s", ft.Start(), ft.End())

	return s
}

// FloatLiteral represesnts the AST Node for float type token i.e. 3.412
type FloatLiteral struct {
	Range   token.PositionRange
	Token   token.Token
	Literal float64
}

func (fl *FloatLiteral) Start() token.Position {
	return fl.Range.Start
}

func (fl *FloatLiteral) End() token.Position {
	return fl.Range.End
}
func (fl *FloatLiteral) String() string {
	s := fmt.Sprintf("%s - %s", fl.Start(), fl.End())

	return s
}
