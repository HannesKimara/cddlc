package ast

type Visitor interface {
	Visit(Node) Visitor
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Array:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}
	case *BooleanType:
		// pass
	case *Range:
		Walk(v, n.From)
		Walk(v, n.To)
	case *BytesType:
		// pass
	case *BstrType:
		// pass
	case *CDDL:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}
	case *Comment:
		// pass
	case *ComparatorOpControl:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *Regexp:
		Walk(v, n.Base)
		Walk(v, n.Regex)
	case *SizeOperatorControl:
		Walk(v, n.Size)
		Walk(v, n.Type)
	case *FloatType:
		// pass
	case *FloatLiteral:
		// pass
	case *Group:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}
	case *GroupChoice:
		Walk(v, n.First)
		Walk(v, n.Second)
	case *Identifier:
		// pass
	case *IntegerType:
		// pass
	case *NegativeIntegerType:
		// pass
	case *IntegerLiteral:
		// pass
	case *Map:
		for _, rule := range n.Rules {
			Walk(v, rule)
		}
	case *NullType:
		// pass
	case *Optional:
		Walk(v, n.Item)
	case *NMOccurrence:
		Walk(v, n.N)
		Walk(v, n.M)
		Walk(v, n.Item)
	case *Rule:
		Walk(v, n.Name)
		Walk(v, n.Value)
	case *Tag:
		Walk(v, n.Major)
		Walk(v, n.TagNumber)
		Walk(v, n.Item)
	case *TextLiteral:
		// pass
	case *TstrType:
		// pass
	case *TypeChoice:
		Walk(v, n.First)
		Walk(v, n.Second)
	case *UintLiteral:
		// pass
	case *Unwrap:
		Walk(v, n.Item)
	}
}
