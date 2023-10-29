package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/HannesKimara/cddlc/ast"
	"github.com/HannesKimara/cddlc/ast/astutils"
	"github.com/HannesKimara/cddlc/lexer"
	"github.com/HannesKimara/cddlc/parser"
	"github.com/HannesKimara/cddlc/token"
	"github.com/urfave/cli/v2"

	env "github.com/HannesKimara/cddlc/environment"
)

const PROMPT string = ">>>"

func Repl(cCtx *cli.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	parseVerbose := false

	fmt.Println("Welcome to the cddlc quick repl.")
	environ := env.NewEnvironment()

	for {
		fmt.Printf("%s ", PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return errors.New("could not scan")
		}

		text := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(text, ":help") {
			fmt.Println(":help - Prints this help message")
			fmt.Println(":tree - Prints the syntax tree for a value in scope")
			fmt.Println()
			fmt.Println(":pv   - Toggle printing the type after every execution")
			fmt.Println()
			fmt.Println(":exit - Exits the REPL")
			continue
		}

		if strings.HasPrefix(text, ":pv") {
			parseVerbose = !parseVerbose
			continue
		}

		if strings.HasPrefix(text, ":exit") {
			fmt.Println("exiting...")
			os.Exit(0)
		}

		if strings.HasPrefix(text, ":tree") {
			args := strings.Split(text, " ")[1:]
			for _, arg := range args {
				val := environ.Get(arg)
				if val != nil {
					fmt.Printf(arg + ": ")
					astutils.Print(val)
				}
			}

			continue
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

		if parseVerbose {
			printcddlShort(cddl)
		}

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
