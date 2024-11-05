package gogen_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	gogen "github.com/HannesKimara/cddlc/transforms/codegen/golang"
	"github.com/iancoleman/strcase"
)

const (
	TEST_PATH = "./test-cases"
)

func TestGen(t *testing.T) {
	gen := gogen.NewGenerator("lib")
	gen.SetWithValidators(false)

	tt := []struct {
		name      string
		input     string
		expected  string
		generator *gogen.Generator
	}{
		{
			name:     "var",
			input:    `attire = "bow tie"`,
			expected: "package lib\n\n// (cddlc) Ident: attire\nvar Attire = bow tie\n",
		},
		{
			name:     "type definition",
			input:    `request-id = uint`,
			expected: "package lib\n\n// (cddlc) Ident: request-id\ntype RequestId uint\n",
		},
		{
			name: "structs",
		},
		{
			name: "enum",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input
			expected := tc.expected
			if len(tc.input) == 0 {
				input, expected = mustLoadTestCase(tc.name, t)
			}

			l := lexer.NewLexer([]byte(input))
			p := parser.NewParser(l)

			cddl, err := p.ParseFile()
			if err != nil {
				t.Fatal(err)
			}

			gen.Visit(cddl)

			buf := bytes.NewBuffer(nil)
			gen.String(buf)
			actual := buf.String()
			if actual != expected {
				fmt.Printf("expected:\n%s\ngot:\n%s\n", expected, actual)
				t.Fatalf("expected:\n%#v\ngot:\n%#v", expected, actual)
			}

			gen.Reset()
		})
	}
}

func mustLoadTestCase(name string, t *testing.T) (input, expected string) {
	input = mustLoadFile(name+".cddl", t)
	expected = mustLoadFile(name+".expected", t)
	return
}

func mustLoadFile(name string, t *testing.T) string {
	buf, err := os.ReadFile(filepath.Join(TEST_PATH, name))
	if err != nil {
		t.Fatal(err)
	}

	// Windows newlines
	buf = bytes.ReplaceAll(buf, []byte("\r\n"), []byte("\n"))

	return string(buf)
}

func TestIdentifierFormatter(t *testing.T) {
	gen := gogen.NewGenerator("lib")
	gen.SetWithValidators(false)

	// Usage example: integer keys have meaning in the trailing comments that
	// is useful as struct field names.
	gen.SetIdentifierFormatter(func(_ string, trailingComment string, suggestion string) string {
		name := suggestion

		if len(trailingComment) > 0 {
			name = strcase.ToCamel(trailingComment)
		}

		return name
	})
	input := `status-message = {
		0: text ; status
	}`
	expected := "package lib\n\n// (cddlc) Ident: status-message\ntype StatusMessage struct {\n\tStatus string `cbor:\"0\"`\n}\n"

	l := lexer.NewLexer([]byte(input))
	p := parser.NewParser(l)

	cddl, err := p.ParseFile()
	if err != nil {
		t.Fatal(err)
	}

	gen.Visit(cddl)

	buf := bytes.NewBuffer(nil)
	gen.String(buf)
	actual := buf.String()
	if actual != expected {
		fmt.Printf("expected:\n%s\ngot:\n%s\n", expected, actual)
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected, actual)
	}
}
