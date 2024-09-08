package gogen_test

import (
	"testing"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	gogen "github.com/HannesKimara/cddlc/transforms/codegen/golang"
)

type testCommentHandler struct {
	comment          string
	ruleComment      string
	ruleCddlTypeName string
	ruleGoTypeName   string
}

func (h *testCommentHandler) HandleComment(comment string) {
	h.comment = comment
}

func (h *testCommentHandler) HandleRuleComment(comment string, cddlTypeName string, goTypeName string) {
	h.ruleComment = comment
	h.ruleCddlTypeName = cddlTypeName
	h.ruleGoTypeName = goTypeName
}

func TestCommentHandler(t *testing.T) {
	gen := gogen.NewGenerator("lib")
	h := &testCommentHandler{}
	gen.SetCommentHandler(h)

	input := `; Some header comment
	; to my my cddl definition.
	
	; type key 12
	status-message = {
		status: text
	}`

	l := lexer.NewLexer([]byte(input))
	p := parser.NewParser(l)

	cddl, err := p.ParseFile()
	if err != nil {
		t.Fatal(err)
	}

	gen.Visit(cddl)

	expectedComment := "Some header comment\nto my my cddl definition."
	if h.comment != expectedComment {
		t.Errorf("Expected comment %#v, got %#v", expectedComment, h.comment)
	}
	expectedRuleComment := "type key 12"
	if h.ruleComment != expectedRuleComment {
		t.Errorf("Expected rule comment %#v, got %#v", expectedRuleComment, h.ruleComment)
	}
	expectedRuleCddlTypeName := "status-message"
	if h.ruleCddlTypeName != expectedRuleCddlTypeName {
		t.Errorf("Expected rule CDDL type name %#v, got %#v", expectedRuleCddlTypeName, h.ruleCddlTypeName)
	}
	expectedRuleGoTypeName := "StatusMessage"
	if h.ruleGoTypeName != expectedRuleGoTypeName {
		t.Errorf("Expected rule Go type name %#v, got %#v", expectedRuleGoTypeName, h.ruleGoTypeName)
	}
}
