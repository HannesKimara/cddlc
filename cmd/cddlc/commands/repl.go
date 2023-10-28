package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	"github.com/HannesKimara/cddlc/token"
	"github.com/urfave/cli/v2"

	env "github.com/HannesKimara/cddlc/environment"
)

const PROMPT string = ">>>"

func Repl(cCtx *cli.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the cddlc quick repl.")
	environ := env.NewEnvironment()

	for {
		fmt.Printf("%s ", PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return errors.New("could not scan")
		}

		if cCtx.Bool("lex") {
			printTokens(scanner.Bytes())
		}

		l := lexer.NewLexer(scanner.Bytes())
		p := parser.NewParser(l,
			parser.WithEnv(environ),
		)

		cddl, errs := p.ParseFile()
		if len(errs) != 0 {
			fmt.Println(errs.String())
		}

		printcddlShort(cddl)

	}
}

func printTokens(src []byte) {
	lex := lexer.NewLexer(src)
	for {
		tok, pos, lit := lex.Scan()
		fmt.Printf("%s: %s -> %s\n", pos, tok, lit)

		if tok == token.EOF {
			break
		}
	}

}

func printcddlShort(cddl *ast.CDDL) {
	for _, rule := range cddl.Rules {
		if r, ok := rule.(*ast.Rule); ok {
			fmt.Printf("%s ; %T\n", r.Name.Name, r.Value)
		}

	}
}
