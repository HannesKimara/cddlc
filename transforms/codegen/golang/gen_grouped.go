package gogen

import (
	"errors"
	"fmt"
	gast "go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/HannesKimara/cddlc/ast"
)

func (g *Generator) Visit(node ast.Node) *Generator {
	switch val := node.(type) {
	case *ast.Rule:
		var decl gast.Decl
		var outExpr gast.Expr
		var declToken token.Token
		var specs []gast.Spec

		stct, err := g.transpileNode(val.Value)
		if err != nil {
			panic(err)
		}
		if v, ok := stct.node.(gast.Expr); !ok {
			panic(fmt.Sprintf("unexpected transpilation product for source %s - %s (%T)", val.Start(), val.End(), stct.node))
		} else {
			outExpr = v
		}

		switch val.Value.(type) {

		case *ast.BooleanLiteral, *ast.FloatLiteral, *ast.IntegerLiteral, *ast.TextLiteral, *ast.UintLiteral:
			declToken = token.VAR
			specs = []gast.Spec{
				&gast.ValueSpec{
					Names:  []*gast.Ident{g.transpileIdentifier(val.Name)},
					Values: []gast.Expr{outExpr},
				},
			}

		default:
			declToken = token.TYPE
			specs = []gast.Spec{
				&gast.TypeSpec{
					Name: g.transpileIdentifier(val.Name),
					Type: outExpr,
				},
			}
			// panic(fmt.Sprintf("unexpected type %+v", val.Value))
		}

		decl = &gast.GenDecl{
			Tok:   declToken,
			Specs: specs,
			Doc:   &gast.CommentGroup{List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + val.Name.Name + "\n"}}},
		}

		// stct.spec = decl
		g.file.Decls = append(g.file.Decls, decl)

		log.Println("Len("+val.Name.Name+"): ", len(stct.validators))

		valDecl := g.bundleValidators(stct.validators, val)
		g.file.Decls = append(g.file.Decls, valDecl)
	case *ast.CDDL:
		for _, rule := range val.Rules {
			g.Visit(rule)
		}
	}

	return g
}

func (g *Generator) bundleValidators(validators []gast.CallExpr, rule *ast.Rule) *gast.FuncDecl {
	selfIdent := gast.NewIdent(strings.ToLower(rule.Name.Name))

	funcDecl := &gast.FuncDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{
				{
					Text: "\n// Valid evaluates type constraints on " + rule.Name.Name + " and returns nil if valid \n// else it returns a list of validation errors",
				},
			},
		},
		Recv: &gast.FieldList{
			List: []*gast.Field{{
				Names: []*gast.Ident{
					selfIdent,
				},
				Type: &gast.StarExpr{
					X: gast.NewIdent(g.transpileIdentifier(rule.Name).String()),
				},
			}},
		},
		Name: gast.NewIdent("Valid"),
		Type: &gast.FuncType{
			Results: &gast.FieldList{
				List: []*gast.Field{{
					Type: gast.NewIdent("error"),
				}},
			},
		},
	}

	block := &gast.BlockStmt{
		List: []gast.Stmt{},
	}

	if len(validators) > 0 {
		for _, validator := range validators {
			ex := &gast.ExprStmt{
				X: &validator,
			}
			block.List = append(block.List, ex)
		}
	}

	endStmt := &gast.ReturnStmt{
		Results: []gast.Expr{
			gast.NewIdent("nil"),
		},
	}

	block.List = append(block.List, endStmt)
	funcDecl.Body = block

	return funcDecl
}

func (g *Generator) transpileGroupLike(entries []ast.GroupEntry) (*structure, error) {
	fl := &gast.FieldList{}
	for _, entry := range entries {
		var field *gast.Field
		switch val := entry.(type) {
		case *ast.Entry:
			stct, err := g.transpileEntry(val)
			if err != nil {
				panic(err)
			}
			field = stct.node.(*gast.Field)

		case *ast.Optional:
			stct, err := g.transpileNode(val.Item)
			if err != nil {
				return nil, err
			}
			if cast, ok := stct.node.(*gast.Field); ok {
				cast.Tag = &gast.BasicLit{Kind: token.STRING, Value: "`cbor:\",omitempty\"`"}
				field = cast
			} else {
				return nil, err
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
	stctRet := newStructure(fl)
	return stctRet, nil
}

func (g *Generator) transpileGroup(group *ast.Group) (*structure, error) {
	fl, err := g.transpileGroupLike(group.Entries)

	if err != nil {
		return fl, err
	}

	st := &gast.StructType{Fields: fl.node.(*gast.FieldList)}
	ret := newStructure(st)
	ret.Embed(fl)

	return ret, nil
}

func (g *Generator) transpileArray(arr *ast.Array) (*structure, error) {
	fl := &gast.FieldList{}
	fl.List = append(fl.List, &gast.Field{
		Names: []*gast.Ident{gast.NewIdent("_")},
		Type: &gast.StructType{
			Fields: &gast.FieldList{},
		},
		Tag: &gast.BasicLit{Kind: token.STRING, Value: "`cbor:\",toarray\"`"},
	})

	fl2, err := g.transpileGroupLike(arr.Rules)
	if err != nil {
		return nil, err
	}

	fl.List = append(fl.List, fl2.node.(*gast.FieldList).List...)
	return newStructure(&gast.StructType{Fields: fl}), nil
}

func (g *Generator) transpileEntry(entry *ast.Entry) (*structure, error) {
	ident := g.transpileIdentifier(entry.Name)
	stct, err := g.transpileNode(entry.Value)

	if err != nil {
		return nil, err
	}

	field := &gast.Field{
		Names: []*gast.Ident{ident},
		Type:  stct.node.(gast.Expr),
	}

	stctRet := newStructure(field)

	log.Println("Embed called")
	stctRet.Embed(stct)

	return stctRet, nil
}

func (g *Generator) transpileNode(node ast.Node) (*structure, error) {
	if node == nil {
		return nil, errors.New("transpiler: node should not be nil")
	}
	switch val := node.(type) {
	case *ast.Group:
		return g.transpileGroup(val)
	case *ast.Array:
		return g.transpileArray(val)
	case *ast.Entry:
		return g.transpileEntry(val)
	case *ast.BooleanType:
		return newStructure(g.transpileBoolType(val)), nil
	case *ast.Identifier:
		return newStructure(g.transpileIdentifier(val)), nil
	case *ast.TstrType:
		return newStructure(g.transpileTstrType(val)), nil
	case *ast.BytesType:
		return newStructure(g.transpileBytesType(val)), nil
	case *ast.Comment:
		return newStructure(g.transpileComment(val)), nil
	case *ast.NullType:
		return newStructure(g.transpileNullType(val)), nil
	case *ast.IntegerType:
		return newStructure(g.transpileIntegerType(val)), nil
	case *ast.UintType:
		return newStructure(g.transpileUintType(val)), nil
	case *ast.IntegerLiteral:
		return newStructure(g.transpileIntegerLiteral(val)), nil
	case *ast.BooleanLiteral:
		return newStructure(g.transpileBoolLiteral(val)), nil
	case *ast.TextLiteral:
		return newStructure(g.transpileTextLiteral(val)), nil
	case *ast.FloatLiteral:
		return newStructure(g.transpileFloatLiteral(val)), nil
	case *ast.NMOccurrence:
		return newStructure(g.transpileNMOccurence(val)), nil
	case *ast.SizeOperatorControl:
		return g.transformSizeOp(val)
	default:
		panic(fmt.Sprintf("unexpected type %T", val))
	}
}
