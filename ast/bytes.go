package ast

import (
	"fmt"

	"github.com/HannesKimara/cddlc/token"
)

// BstrType represents the AST Node for the `bstr` type definition token.
type BstrType struct {
	Pos   token.Position
	Token token.Token
}

// BytesType represents the AST Node for the `bytes` type definition token.
type BytesType struct {
	Pos   token.Position
	Token token.Token
}

func (b *BytesType) Start() token.Position {
	return b.Pos
}

func (b *BytesType) End() token.Position {
	return b.Pos.To(5) // length of `bytes`
}

func (b *BytesType) String() string {
	s := fmt.Sprintf("%s - %s", b.Start(), b.End())

	return s
}

func (b *BstrType) Start() token.Position {
	return b.Pos
}

func (b *BstrType) End() token.Position {
	return b.Pos.To(5) // length of `bytes`
}

func (b *BstrType) String() string {
	s := fmt.Sprintf("%s - %s", b.Start(), b.End())

	return s
}
