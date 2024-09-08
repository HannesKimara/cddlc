package gogen

import (
	"errors"
	"fmt"
	gast "go/ast"
	gtoken "go/token"
	"strconv"
	"strings"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/iancoleman/strcase"
)

type outputType uint

const (
	outputTypeLiteral = iota
	outputTypeTypeDef
	outputTypeEnum
	outputTypeStruct
)

func (g *Generator) isChoiceGroup(group *ast.Group) bool {
	if len(group.Entries) > 0 {
		if entry, ok := group.Entries[0].(*ast.Entry); ok {
			if _, ok := entry.Value.(*ast.IntegerLiteral); ok {
				// This is a group used as a choice.
				return true
			}
		}
	}
	return false
}

func (g *Generator) handleRuleComment(cddlTypeName string, goTypeName string) {
	if g.lastRootComment == nil {
		return
	}
	if g.commentHandler != nil {
		g.commentHandler.HandleRuleComment(strings.TrimSpace(g.lastRootComment.Text), cddlTypeName, goTypeName)
	}

	g.lastRootComment = nil
}

func (g *Generator) Visit(node ast.Node) *Generator {
	switch val := node.(type) {
	case *ast.CommentGroup:
		if g.lastRootComment != nil && g.commentHandler != nil {
			g.commentHandler.HandleComment(strings.TrimSpace(g.lastRootComment.Text))
		}

		comments := []string{}
		for _, comment := range val.List {
			comments = append(comments, strings.TrimSpace(comment.Text))
		}
		g.lastRootComment = &ast.Comment{Text: strings.Join(comments, "\n")}

	case *ast.Comment:
		if g.lastRootComment != nil && g.commentHandler != nil {
			g.commentHandler.HandleComment(strings.TrimSpace(g.lastRootComment.Text))
		}
		g.lastRootComment = val

	case *ast.Rule:
		var outType outputType
		switch val.Value.(type) {
		case *ast.BooleanLiteral, *ast.FloatLiteral, *ast.IntegerLiteral, *ast.TextLiteral, *ast.UintLiteral:
			outType = outputTypeLiteral
		case *ast.UintType, *ast.IntegerType, *ast.FloatType:
			outType = outputTypeTypeDef
		case *ast.Enumeration:
			outType = outputTypeEnum
		case *ast.Map, *ast.Array:
			outType = outputTypeStruct
		case *ast.Group:
			outType = outputTypeStruct
			if group, ok := val.Value.(*ast.Group); ok {
				if g.isChoiceGroup(group) {
					// This is a group used as a choice.
					outType = outputTypeEnum
				}
			}
		}

		switch outType {
		case outputTypeLiteral:
			stct, err := g.transpileNode(val.Value)
			if err != nil {
				panic(err)
			}
			name := g.transpileIdentifier(val.Name, nil)
			g.handleRuleComment(val.Name.Name, name.Name)

			expr, ok := stct.node.(gast.Expr)
			if !ok {
				panic(fmt.Sprintf("unexpected transpilation product for source %s - %s (%T)", val.Start(), val.End(), stct.node))
			}

			decl := &gast.GenDecl{
				Tok: gtoken.VAR,
				Specs: []gast.Spec{
					&gast.ValueSpec{
						Names:  []*gast.Ident{name},
						Values: []gast.Expr{expr},
					},
				},
				Doc: &gast.CommentGroup{
					List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + val.Name.Name + "\n"}},
				},
			}
			g.appendDecl(decl)

		case outputTypeTypeDef:
			name := g.transpileIdentifier(val.Name, nil)
			g.handleRuleComment(val.Name.Name, name.Name)
			typ, err := g.transpileNode(val.Value)
			if err != nil {
				panic(fmt.Sprintf("unexpected type for source %s - %s (%T)", val.Start(), val.End(), val.Value))
			}
			decl := &gast.GenDecl{
				Tok: gtoken.TYPE,
				Specs: []gast.Spec{
					&gast.TypeSpec{
						Name: name,
						Type: typ.node.(*gast.Ident),
					},
				},
				Doc: &gast.CommentGroup{
					List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + val.Name.Name + "\n"}},
				},
			}
			g.appendDecl(decl)

		case outputTypeEnum: // Enumeration (& choice group)
			g.currentGroupName = strcase.ToCamel(val.Name.Name)
			g.currentGroupType = ast.GroupTypeEnum
			g.outputEnum(val.Name, val.Value)
			g.handleRuleComment("", val.Name.Name)

		case outputTypeStruct: // Map, Group & Array
			g.currentGroupName = strcase.ToCamel(val.Name.Name)
			g.currentGroupType = ast.GroupTypeStruct

			stct, err := g.transpileNode(val.Value)
			if err != nil {
				panic(err)
			}
			name := g.transpileIdentifier(val.Name, nil)
			g.handleRuleComment(val.Name.Name, name.Name)

			typeExpr, ok := stct.node.(gast.Expr)
			if !ok {
				panic(fmt.Sprintf("unexpected transpilation product for source %s - %s (%T)", val.Start(), val.End(), stct.node))
			}

			decl := &gast.GenDecl{
				Tok: gtoken.TYPE,
				Specs: []gast.Spec{
					&gast.TypeSpec{
						Name: name,
						Type: typeExpr,
					},
				},
				Doc: &gast.CommentGroup{
					List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + val.Name.Name + "\n"}},
				},
			}
			g.appendDecl(decl)

			if g.withValidators {
				valDecl := g.bundleValidators(stct.validators, val)
				g.file.Decls = append(g.file.Decls, valDecl)
			}

		default:
			panic(fmt.Sprintf("unexpected type %+v", val.Value))
		}

	case *ast.CDDL:
		for _, rule := range val.Rules {
			g.Visit(rule)
		}

	default:
		fmt.Printf("Skipping node %T", node)
	}

	return g
}

func (g *Generator) outputEnum(eName *ast.Identifier, eValue ast.Node) gast.Expr {
	// Enum type
	name := g.transpileIdentifier(eName, nil)
	decl := &gast.GenDecl{
		Tok: gtoken.TYPE,
		Specs: []gast.Spec{
			&gast.TypeSpec{
				Name: name,
				Type: g.transpileUintType(nil),
			},
		},
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{Text: "\n// (cddlc) Ident: " + name.Name + "\n"}},
		},
	}
	g.appendDecl(decl)

	stct, err := g.transpileEnumerationValue(eValue)
	if err != nil {
		panic(err)
	}

	valuesDecl, ok := stct.node.(*gast.GenDecl)
	if !ok {
		panic(fmt.Sprintf("unexpected transpilation product for source %s - %s (%T)", eValue.Start(), eValue.End(), stct.node))
	}
	g.appendDecl(valuesDecl)

	return name
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
					X: gast.NewIdent(g.transpileIdentifier(rule.Name, nil).String()),
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
				typ := cast.Type
				if _, ok := cast.Type.(*gast.ArrayType); !ok {
					cast.Type = &gast.StarExpr{
						X: typ,
					}
				}
				tag := cast.Tag.Value
				if len(tag) > 3 {
					cast.Tag.Value = tag[:len(tag)-2] + ",omitempty" + tag[len(tag)-2:]
				}
				field = cast
			} else {
				return nil, err
			}
		case *ast.Identifier:
			field = &gast.Field{
				Type: g.transpileIdentifier(val, nil),
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
	var ret *structure
	switch g.currentGroupType {
	case ast.GroupTypeStruct:
		fl, err := g.transpileGroupLike(group.Entries)
		if err != nil {
			return nil, err
		}

		st := &gast.StructType{Fields: fl.node.(*gast.FieldList)}
		ret = newStructure(st)
		ret.Embed(fl)

	case ast.GroupTypeEnum:
		specs := []gast.Spec{}
		for _, entry := range group.Entries {
			var spec *gast.ValueSpec
			switch val := entry.(type) {
			case *ast.Entry:
				spec = &gast.ValueSpec{
					Names: []*gast.Ident{
						gast.NewIdent(g.currentGroupName + strcase.ToCamel(val.Name.Name)),
					},
					Type: g.formatIdentifier(g.currentGroupName, nil),
					Values: []gast.Expr{
						&gast.BasicLit{
							Kind:  gtoken.INT,
							Value: strconv.FormatInt(val.Value.(*ast.IntegerLiteral).Literal, 10),
						},
					},
				}

			default:
				panic(fmt.Sprintf("What was that? %T: `%+v`", val, val))
			}

			specs = append(specs, spec)
		}

		ret = newStructure(&gast.GenDecl{
			Tok:   gtoken.CONST,
			Specs: specs,
		})
	}

	return ret, nil
}

func (g *Generator) transpileArray(arr *ast.Array) (*structure, error) {
	if len(arr.Rules) == 1 {
		if mn, ok := arr.Rules[0].(*ast.NMOccurrence); ok {
			return newStructure(g.transpileNMOccurence(mn)), nil
		}
	}

	// fxamacker/cbor way to indicate encoding without field names:
	// _ struct{} `cbor:",toarray"`
	// TODO: make customizable since other libraries have different convention.
	// E.g.: ugorji/go/codec style:
	// _struct bool    `codec:",toarray"`
	fieldList := &gast.FieldList{
		List: []*gast.Field{{
			Names: []*gast.Ident{gast.NewIdent("_")},
			Type: &gast.StructType{
				Fields: &gast.FieldList{},
			},
			Tag: &gast.BasicLit{Kind: gtoken.STRING, Value: "`cbor:\",toarray\"`"},
		}},
	}

	groupFieldList, err := g.transpileGroupLike(arr.Rules)
	if err != nil {
		return nil, err
	}

	fieldList.List = append(fieldList.List, groupFieldList.node.(*gast.FieldList).List...)
	return newStructure(&gast.StructType{Fields: fieldList}), nil
}

func (g *Generator) transpileMap(m *ast.Map) (*structure, error) {
	fl := &gast.FieldList{}
	for _, rule := range m.Rules {
		var field *gast.Field
		switch val := rule.(type) {
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
				typ := cast.Type
				if _, ok := cast.Type.(*gast.ArrayType); !ok {
					cast.Type = &gast.StarExpr{
						X: typ,
					}
				}
				tag := cast.Tag.Value
				if len(tag) > 3 {
					cast.Tag.Value = tag[:len(tag)-2] + ",omitempty" + tag[len(tag)-2:]
				}
				field = cast
			} else {
				return nil, err
			}
		case *ast.Identifier:
			field = &gast.Field{
				Type: g.transpileIdentifier(val, nil),
			}
		default:
			panic(fmt.Sprintf("What was that? %T: `%+v`", val, val))
		}
		fl.List = append(fl.List, field)
	}

	return newStructure(&gast.StructType{Fields: fl}), nil
}

func (g *Generator) transpileEnumeration(e *ast.Enumeration) (*structure, error) {
	return g.transpileEnumerationValue(e.Value)
}

func (g *Generator) transpileEnumerationValue(eValue ast.Node) (*structure, error) {
	return g.transpileNode(eValue)
}

func (g *Generator) transpileInlineEnumeration(name string, e *ast.Enumeration) (*structure, error) {
	outerGroupType := g.currentGroupType
	outerGroupName := g.currentGroupName
	defer func() {
		g.currentGroupType = outerGroupType
		g.currentGroupName = outerGroupName
	}()
	g.currentGroupType = ast.GroupTypeEnum
	g.currentGroupName = name

	eName := &ast.Identifier{
		Name: name,
	}
	return newStructure(g.outputEnum(eName, e.Value)), nil
}

func (g *Generator) tagPartWithLabel(label, value string) string {
	return fmt.Sprintf(`%s:"%s"`, label, value)
}
func (g *Generator) combineTags(parts []string) string {
	tag := strings.Join(parts, ",")
	return fmt.Sprintf("`%s`", tag)
}

func (g *Generator) isInlineEnumeration(e *ast.Enumeration) bool {
	if _, ok := e.Value.(*ast.Group); ok {
		return true
	}
	return false
}

func (g *Generator) transpileEntry(entry *ast.Entry) (*structure, error) {
	fieldName := g.transpileIdentifier(entry.Name, entry.TrailingComment)

	var typeExpr gast.Expr
	if e, ok := entry.Value.(*ast.Enumeration); ok && g.isInlineEnumeration(e) {
		inlineName := strcase.ToCamel(g.currentGroupName + "-" + fieldName.Name)
		stct, err := g.transpileInlineEnumeration(inlineName, e)
		if err != nil {
			return nil, err
		}
		typeExpr = stct.node.(gast.Expr)
	} else {
		stct, err := g.transpileNode(entry.Value)
		if err != nil {
			return nil, err
		}
		typeExpr = stct.node.(gast.Expr)
	}

	tag := g.combineTags(
		[]string{
			g.tagPartWithLabel(g.fieldTagName, entry.Name.Name),
		},
	)
	field := &gast.Field{
		Names: []*gast.Ident{fieldName},
		Type:  typeExpr,
		Tag:   &gast.BasicLit{Kind: gtoken.STRING, Value: tag},
	}

	stctRet := newStructure(field)
	// stctRet.Embed(stct)

	return stctRet, nil
}

func (g *Generator) transpileTypeChoice(c *ast.TypeChoice) (*structure, error) {
	g.currentGroupType = outputTypeEnum
	return g.transpileNode(c.First)
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
	case *ast.Map:
		return g.transpileMap(val)
	case *ast.Enumeration:
		return g.transpileEnumeration(val)
	case *ast.Entry:
		return g.transpileEntry(val)
	case *ast.TypeChoice:
		return g.transpileTypeChoice(val)
	case *ast.BooleanType:
		return newStructure(g.transpileBoolType(val)), nil
	case *ast.Identifier:
		return newStructure(g.transpileIdentifier(val, nil)), nil
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
	case *ast.FloatType:
		return newStructure(g.transpileFloatType(val)), nil
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
		panic(fmt.Sprintf("unexpected type %T: %s", val, val))
	}
}
