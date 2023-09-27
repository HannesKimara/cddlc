package gogen

import (
	"fmt"
	gast "go/ast"
	"go/format"
	"go/token"
	"io"
)

const (
	VALIDATOR_PKG = "github.com/flowfunction/cddl/runtime/validators"
)

type structure struct {
	node gast.Node
	// spec gast.Decl

	validators []gast.CallExpr
}

func (s *structure) Embed(em *structure) {
	s.validators = append(s.validators, em.validators...)
}

func newStructure(node gast.Node) *structure {
	return &structure{
		node: node,
	}
}

func (s *structure) addValidatorCall(X, sel string, args ...interface{}) {
	call := &gast.SelectorExpr{
		X:   gast.NewIdent(X),
		Sel: gast.NewIdent(sel),
	}

	exprArgs := []gast.Expr{}

	for _, arg := range args {
		expr := gast.NewIdent(fmt.Sprintf("%v", arg))
		exprArgs = append(exprArgs, expr)
	}

	callexpr := gast.CallExpr{
		Fun:  call,
		Args: exprArgs,
	}

	s.validators = append(s.validators, callexpr)
}

func (s *structure) bindReferences(ref string) {
}

// Generator contains the internal representation of the generation step
type Generator struct {
	pkg string

	// structures []*structure
	imports []*gast.ImportSpec

	file *gast.File
	fset *token.FileSet
}

// String flushes the generated tree to an output
func (g *Generator) String(output io.Writer) (err error) {
	g.file.Imports = append(g.file.Imports, g.imports...)
	gast.SortImports(g.fset, g.file)

	if len(g.file.Imports) > 0 {
		importDecl := &gast.GenDecl{Specs: []gast.Spec{}, Tok: token.IMPORT}

		for _, spec := range g.file.Imports {
			importDecl.Specs = append(importDecl.Specs, spec)
		}

		g.file.Decls = append([]gast.Decl{importDecl}, g.file.Decls...)
	}

	return format.Node(output, g.fset, g.file)
}

// Package returns the name of the generated golang package
func (g *Generator) Package() string {
	return g.pkg
}

func (g *Generator) addImport(value string, name string) {
	var ident *gast.Ident
	if name != "" { // nullable ident to prevent space on generated output
		ident = gast.NewIdent(name)
	}

	im := &gast.ImportSpec{
		Name: ident,
		Path: &gast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + value + "\"",
		},
	}

	g.imports = append(g.imports, im)
}

func (g *Generator) emit(stct *structure) {}

// NewGenerator returns a new Generator given the golang package name
func NewGenerator(pkgName string) *Generator {
	fset := token.NewFileSet()
	// tkfile := fset.AddFile("schema.go", fset.Base()+1, 100000)

	file := &gast.File{
		Name: gast.NewIdent(pkgName),
	}

	gen := &Generator{
		pkg:  pkgName,
		file: file,
		fset: fset,
	}

	return gen
}
