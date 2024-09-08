package gogen

import (
	"errors"
	gast "go/ast"
	"log"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/HannesKimara/cddlc/token"
)

func (g *Generator) transformSizeOp(op *ast.SizeOperatorControl) (*structure, error) {
	var size *ast.IntegerLiteral

	switch val := op.Size.(type) {
	case *ast.IntegerLiteral:
		size = val
	case *ast.Identifier:
		// TODO: resolve and cast
		panic("ident literal size not implemented")
	}

	baseStct, err := g.transpileNode(op.Type)
	if err != nil {
		return nil, err
	}
	stct := &structure{
		node: baseStct.node,
	}

	switch op.Type.(type) {
	case *ast.UintType:
		stct.addValidatorCall("validators", "CheckSize", "self", size.Literal) // test
	case *ast.BstrType, *ast.BytesType:
		stct.addValidatorCall("validators", "CheckSize", "self", size.Literal)
	case *ast.TstrType:
		stct.addValidatorCall("validators", "CheckSize", "self", size.Literal)
	default:
		panic("fix later")
	}
	if g.withValidators {
		g.addImport(VALIDATOR_PKG, "")
	}

	log.Println("Len validators: ", len(stct.validators))

	return stct, nil
}

func (g *Generator) transpileNMOccurence(nm *ast.NMOccurrence) *gast.ArrayType {
	stct, err := g.transpileNode(nm.Item)
	if err != nil {
		panic("fix later")
	}
	return &gast.ArrayType{
		Elt: stct.node.(gast.Expr),
	}
}

func (g *Generator) transpileComparatorOp(op ast.ComparatorOpControl) (*structure, error) {
	_, err := g.transpileNode(op.Left)
	if err != nil {
		return nil, err
	}

	_, err = g.transpileNode(op.Right)
	if err != nil {
		return nil, err
	}

	switch op.Operator {
	case token.LT.String(): // .lt
	case token.LE.String(): // .le
	case token.GT.String(): // .gt
	case token.GE.String(): // .gt
	case token.EQ.String(): // .eq
	case token.EQ.String(): // .ne

	default:
		return nil, errors.New("comparator operator `" + op.Operator + "` not implemented")
	}

	return nil, nil
}
