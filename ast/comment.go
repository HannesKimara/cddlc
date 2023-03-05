package ast

import (
	"fmt"

	"github.com/flowfunction/cddl/token"
)

// Comment represents the AST Node for ;-style single line comment.
type Comment struct {
	Pos  token.Position
	Text string
}

func (c *Comment) String() string {
	return fmt.Sprintf("Comment ( Text:%s )", c.Text)
}

func (c *Comment) Start() token.Position {
	return c.Pos
}

func (c *Comment) End() token.Position {
	return c.Pos.To(len(c.Text))
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

func (cg *CommentGroup) Start() token.Position {
	if len(cg.List) == 0 {
		return token.Position{Offset: -1}
	}
	return cg.List[0].Start()
}

func (cg *CommentGroup) End() token.Position {
	if len(cg.List) == 0 {
		return token.Position{Offset: -1}
	}
	return cg.List[len(cg.List)-1].End()
}
