package environment

import (
	"errors"

	"github.com/HannesKimara/cddlc/ast"
)

var (
	ErrSymbolExists = errors.New("symbol already exists")
)

type Environment struct {
	symbols map[string]ast.Node
}

// Add a new symbol to the symbol table with a pointer to its Node
func (e *Environment) Add(ident string, value ast.Node) error {
	if e.Exists(ident) {
		return ErrSymbolExists
	}
	e.symbols[ident] = value
	return nil
}

// Exists checks whether the symbol exists in the symbol table
func (e *Environment) Exists(ident string) bool {
	_, ok := e.symbols[ident]
	return ok
}

func (e *Environment) Get(ident string) ast.Node {
	if e.Exists(ident) {
		return e.symbols[ident]
	}
	return nil
}

// NewEnvironment returns a new Environment
func NewEnvironment() *Environment {
	return &Environment{
		symbols: make(map[string]ast.Node),
	}
}
