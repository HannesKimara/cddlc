package main

import (
	"fmt"
	"log"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
)

func main() {
	src := `
		min-age = 18
		max-age = 150

		byte = uint .size 1
		public-key = [24*24 byte]
		person = (name: tstr, public-key: public-key)

		adult = (~person, age: min-age .. max-age) ; adults are composed from person
	`

	lex := lexer.NewLexer([]byte(src))
	p := parser.NewParser(lex)
	cddl, err := p.ParseFile()

	if err != nil {
		log.Fatal()
	}

	fmt.Printf("Found %d rules\n", cddl.Rules)
}
