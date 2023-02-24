package ast

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

type CDDLType interface {
	IntegerType | BooleanType | TstrType | FloatType | BstrType | BytesType | NullType
}

type Node interface {
	Start() int
	End() int
	String() string
}

// Comment represents a single ;-style comment
type Comment struct {
	Pos  token.Position
	Text string
}

// CommentGroup represents a sequence of comments with no empty lines
type CommentGroup struct {
	List []*Comment
}

func (cg *CommentGroup) String() (out string) {
	if cg == nil {
		return ""
	}

	for _, c := range cg.List {
		out += c.Text[1:] + "\n"
	}
	return
}

type Identifer struct {
	Pos  token.Position
	Name string
}

func (i *Identifer) String() string {
	return fmt.Sprintf("Identifier( Name: %s) ", i.Name)
}

type Entry interface {
	String() string
}

type Rule struct {
	Pos   token.Position
	Token token.Token
	Name  *Identifer
	Value Entry
}

type (
	IntegerType struct {
		Pos   token.Position
		Token token.Token
		Base  int // 16, 32, 64
	}

	BooleanType struct {
		Pos   token.Position
		Token token.Token
	}

	TstrType struct {
		Pos   token.Position
		Token token.Token
	}

	FloatType struct {
		Pos   token.Position
		Token token.Token
		Base  int // 16, 32, 64
	}

	BstrType struct {
		Pos   token.Position
		Token token.Token
	}

	BytesType struct {
		Pos   token.Position
		Token token.Token
	}

	NullType struct {
		Pos   token.Position
		Token token.Token
	}
)

type (
	IntegerLiteral struct {
		Pos     token.Position
		Token   token.Token
		Literal int
	}
)

func (il *IntegerLiteral) String() string {
	return fmt.Sprintf("IntegerLiteral ( Literal: %d )", il.Literal)
}

// Control Operators
type (
	SizeOperatorControl struct {
		Pos   token.Position
		Token token.Token

		// Size: number of bytes
		Size *IntegerLiteral
		Type Entry
	}
)

func (t *SizeOperatorControl) String() string {
	return "SizeOperatorControl ( Type: " + t.Type.String() + fmt.Sprintf("Size: %d", t.Size.Literal) + ")"
}

type TypeChoice struct {
	Pos           token.Position
	Token         token.Token
	First, Second Entry
}

func (t *TypeChoice) String() string {
	return "TypeChoice ( First: " + t.First.String() + " Second: " + t.Second.String() + ")"
}

func (it *IntegerType) String() string {
	return "IntegerType ( Value:" + it.Token.String() + ")"
}

func (bt *BooleanType) String() string {
	return "BoolType ()"
}

func (tt *TstrType) String() string {
	return "TstrType () "
}

func (ft *FloatType) String() string {
	return "FloatType ()"
}

func (bt *BstrType) String() string {
	return "BstrType () "
}

func (bt *BytesType) String() string {
	return "BytesType ()"
}

func (nt *NullType) String() string {
	return "NullType ()"
}

func (r *Rule) Start() int { return r.Pos.Offset }
func (r *Rule) End() int   { return r.Pos.Offset }
func (r *Rule) String() string {
	out := ""

	if r.Name != nil {
		out += "Rule( Name: " + r.Name.String()
	}
	if r.Value != nil {
		out += "Value: " + r.Value.String() + ") "
	}

	return out
}

type Group struct {
	Pos   token.Position
	Rules []Entry
}

func (g *Group) String() string {
	out := "Group ( Rules: ["

	for _, rule := range g.Rules {
		out += rule.String() + ","
	}
	out += "])"

	return out
}

// File Node represents a cddl source file
type CDDL struct {
	Pos   token.Position
	Rules []*Rule
}

func (c *CDDL) String() string {
	out := "CDDL ( Rules: [ "
	for _, rule := range c.Rules {
		out += rule.String() + ","
	}
	out += " ])"

	return out
}

func (c *CDDL) Start() int { return c.Pos.Offset }
func (c *CDDL) End() int   { return c.Pos.Offset }
