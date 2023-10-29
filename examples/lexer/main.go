package main

import (
	"fmt"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/token"
)

func main() {
	src := `
		min-age = 18
		max-age = 150

		byte = uint .size 1
		public-key = [24*24 byte]
		person = (name: tstr, public-key: public-key)

		adult = (person, age: min-age .. max-age) ; adults are composed from person
	`

	lex := lexer.NewLexer([]byte(src))

	for {
		tok, pos, lit := lex.Scan()
		fmt.Printf("%s: %s -> %s\n", pos, tok, lit)
		if tok == token.EOF {
			break
		}
	}
}
