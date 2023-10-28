package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	"github.com/urfave/cli/v2"
)

const PROMPT string = ">>>"

func Repl(cCtx *cli.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the cddlc quick repl.\n ! Symbol table not supported yet")
	for {
		fmt.Printf("%s ", PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return errors.New("could not scan")
		}

		l := lexer.NewLexer(scanner.Bytes())
		p := parser.NewParser(l)

		cddl, errs := p.ParseFile()
		if len(errs) != 0 {
			fmt.Println(errs.String())
		}

		fmt.Println(cddl)

	}
}
