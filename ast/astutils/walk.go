package astutils

import (
	"fmt"

	"github.com/flowfunction/cddl/ast"
)

type Visitor interface {
	Visit(ast.Node) Visitor
}

func Walk(v Visitor, node ast.Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {

	case *ast.Array:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}

	case *ast.BooleanType:
		// pass

	case *ast.BytesType:
		// pass

	case *ast.BstrType:
		// pass

	case *ast.CDDL:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}

	case *ast.Comment:
		// pass

	case *ast.ComparatorOpControl:
		if n.Left != nil {
			Walk(v, n.Left)
		}
		if n.Right != nil {
			Walk(v, n.Right)
		}

	case *ast.Regexp:
		if n.Base != nil {
			Walk(v, n.Base)
		}
		if n.Regex != nil {
			Walk(v, n.Regex)
		}

	case *ast.SizeOperatorControl:
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.Size != nil {
			Walk(v, n.Size)
		}

	case *ast.Entry:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}
		if n.TrailingComment != nil {
			Walk(v, n.TrailingComment)
		}

	case *ast.FloatType:
		// pass

	case *ast.FloatLiteral:
		// pass

	case *ast.Group:
		for _, rule := range n.Entries {
			Walk(v, rule)
		}

	case *ast.GroupChoice:
		if n.First != nil {
			Walk(v, n.First)
		}
		if n.Second != nil {
			Walk(v, n.Second)
		}

	case *ast.Identifier:
		// pass

	case *ast.IntegerType:
		// pass

	case *ast.NegativeIntegerType:
		// pass

	case *ast.IntegerLiteral:
		// pass

	case *ast.Map:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}

	case *ast.NullType:
		// pass

	case *ast.Optional:
		if n.Item != nil {
			Walk(v, n.Item)
		}

	case *ast.NMOccurrence:
		if n.N != nil {
			Walk(v, n.N)
		}
		if n.M != nil {
			Walk(v, n.M)
		}
		if n.Item != nil {
			Walk(v, n.Item)
		}

	case *ast.Range:
		if n.From != nil {
			Walk(v, n.From)
		}
		if n.To != nil {
			Walk(v, n.To)
		}

	case *ast.Rule:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}
		if n.TrailingComment != nil {
			Walk(v, n.TrailingComment)
		}

	case *ast.Tag:
		if n.Major != nil {
			Walk(v, n.Major)
		}
		if n.TagNumber != nil {
			Walk(v, n.TagNumber)
		}
		if n.Item != nil {
			Walk(v, n.Item)
		}

	case *ast.TextLiteral:
		// pass

	case *ast.TstrType:
		// pass

	case *ast.TypeChoice:
		if n.First != nil {
			Walk(v, n.First)
		}
		if n.Second != nil {
			Walk(v, n.Second)
		}

	case *ast.UintLiteral:
		// pass

	case *ast.Unwrap:
		if n.Item != nil {
			Walk(v, n.Item)
		}
	default:
		panic(fmt.Sprintf("walk: unexpected node of type %T", n))
	}

	v.Visit(nil)
}
