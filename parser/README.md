# Package parser

The parser package builds an abstract syntax tree from a set of source tokens. 

## Example

This example covers parsing a cddl source with literals, control operators, ranges and compositions. For more examples go to the [examples](../examples/) folder.

```go
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
```

## License

This project is licensed under the Apache-2.0 license. Please see the [LICENSE](../LICENSE) file for more details.
