package ast

type CDDLEntry interface {
	Node
	cddlEntry()
}
