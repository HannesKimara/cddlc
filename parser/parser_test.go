package parser_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/flowfunction/cddl/ast"
	"github.com/flowfunction/cddl/errors"
	"github.com/flowfunction/cddl/lexer"
	"github.com/flowfunction/cddl/parser"
	"github.com/flowfunction/cddl/token"
)

func testWalk(t *testing.T, valid, parsed ast.Node) {
	if t.Failed() {
		return
	}
	switch val := valid.(type) {
	case *ast.Identifier:
		if p, ok := parsed.(*ast.Identifier); ok {
			if val.Name != p.Name {
				t.Errorf("expected %s got %s", val.Name, p.Name)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.CDDL:
		if p, ok := parsed.(*ast.CDDL); ok && len(val.Rules) == len(p.Rules) {
			for i := 0; i < len(val.Rules); i++ {
				testWalk(t, val.Rules[i], p.Rules[i])
			}
		}
	case *ast.Rule:
		if p, ok := parsed.(*ast.Rule); ok {
			testWalk(t, val.Name, p.Name)
			testWalk(t, val.Value, p.Value)
			if p.TrailingComment != nil {
				testWalk(t, val.TrailingComment, p.TrailingComment)
			}
			t.Logf("Valid: %T, Parsed: %T\n", valid, parsed)
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.BooleanType, *ast.BytesType, *ast.FloatType, *ast.UintType, *ast.NullType,
		*ast.IntegerType, *ast.NegativeIntegerType, *ast.TstrType, *ast.BstrType:

		if !reflect.DeepEqual(valid, parsed) { // compare everything including position
			t.Fatalf("expected nodes %T:%+v(valid) and %T:%+v(parsed) to be equal", valid, valid, parsed, parsed)
			return
		}

	case *ast.TextLiteral:
		if p, ok := parsed.(*ast.TextLiteral); ok {
			if val.Literal != p.Literal {
				t.Errorf("Text literals do not match. Expected %s got %s", val.Literal, p.Literal)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.IntegerLiteral:
		if p, ok := parsed.(*ast.IntegerLiteral); ok {
			if val.Literal != p.Literal {
				t.Errorf("Integer literals do not match. Expected %d got %d", val.Literal, p.Literal)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.UintLiteral:
		if p, ok := parsed.(*ast.UintLiteral); ok {
			if val.Literal != p.Literal {
				t.Errorf("Uint literals do not match. Expected %d got %d", val.Literal, p.Literal)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.FloatLiteral:
		if p, ok := parsed.(*ast.FloatLiteral); ok {
			if val.Literal != p.Literal {
				t.Errorf("Float literals do not match. Expected %f got %f", val.Literal, p.Literal)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.Regexp:
		if p, ok := parsed.(*ast.Regexp); ok {
			testWalk(t, val.Base, p.Base)
			testWalk(t, val.Regex, p.Regex)
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.Comment:
		if p, ok := parsed.(*ast.Comment); ok {
			if val.Text != p.Text {
				t.Errorf("Comment texts do not match. Expected %s got %s", val.Text, p.Text)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	case *ast.CommentGroup:
		if p, ok := parsed.(*ast.CommentGroup); ok && len(val.List) == len(p.List) {
			for i := 0; i < len(val.List); i++ {
				testWalk(t, val.List[i], p.List[i])
			}
		}
	case *ast.Tag:
		if p, ok := parsed.(*ast.Tag); ok {
			if val.Major != nil && p.Major != nil {
				testWalk(t, val.Major, p.Major)
			}
			if val.TagNumber != nil && p.TagNumber != nil {
				testWalk(t, val.TagNumber, p.TagNumber)
			}
			if val.Item != nil && p.Item != nil {
				testWalk(t, val.Item, p.Item)
			}
		}
	case *ast.Array:
		if p, ok := parsed.(*ast.Array); ok && len(val.Rules) == len(p.Rules) {
			for i := 0; i < len(val.Rules); i++ {
				testWalk(t, val.Rules[i], p.Rules[i])
			}
		}
	case *ast.Group:
		if p, ok := parsed.(*ast.Group); ok && len(val.Entries) == len(p.Entries) {
			t.Logf("Valid: %T, Parsed: %T\n", valid, parsed)
			t.Logf("Len valid: %d, Len parsed: %d", len(val.Entries), len(p.Entries))
			for i := 0; i < len(val.Entries); i++ {
				testWalk(t, val.Entries[i], p.Entries[i])
			}
		}
	case *ast.TypeChoice:
		if p, ok := parsed.(*ast.TypeChoice); ok {
			testWalk(t, val.First, p.First)
			testWalk(t, val.Second, p.Second)
		}

	case *ast.SizeOperatorControl:
		if p, ok := parsed.(*ast.SizeOperatorControl); ok {
			testWalk(t, val.Type, p.Type)
			testWalk(t, val.Size, p.Size)
		}
	case *ast.Range:
		if p, ok := parsed.(*ast.Range); ok {
			testWalk(t, val.From, p.From)
			testWalk(t, val.To, p.To)
		}
	case *ast.Entry:
		t.Logf("Valid: %T-%+v, Parsed: %T\n", valid, valid, parsed)
		if p, ok := parsed.(*ast.Entry); ok {
			testWalk(t, val.Name, p.Name)
			testWalk(t, val.Value, p.Value)
			if p.TrailingComment != nil {
				testWalk(t, val.TrailingComment, p.TrailingComment)
			}
		} else {
			t.Fatalf("expected node of type %T but found %T", valid, parsed)
			return
		}
	default:
		t.Fatalf("unknown node type %T", val)
		return
	}
}

func assertEqualDiagnostic(t *testing.T, expected, parsed errors.Diagnostic) {
	if expected.Diagnostic() != parsed.Diagnostic() {
		t.Logf("Errors expected: %s, Errors parsed: %s", expected.Diagnostic(), parsed.Diagnostic())
		t.Fail()
	}
}

// Test Parsing of rules in the form name = `type` where `type` is a predefined name.
// Covers https://www.rfc-editor.org/rfc/rfc8610#section-3.3
func TestParseIdentiferToTypeRule(t *testing.T) {
	name := &ast.Identifier{Name: "name"}
	typePos := token.Position{Line: 1, Column: 8, Offset: 7}
	tests := []struct {
		src   string
		value ast.Node
		err   errors.Diagnostic
	}{
		{"name = bool", &ast.BooleanType{Pos: typePos, Token: token.BOOL}, nil},
		{"name = uint", &ast.UintType{Range: token.PositionRange{Start: typePos, End: typePos.To(4)}, Token: token.UINT}, nil},
		{"name = nint", &ast.NegativeIntegerType{Pos: typePos, Token: token.NINT}, nil},
		{"name = int", &ast.IntegerType{Pos: typePos, Token: token.INT}, nil},
		{"name = float", &ast.FloatType{Pos: typePos, Token: token.FLOAT}, nil},
		{"name = float16", &ast.FloatType{Pos: typePos, Token: token.FLOAT16}, nil},
		{"name = float32", &ast.FloatType{Pos: typePos, Token: token.FLOAT32}, nil},
		{"name = float64", &ast.FloatType{Pos: typePos, Token: token.FLOAT64}, nil},

		{"name = bytes", &ast.BytesType{Pos: typePos, Token: token.BYTES}, nil},
		{"name = bstr", &ast.BstrType{Pos: typePos, Token: token.BSTR}, nil},

		{"name = tstr", &ast.TstrType{Pos: typePos, Token: token.TSTR}, nil},
		{"name = text", &ast.TstrType{Pos: typePos, Token: token.TEXT}, nil},

		{"name = null", &ast.NullType{Pos: typePos, Token: token.NULL}, nil},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}

		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, err := p.ParseFile()
		if err != nil {
			t.Fatal(tst.src, ": -> ", err)
		}
		testWalk(t, trueAst, parsed)
	}
}

// Test parsing text literals of the form "text here"
func TestParseTextLiteral(t *testing.T) {
	name := &ast.Identifier{Name: "name"}
	tests := []struct {
		src   string
		value *ast.TextLiteral
		err   parser.ErrorList
	}{
		{`name = "text"`, &ast.TextLiteral{Literal: "text"}, parser.ErrorList{}},
		{`name = "'red' pen"`, &ast.TextLiteral{Literal: "'red' pen"}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestNumericLiteral(t *testing.T) {
	name := &ast.Identifier{Name: "num"}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		{"num = 1", &ast.IntegerLiteral{Literal: 1, Token: token.INT}, parser.ErrorList{}},
		{"num = 2.4", &ast.FloatLiteral{Literal: 2.4, Token: token.FLOAT}, parser.ErrorList{}},
		{"num = 0x10", &ast.IntegerLiteral{Literal: 16, Token: token.INT}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestRegexpOperator(t *testing.T) {
	name := &ast.Identifier{Name: "some-text"}
	tstrPos := token.Position{Offset: 12, Line: 1, Column: 13}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		{`some-text = tstr .regexp ""`, &ast.Regexp{
			Base:  &ast.TstrType{Pos: tstrPos, Token: token.TSTR},
			Regex: &ast.TextLiteral{Literal: "", Token: token.TEXT_LITERAL},
		}, parser.ErrorList{}},
		{`some-text = tstr .regexp "[A-Za-z0-9]+@[A-Za-z0-9]+(\\.[A-Za-z0-9]+)+"`, &ast.Regexp{
			Base:  &ast.TstrType{Pos: tstrPos, Token: token.TSTR},
			Regex: &ast.TextLiteral{Literal: `[A-Za-z0-9]+@[A-Za-z0-9]+(\\.[A-Za-z0-9]+)+`, Token: token.TEXT_LITERAL},
		}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestCommentGroups(t *testing.T) {
	tests := []struct {
		src   string
		value *ast.CommentGroup
		err   parser.ErrorList
	}{
		{"; first line\n;second line", &ast.CommentGroup{
			List: []*ast.Comment{
				{Text: " first line"},
				{Text: "second line"},
			},
		}, parser.ErrorList{},
		},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{tst.value}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestTag(t *testing.T) {
	name := &ast.Identifier{Name: "tag"}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{

		// Test empty tag. The any keyword
		{"tag = #", &ast.Tag{}, parser.ErrorList{}},

		// Tags with a major tag only
		{"tag = #7", &ast.Tag{Major: &ast.UintLiteral{Literal: 7}}, parser.ErrorList{}},
		{"tag = #0", &ast.Tag{Major: &ast.UintLiteral{Literal: 0}}, parser.ErrorList{}},

		// Tags with major tag and tag number only
		{"tag = #2.63", &ast.Tag{Major: &ast.UintLiteral{Literal: 2}, TagNumber: &ast.UintLiteral{Literal: 63}}, parser.ErrorList{}},
		{"tag = #7.0", &ast.Tag{Major: &ast.UintLiteral{Literal: 7}, TagNumber: &ast.UintLiteral{Literal: 0}}, parser.ErrorList{}},

		// Tags with a Major, tag number and inner type
		{"tag = #6.2(bstr)", &ast.Tag{
			Major:     &ast.UintLiteral{Literal: 6},
			TagNumber: &ast.UintLiteral{Literal: 2},
			Item:      &ast.BstrType{Pos: token.Position{Offset: 11, Line: 1, Column: 12}, Token: token.BSTR},
		}, parser.ErrorList{},
		},

		// Tags with complex inner types
		// {"tag = #6.999([liquid, solid])", &ast.Tag{
		// 	Major:     &ast.UintLiteral{Literal: 6},
		// 	TagNumber: &ast.UintLiteral{Literal: 999},
		// 	Item:      &ast.Array{Rules: []ast.Node{&ast.Identifier{Name: "liquid"}, &ast.Identifier{Name: "solid"}}},
		// }, errors.New("parser error: identifier liquid referenced does not exist")},
	}
	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestTypeChoice(t *testing.T) {
	name := &ast.Identifier{Name: "choice"}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		// nested literal choice
		{`choice = "bow tie" / "necktie" / "Internet attire"`, &ast.TypeChoice{
			First: &ast.TextLiteral{Literal: "bow tie"},
			Second: &ast.TypeChoice{
				First:  &ast.TextLiteral{Literal: "necktie"},
				Second: &ast.TextLiteral{Literal: "Internet attire"},
			},
		}, parser.ErrorList{}},
		{`choice = 6 / 17`, &ast.TypeChoice{First: &ast.IntegerLiteral{Literal: 6}, Second: &ast.IntegerLiteral{Literal: 17}}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}
		testWalk(t, trueAst, parsed)
	}
}

// Test parsing of the .size control operator according to
// https://www.rfc-editor.org/rfc/rfc8610#section-3.8.1
func TestOperatorSize(t *testing.T) {
	name := &ast.Identifier{Name: "item"}
	basePos := token.Position{Offset: 7, Line: 1, Column: 8}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		{`item = bstr .size 4`, &ast.SizeOperatorControl{Type: &ast.BstrType{Pos: basePos, Token: token.BSTR}, Size: &ast.IntegerLiteral{Literal: 4}}, parser.ErrorList{}},
		{`item = bstr .size (1..63)`, &ast.SizeOperatorControl{
			Type: &ast.BstrType{Pos: basePos, Token: token.BSTR},
			Size: &ast.Group{Entries: []ast.GroupEntry{
				&ast.Range{From: &ast.IntegerLiteral{Literal: 1}, To: &ast.IntegerLiteral{Literal: 63}}},
			},
		}, parser.ErrorList{}},
		{`item = uint .size 3`, &ast.SizeOperatorControl{Type: &ast.UintType{Range: token.PositionRange{Start: basePos, End: basePos.To(4)}, Token: token.UINT}, Size: &ast.IntegerLiteral{Literal: 3}}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestParseGroup(t *testing.T) {
	name := &ast.Identifier{Name: "item"}
	basePos := token.Position{Offset: 7, Line: 1, Column: 8}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		{`item = (name: tstr)`, &ast.Group{Pos: basePos, Entries: []ast.GroupEntry{
			&ast.Entry{Name: &ast.Identifier{Name: "name"}, Value: &ast.TstrType{Pos: token.Position{Line: 1, Column: 15, Offset: 14}, Token: token.TSTR}},
		}}, parser.ErrorList{}},
	}

	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}

		testWalk(t, trueAst, parsed)
	}
}

func TestRange(t *testing.T) {
	name := &ast.Identifier{Name: "range"}
	// basePos := token.Position{Offset: 9, Line: 1, Column: 10}
	tests := []struct {
		src   string
		value ast.Node
		err   parser.ErrorList
	}{
		{`range = 0..10.0`, &ast.Range{}, parser.ErrorList{parser.NewError("cannot use float literal as upper bound to int range", token.Position{Line: 1, Column: 12}, token.Position{Line: 1, Column: 15})}},
	}
	for _, tst := range tests {
		trueAst := &ast.CDDL{Rules: []ast.CDDLEntry{&ast.Rule{Name: name, Value: tst.value}}}
		l := lexer.NewLexer([]byte(tst.src))
		p := parser.NewParser(l)

		parsed, errs := p.ParseFile()
		if len(errs) == len(tst.err) {
			for i := 0; i < len(errs); i++ {
				assertEqualDiagnostic(t, tst.err[i], errs[i])
			}
		}
		testWalk(t, trueAst, parsed)
	}

}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func readSource(filename string) (src []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	src, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	return
}

func TestE2EFast(t *testing.T) {
	root := rootDir()
	testData := filepath.Join(root, "testdata", "language")

	files, err := ioutil.ReadDir(testData)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".cddl" {
			fp := filepath.Join(testData, file.Name())
			t.Logf("Testing cddl file %s", fp)
			src, err := readSource(fp)
			if err != nil {
				t.Fatal(err)
			}

			l := lexer.NewLexer(src)
			p := parser.NewParser(l)

			// Parse and check that no errors exists. TODO(HannesKimara): Verify that the source is parsed correctly into the tree
			// or that the output cddl strictly describes an annotation data file.
			_, errs := p.ParseFile()
			if len(errs) != 0 {
				t.Log(errs)
				t.Fail()
			}
		}
	}
}
