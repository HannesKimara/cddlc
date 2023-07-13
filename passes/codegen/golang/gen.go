package gogen

import (
	"errors"
	"fmt"
	gast "go/ast"
	"go/printer"
	"go/token"
	"io"
	"strconv"
	"strings"

	"github.com/flowfunction/cddl/ast"
	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/ast/astutil"
)

type Generator struct {
	file *gast.File
	fset *token.FileSet
}

func (g *Generator) String(output io.Writer) error {
	return printer.Fprint(output, g.fset, g.file)
}

func (g *Generator) Visit(node ast.Node) *Generator {
	switch val := node.(type) {
	case *ast.Rule:
		var decl gast.Decl
		var outExpr gast.Expr
		var declToken token.Token
		var specs []gast.Spec
		// var ident = val.Name.Name

		out, err := g.transpileNode(val.Value)
		if err != nil {
			panic(err)
		}
		if v, ok := out.(gast.Expr); !ok {
			panic(fmt.Sprintf("unexpected transpilation product for source %s - %s", val.Start(), val.End()))
		} else {
			outExpr = v
		}

		switch val.Value.(type) {
		case *ast.BstrType, *ast.IntegerType, *ast.BytesType, *ast.FloatType,
			*ast.NullType, *ast.UintType, *ast.TstrType, *ast.BooleanType, *ast.NegativeIntegerType, *ast.Group, *ast.Map, *ast.Array:
			declToken = token.TYPE
			specs = []gast.Spec{
				&gast.TypeSpec{
					Name: g.transpileIdentifier(val.Name),
					Type: outExpr,
				},
			}

		case *ast.BooleanLiteral, *ast.FloatLiteral, *ast.IntegerLiteral, *ast.TextLiteral, *ast.UintLiteral:
			declToken = token.VAR
			specs = []gast.Spec{
				&gast.ValueSpec{
					Names:  []*gast.Ident{g.transpileIdentifier(val.Name)},
					Values: []gast.Expr{outExpr},
				},
			}

		default:
			panic(fmt.Sprintf("unexpected type %+v", val.Value))
		}

		decl = &gast.GenDecl{
			Tok:   declToken,
			Specs: specs,
			Doc:   &gast.CommentGroup{List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + val.Name.Name + "\n"}}},
		}

		g.file.Decls = append(g.file.Decls, decl)
	case *ast.CDDL:
		for _, rule := range val.Rules {
			g.Visit(rule)
		}
	}
	return g
}

func (g *Generator) transpileGroupLike(entries []ast.GroupEntry) *gast.FieldList {
	fl := &gast.FieldList{}
	for _, entry := range entries {
		var field *gast.Field
		switch val := entry.(type) {
		case *ast.Entry:
			field = g.transpileEntry(val)

		case *ast.Optional:
			node, err := g.transpileNode(val.Item)
			if err != nil {
				panic("fix later")
			}
			if cast, ok := node.(*gast.Field); ok {
				cast.Tag = &gast.BasicLit{Kind: token.STRING, Value: "`cbor:\",omitempty\"`"}
				field = cast
			} else {
				panic("fix later")
			}
		case *ast.Identifier:
			field = &gast.Field{
				Type: g.transpileIdentifier(val),
			}
		case *ast.NMOccurrence:
			field = &gast.Field{
				Type: g.transpileNMOccurence(val),
			}
		default:
			panic(fmt.Sprintf("What was that? %T: `%+v`", val, val))
		}

		fl.List = append(fl.List, field)
	}
	return fl
}

func (g *Generator) transpileGroup(group *ast.Group) *gast.StructType {
	fl := g.transpileGroupLike(group.Entries)
	return &gast.StructType{Fields: fl}
}

func (g *Generator) transpileArray(arr *ast.Array) *gast.StructType {
	fl := &gast.FieldList{}
	fl.List = append(fl.List, &gast.Field{
		Names: []*gast.Ident{gast.NewIdent("_")},
		Type: &gast.StructType{
			Fields: &gast.FieldList{},
		},
		Tag: &gast.BasicLit{Kind: token.STRING, Value: "`cbor:\",toarray\"`"},
	})

	fl2 := g.transpileGroupLike(arr.Rules)

	fl.List = append(fl.List, fl2.List...)
	return &gast.StructType{
		Fields: fl,
	}
}

func (g *Generator) transpileBoolLiteral(bl *ast.BooleanLiteral) *gast.Ident {
	return &gast.Ident{
		Name: fmt.Sprintf("%t", bl.Bool),
	}
}

func (g *Generator) transpileBoolType(bt *ast.BooleanType) *gast.Ident {
	return &gast.Ident{
		Name: "bool",
	}
}

func (g *Generator) transpileIdentifier(ident *ast.Identifier) *gast.Ident {
	formatted := strcase.ToCamel(ident.Name)

	if ident.IsSocket() || ident.IsPlug() {
		formatted = strings.TrimLeft(ident.Name, "$")
	}

	if token.IsIdentifier(formatted) {
		return &gast.Ident{
			Name: formatted,
		}
	}
	_, err := strconv.ParseInt(formatted, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("Identifier %s -> %s could not be transformed to valid go identifier", ident.Name, formatted))
	}
	return &gast.Ident{
		Name: "IntKey_" + formatted,
	}
}

func (g *Generator) transpileNMOccurence(nm *ast.NMOccurrence) *gast.ArrayType {
	element, err := g.transpileNode(nm.Item)
	if err != nil {
		panic("fix later")
	}
	return &gast.ArrayType{
		Elt: element.(gast.Expr),
	}
}

func (g *Generator) transpileTextLiteral(lit *ast.TextLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.STRING,
		Value: lit.Literal,
	}
}

func (g *Generator) transpileFloatLiteral(lit *ast.FloatLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.FLOAT,
		Value: strconv.FormatFloat(lit.Literal, 'f', -1, 64),
	}
}

func (g *Generator) transpileTstrType(tstr *ast.TstrType) *gast.Ident {
	return &gast.Ident{
		Name: "string",
	}
}

func (g *Generator) transpileBytesType(bt *ast.BytesType) *gast.ArrayType {
	return &gast.ArrayType{
		Elt: &gast.Ident{
			Name: "byte",
		},
	}
}

func (g *Generator) transpileComment(com *ast.Comment) *gast.Comment {
	return &gast.Comment{
		Text: com.Text,
	}
}

func (g *Generator) transpileNullType(nt *ast.NullType) *gast.Ident {
	return &gast.Ident{
		Name: "nil",
	}
}

func (g *Generator) transpileIntegerType(it *ast.IntegerType) *gast.Ident {
	return &gast.Ident{
		Name: "int",
	}
}

func (g *Generator) transpileUintType(it *ast.UintType) *gast.Ident {
	return &gast.Ident{
		Name: "uint",
	}
}

func (g *Generator) transpileIntegerLiteral(il *ast.IntegerLiteral) *gast.BasicLit {
	return &gast.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", il.Literal),
	}
}

func (g *Generator) transpileEntry(entry *ast.Entry) *gast.Field {
	ident := g.transpileIdentifier(entry.Name)
	value, err := g.transpileNode(entry.Value)
	if err != nil {
		panic("fix later")
	}
	field := &gast.Field{
		Names: []*gast.Ident{ident},
		Type:  value.(gast.Expr),
	}
	return field
}

func (g *Generator) transpileNode(node ast.Node) (gast.Node, error) {
	if node == nil {
		return nil, errors.New("transpiler: node should not be nil")
	}
	switch val := node.(type) {
	case *ast.Group:
		return g.transpileGroup(val), nil
	case *ast.Array:
		return g.transpileArray(val), nil
	case *ast.Entry:
		return g.transpileEntry(val), nil
	case *ast.BooleanType:
		return g.transpileBoolType(val), nil
	case *ast.Identifier:
		return g.transpileIdentifier(val), nil
	case *ast.TstrType:
		return g.transpileTstrType(val), nil
	case *ast.BytesType:
		return g.transpileBytesType(val), nil
	case *ast.Comment:
		return g.transpileComment(val), nil
	case *ast.NullType:
		return g.transpileNullType(val), nil
	case *ast.IntegerType:
		return g.transpileIntegerType(val), nil
	case *ast.UintType:
		return g.transpileUintType(val), nil
	case *ast.IntegerLiteral:
		return g.transpileIntegerLiteral(val), nil
	case *ast.BooleanLiteral:
		return g.transpileBoolLiteral(val), nil
	case *ast.TextLiteral:
		return g.transpileTextLiteral(val), nil
	case *ast.FloatLiteral:
		return g.transpileFloatLiteral(val), nil
	case *ast.NMOccurrence:
		return g.transpileNMOccurence(val), nil
	default:
		panic(fmt.Sprintf("unexpected type %T", val))
	}
}

func NewGenerator() *Generator {
	fset := token.NewFileSet()
	// tkfile := fset.AddFile("schema.go", fset.Base()+1, 100000)

	pkgName := "lib"
	file := &gast.File{
		Name: gast.NewIdent(pkgName),
	}

	// header := `
	// //
	// // Code generated by github.com/flowfunction/cddl - DO NOT EDIT THIS FILE DIRECTLY
	// //

	// package
	// `
	// header += pkgName
	// file, err := parser.ParseFile(fset, pkgName, header, parser.ParseComments)
	// if err != nil {
	// 	panic("cddlc/gen: Failed to create file header from internal template")
	// }

	astutil.AddImport(fset, file, "github.com/flowfunction/cddl/runtime/validators")
	astutil.AddImport(fset, file, "github.com/fxamacker/cbor/v2")

	gen := &Generator{
		file: file,
		fset: fset,
	}

	return gen
}
