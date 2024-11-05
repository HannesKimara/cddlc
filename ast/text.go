package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// TextLiteral represents the AST Node for a text literal
type TextLiteral struct {
	Pos     token.Position
	Token   token.Token
	Literal string
}

func (tl *TextLiteral) Start() token.Position {
	return tl.Pos
}

func (tl *TextLiteral) End() token.Position {
	return tl.Pos.To(len(tl.Literal))
}

func (tl *TextLiteral) String() string {
	s := fmt.Sprintf("%s - %s", tl.Start(), tl.End())

	return s
}

// TstrType represents the AST Node for the `tstr` type definition token
type TstrType struct {
	Pos   token.Position
	Token token.Token
}

func (tt *TstrType) Start() token.Position {
	return tt.Pos
}

func (tt *TstrType) End() token.Position {
	return tt.Pos.To(4)
}

func (tt *TstrType) String() string {
	s := fmt.Sprintf("%s - %s", tt.Start(), tt.End())

	return s
}
